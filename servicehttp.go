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
	name   string
	l      *zap.Logger
}

func NewServiceHTTP(l *zap.Logger, name, addr string, handler http.Handler, middlewares ...middleware.Middleware) *ServiceHTTP {
	if l == nil {
		l = log.Logger()
	}
	// enrich the log
	l = log.WithHTTPServerName(l, name)

	return &ServiceHTTP{
		server: &http.Server{
			Addr:     addr,
			ErrorLog: zap.NewStdLog(l),
			Handler:  middleware.Compose(l, name, handler, middlewares...),
		},
		name: name,
		l:    l,
	}
}

func (s *ServiceHTTP) Name() string {
	return s.name
}
//
//func (s *ServiceHTTP) Logger() *zap.Logger {
//	return s.l
//}
//
//func (s *ServiceHTTP) Server() *http.Server {
//	return s.server
//}

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
