package keeltest

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"

	"go.uber.org/zap"

	"github.com/foomo/keel/log"
	"github.com/foomo/keel/net/http/middleware"
)

// ServiceHTTP struct
type ServiceHTTP struct {
	server *httptest.Server
	name   string
	l      *zap.Logger
}

func NewServiceHTTP(l *zap.Logger, name string, handler http.Handler, middlewares ...middleware.Middleware) *ServiceHTTP {
	if l == nil {
		l = log.Logger()
	}
	// enrich the log
	l = log.WithHTTPServerName(l, name)

	server := httptest.NewUnstartedServer(middleware.Compose(l, name, handler, middlewares...))
	server.Config.ErrorLog = zap.NewStdLog(l)

	return &ServiceHTTP{
		server: server,
		name:   name,
		l:      l,
	}
}

func (s *ServiceHTTP) Name() string {
	return s.name
}

func (s *ServiceHTTP) Logger() *zap.Logger {
	return s.l
}

func (s *ServiceHTTP) URL() string {
	return s.server.URL
}

func (s *ServiceHTTP) Start(ctx context.Context) error {
	var fields []zap.Field

	if value := strings.Split(s.server.Listener.Addr().String(), ":"); len(value) == 2 {
		ip, port := value[0], value[1]
		if ip == "" {
			ip = "0.0.0.0"
		}

		fields = append(fields, log.FNetHostIP(ip), log.FNetHostPort(port))
	}

	s.l.Info("starting http test service", fields...)
	s.server.Config.BaseContext = func(_ net.Listener) context.Context { return ctx }
	s.server.Start()

	return nil
}

func (s *ServiceHTTP) Close(_ context.Context) error {
	s.l.Info("stopping http test service")
	s.server.Close()

	return nil
}
