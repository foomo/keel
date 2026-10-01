package keeltest

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"

	keelhttp "github.com/foomo/keel/net/http"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

// ServiceHTTP is a [Service] serving an [http.Handler] on an
// [httptest.Server] bound to a random local port.
type ServiceHTTP struct {
	server *httptest.Server
	name   string
	l      *zap.Logger
}

// NewServiceHTTP returns a [ServiceHTTP] named name serving handler wrapped
// in middlewares. A nil l defaults to the global keel logger.
func NewServiceHTTP(l *zap.Logger, name string, handler http.Handler, middlewares ...keelhttp.Middleware) *ServiceHTTP {
	if l == nil {
		l = log.Logger()
	}
	// enrich the log
	l = log.WithHTTPServerName(l, name)

	server := httptest.NewUnstartedServer(keelhttp.Compose(l, name, handler, middlewares...))
	server.Config.ErrorLog = zap.NewStdLog(l)

	return &ServiceHTTP{
		server: server,
		name:   name,
		l:      l,
	}
}

// Name returns the service name.
func (s *ServiceHTTP) Name() string {
	return s.name
}

// Logger returns the service logger.
func (s *ServiceHTTP) Logger() *zap.Logger {
	return s.l
}

// URL returns the base URL of the server. It is empty until
// [ServiceHTTP.Start] is called.
func (s *ServiceHTTP) URL() string {
	return s.server.URL
}

// Start starts the server with ctx as the base context of all requests. It
// always returns nil.
func (s *ServiceHTTP) Start(ctx context.Context) error {
	var fields []zap.Field

	if value := strings.Split(s.server.Listener.Addr().String(), ":"); len(value) == 2 {
		ip, port := value[0], value[1]
		if ip == "" {
			ip = "0.0.0.0"
		}

		fields = append(fields, log.Attributes(semconv.ServerAddress(ip), semconv.ServerPortKey.String(port))...)
	}

	s.l.Info("starting http test service", fields...)
	s.server.Config.BaseContext = func(_ net.Listener) context.Context { return ctx }
	s.server.Start()

	return nil
}

// Close shuts down the server and blocks until all outstanding requests
// completed. It always returns nil.
func (s *ServiceHTTP) Close(_ context.Context) error {
	s.l.Info("stopping http test service")
	s.server.Close()

	return nil
}
