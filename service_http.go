package keel

import (
	"context"
	"net"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/foomo/keel/log"
	"github.com/foomo/keel/net/http/middleware"
)

// ServiceHTTP struct
type ServiceHTTP struct {
	server *http.Server
	l      *zap.Logger
}

func NewServiceHTTP(l *zap.Logger, addr string, handler http.Handler, middlewares ...middleware.Middleware) *ServiceHTTP {
	if l == nil {
		l = log.Logger()
	}
	errorLog, err := zap.NewStdLogAt(l, zap.ErrorLevel)
	log.Must(l, err, "failed to create std logger")

	return &ServiceHTTP{
		server: &http.Server{
			Addr:     addr,
			ErrorLog: errorLog,
			Handler:  middleware.Compose(l, handler, middlewares...),
		},
		l: l,
	}
}

func (s *ServiceHTTP) SetName(name string) *ServiceHTTP {
	s.l = s.l.With(log.FServiceName(name))
	return s
}

func (s *ServiceHTTP) Start(ctx context.Context) error {
	var fields []zap.Field
	if value := strings.Split(s.server.Addr, ":"); len(value) == 2 {
		ip, port := value[0], value[1]
		if ip == "" {
			ip = "0.0.0.0"
		}
		fields = append(fields, log.FNetHostIP(ip), log.FNetHostPort(port))
	}
	s.l.Info("starting http service", fields...)
	s.server.BaseContext = func(_ net.Listener) context.Context { return ctx }
	if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
		log.WithError(s.l, err).Error("service error")
		return err
	}
	return nil
}

func (s *ServiceHTTP) Close(ctx context.Context) error {
	s.l.Info("shutting down http service")
	return s.server.Shutdown(ctx)
}
