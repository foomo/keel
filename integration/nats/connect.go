package nats

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/foomo/keel"
	"github.com/foomo/keel/log"
	"github.com/foomo/opentelemetry-go/semconv/natsconv"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

// Connect establishes a connection to the NATS server at rawURL with OTel
// instrumentation and structured logging on every lifecycle event.
//
// The connection reconnects indefinitely every 2s (with jitter) and pings
// every 20s. These options and the lifecycle handlers are appended after opts
// and therefore take precedence over them. Disconnects, reconnects and async
// errors are recorded with the meter of s, and the connection is registered
// as a closer on s.
func Connect(s keel.Runtime, rawURL string, opts ...nats.Option) (*nats.Conn, error) {
	l := s.Logger().Named("nats")
	m := s.Meter()

	// Build instruments once; each falls back to a noop on error, so errors are
	// only reported and never block Connect.
	disconnects, err := natsconv.NewClientDisconnects(m)
	if err != nil {
		otel.Handle(err)
	}

	reconnects, err := natsconv.NewClientReconnects(m)
	if err != nil {
		otel.Handle(err)
	}

	asyncErrors, err := natsconv.NewClientAsyncErrors(m)
	if err != nil {
		otel.Handle(err)
	}

	// Background context for callbacks — NATS invokes these outside any request
	// context, and metric recording shouldn't be cancelled by request teardown.
	ctx := context.Background()

	// ConnectedUrlRedacted is empty unless connected, so remember the last
	// connected server for the disconnect, reconnect error and closed handlers.
	var lastURL atomic.Pointer[string]

	serverURL := func(conn *nats.Conn) string {
		if v := conn.ConnectedUrlRedacted(); v != "" {
			lastURL.Store(&v)
			return v
		}

		if v := lastURL.Load(); v != nil {
			return *v
		}

		return ""
	}

	opts = append(opts,
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
		nats.ReconnectJitter(100*time.Millisecond, 1*time.Second),
		nats.PingInterval(20*time.Second),

		nats.ConnectHandler(func(conn *nats.Conn) {
			l.Debug("connected", log.Attributes(serverAttrs(serverURL(conn))...)...)
		}),

		nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
			u := serverURL(conn)
			addr, port := serverAddrPort(u)
			disconnects.Add(ctx, 1, addr, disconnects.AttrServerPort(port))
			log.WithError(l, err).Warn("disconnected", log.Attributes(serverAttrs(u)...)...)
		}),

		nats.ReconnectHandler(func(conn *nats.Conn) {
			u := serverURL(conn)
			addr, port := serverAddrPort(u)
			reconnects.Add(ctx, 1, addr, reconnects.AttrServerPort(port))
			l.Info("reconnected", log.Attributes(serverAttrs(u)...)...)
		}),

		nats.ReconnectErrHandler(func(conn *nats.Conn, err error) {
			log.WithError(l, err).Warn("reconnect failed", log.Attributes(serverAttrs(serverURL(conn))...)...)
		}),

		nats.ErrorHandler(func(conn *nats.Conn, sub *nats.Subscription, err error) {
			kind := classifyAsyncError(err)

			var extraAttrs []attribute.KeyValue
			if sub != nil && sub.Subject != "" {
				extraAttrs = append(extraAttrs, asyncErrors.AttrSubject(sub.Subject))
			}

			u := serverURL(conn)
			if addr, _ := serverAddrPort(u); addr != "" {
				extraAttrs = append(extraAttrs, asyncErrors.AttrServerAddress(addr))
			}

			asyncErrors.Add(ctx, 1, kind, extraAttrs...)

			// kind is the error.type; WithError would add the Go type name instead
			attrs := append(serverAttrs(u),
				semconv.ErrorTypeKey.String(string(kind)),
				semconv.ExceptionMessage(err.Error()),
			)
			if sub != nil {
				attrs = append(attrs, semconv.MessagingDestinationName(sub.Subject))
			}

			l.Warn("async error", log.Attributes(attrs...)...)
		}),

		nats.ClosedHandler(func(conn *nats.Conn) {
			l.Debug("closed", log.Attributes(serverAttrs(serverURL(conn))...)...)
		}),

		nats.LameDuckModeHandler(func(conn *nats.Conn) {
			l.Info("server lame-duck mode", log.Attributes(serverAttrs(serverURL(conn))...)...)
		}),

		nats.NoCallbacksAfterClientClose(),
	)

	conn, err := nats.Connect(rawURL, opts...)
	if err != nil {
		return nil, err
	}

	s.AddCloser(conn)

	return conn, nil
}

// serverAttrs returns OTel semconv attributes describing the server at
// rawURL. Returns nil if rawURL is empty or cannot be parsed.
func serverAttrs(rawURL string) []attribute.KeyValue {
	addr, port := serverAddrPort(rawURL)
	if addr == "" {
		return nil
	}

	return []attribute.KeyValue{
		semconv.ServerAddress(addr),
		semconv.ServerPort(port),
		semconv.NetworkProtocolName("nats"),
	}
}

// serverAddrPort extracts just the address and port from rawURL.
// Returns zero values if parsing fails.
func serverAddrPort(rawURL string) (string, int) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", 0
	}

	port, _ := strconv.Atoi(u.Port())

	return u.Hostname(), port
}

// classifyAsyncError maps nats async errors to low-cardinality kinds suitable
// for the nats.client.error.kind attribute. Unknown errors map to _OTHER.
func classifyAsyncError(err error) natsconv.AsyncErrorKindAttr {
	switch {
	case err == nil:
		return natsconv.AsyncErrorKindOther
	case errors.Is(err, nats.ErrSlowConsumer):
		return natsconv.AsyncErrorKindSlowConsumer
	case errors.Is(err, nats.ErrPermissionViolation):
		return natsconv.AsyncErrorKindPermissionViolation
	case errors.Is(err, nats.ErrAuthExpired):
		return natsconv.AsyncErrorKindAuthExpired
	case errors.Is(err, nats.ErrAuthRevoked):
		return natsconv.AsyncErrorKindAuthRevoked
	default:
		return natsconv.AsyncErrorKindOther
	}
}
