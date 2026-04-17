package nats

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/foomo/keel"
	"github.com/foomo/keel/log"
	"github.com/foomo/opentelemetry-go/semconv/natsconv"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.uber.org/zap"
)

// Connect establishes a connection to a NATS server with OTel instrumentation
// and structured logging on every lifecycle event.
func Connect(s *keel.Server, rawURL string, opts ...nats.Option) (*nats.Conn, error) {
	l := s.Logger().Named("nats")
	m := s.Meter()

	// Build instruments once; each returns a noop on nil meter, so errors here
	// only surface real meter-provider problems and should not block Connect.
	disconnects, err := natsconv.NewClientDisconnects(m)
	if err != nil {
		return nil, err
	}
	reconnects, err := natsconv.NewClientReconnects(m)
	if err != nil {
		return nil, err
	}
	asyncErrors, err := natsconv.NewClientAsyncErrors(m)
	if err != nil {
		return nil, err
	}

	// Background context for callbacks — NATS invokes these outside any request
	// context, and metric recording shouldn't be cancelled by request teardown.
	ctx := context.Background()

	opts = append(opts,
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
		nats.ReconnectJitter(100*time.Millisecond, 1*time.Second),
		nats.PingInterval(20*time.Second),

		nats.ConnectHandler(func(conn *nats.Conn) {
			l.Debug("connected", log.Attributes(serverAttrs(conn)...)...)
		}),

		nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
			addr, port := serverAddrPort(conn)
			disconnects.Add(ctx, 1, addr, disconnects.AttrServerPort(port))
			l.Warn("disconnected", zap.Error(err),
				zap.String("addr", conn.ConnectedUrlRedacted()),
				zap.String("server_id", conn.ConnectedServerId()),
			)
		}),

		nats.ReconnectErrHandler(func(conn *nats.Conn, err error) {
			addr, port := serverAddrPort(conn)
			reconnects.Add(ctx, 1, addr, reconnects.AttrServerPort(port))
			l.Info("reconnect", zap.Error(err),
				zap.String("addr", conn.ConnectedUrlRedacted()),
				zap.String("server_id", conn.ConnectedServerId()),
			)
		}),

		nats.ErrorHandler(func(conn *nats.Conn, sub *nats.Subscription, err error) {
			kind := classifyAsyncError(err)
			extraAttrs := []attribute.KeyValue{}
			if sub != nil && sub.Subject != "" {
				extraAttrs = append(extraAttrs, asyncErrors.AttrSubject(sub.Subject))
			}
			if addr, _ := serverAddrPort(conn); addr != "" {
				extraAttrs = append(extraAttrs, asyncErrors.AttrServerAddress(addr))
			}
			asyncErrors.Add(ctx, 1, kind, extraAttrs...)

			fields := []zap.Field{
				zap.Error(err),
				zap.String("kind", string(kind)),
				zap.String("addr", conn.ConnectedUrlRedacted()),
			}
			if sub != nil {
				fields = append(fields, zap.String("subject", sub.Subject))
			}
			l.Warn("async error", fields...)
		}),

		nats.ClosedHandler(func(conn *nats.Conn) {
			l.Debug("closed",
				zap.String("addr", conn.ConnectedUrlRedacted()),
				zap.String("server_id", conn.ConnectedServerId()),
			)
		}),

		nats.LameDuckModeHandler(func(conn *nats.Conn) {
			l.Info("server lame-duck mode",
				zap.String("addr", conn.ConnectedUrlRedacted()),
			)
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

// serverAttrs returns OTel semconv attributes describing the connected server.
// Returns nil if the connected URL cannot be parsed.
func serverAttrs(nc *nats.Conn) []attribute.KeyValue {
	u, err := url.Parse(nc.ConnectedUrlRedacted())
	if err != nil {
		return nil
	}
	port, _ := strconv.Atoi(u.Port())
	return []attribute.KeyValue{
		semconv.ServerAddress(u.Hostname()),
		semconv.ServerPort(port),
		semconv.NetworkProtocolName("nats"),
	}
}

// serverAddrPort extracts just the address and port from the connected URL.
// Returns zero values if parsing fails.
func serverAddrPort(nc *nats.Conn) (string, int) {
	u, err := url.Parse(nc.ConnectedUrlRedacted())
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
