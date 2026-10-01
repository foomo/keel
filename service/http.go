package service

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync/atomic"

	keelhttp "github.com/foomo/keel/net/http"
	keelsemconv "github.com/foomo/keel/semconv"
	"github.com/pkg/errors"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

// HTTP is a service that serves an [net/http.Server]. Create it with [NewHTTP].
type HTTP struct {
	l       *zap.Logger
	name    string
	server  *http.Server
	running atomic.Bool
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

// NewHTTP returns an [HTTP] service named name listening on addr and serving
// handler wrapped in middlewares. See [keelhttp.NewServer]. A nil l defaults to
// [log.Logger].
func NewHTTP(l *zap.Logger, name, addr string, handler http.Handler, middlewares ...keelhttp.Middleware) *HTTP {
	if l == nil {
		l = log.Logger()
	}
	// enrich the log
	l = log.WithAttributes(l,
		keelsemconv.KeelServiceType("http"),
		keelsemconv.KeelServiceName(name),
	)

	return &HTTP{
		l:      l,
		name:   name,
		server: keelhttp.NewServer(l, name, addr, handler, middlewares...),
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Getter
// ------------------------------------------------------------------------------------------------

// Name returns the service name.
func (s *HTTP) Name() string {
	return s.name
}

// Server returns the underlying HTTP server.
func (s *HTTP) Server() *http.Server {
	return s.server
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

// Healthz returns [ErrServiceNotRunning] unless the server is listening.
func (s *HTTP) Healthz() error {
	if !s.running.Load() {
		return ErrServiceNotRunning
	}

	return nil
}

// String returns a short description of the handler type and address used in
// the readme.
func (s *HTTP) String() string {
	return fmt.Sprintf("`%T` on `%s`", s.server.Handler, s.server.Addr)
}

// Start listens on the configured address and serves requests with ctx as the
// base request context, blocking until the server is shut down. It returns an
// error if the address cannot be bound or serving fails; a graceful shutdown
// returns nil.
func (s *HTTP) Start(ctx context.Context) error {
	var fields []zap.Field

	if value := strings.Split(s.server.Addr, ":"); len(value) == 2 {
		ip, port := value[0], value[1]
		if ip == "" {
			ip = "0.0.0.0"
		}

		fields = append(fields, log.Attributes(semconv.ServerAddress(ip), semconv.ServerPortKey.String(port))...)
	}

	s.l.Info("starting keel service", fields...)
	s.server.BaseContext = func(_ net.Listener) context.Context { return ctx }
	s.server.RegisterOnShutdown(func() {
		s.running.Store(false)
	})

	// Bind before reporting healthy: an occupied port must fail the service
	// rather than leave it marked running while nothing is listening.
	var lc net.ListenConfig

	ln, err := lc.Listen(ctx, "tcp", s.server.Addr)
	if err != nil {
		return errors.Wrapf(err, "failed to listen to addr: %s", s.server.Addr)
	}

	s.running.Store(true)

	if err := s.server.Serve(ln); errors.Is(err, http.ErrServerClosed) {
		return nil
	} else if err != nil {
		return errors.Wrap(err, "failed to start service")
	}

	return nil
}

// Close gracefully shuts down the server, bounded by ctx.
func (s *HTTP) Close(ctx context.Context) error {
	s.l.Info("stopping keel service")

	if err := s.server.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "failed to stop service")
	}

	return nil
}
