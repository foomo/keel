package service

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
	"github.com/foomo/keel/net/http/middleware"
)

// HTTP struct
type HTTP struct {
	running atomic.Bool
	server  *http.Server
	name    string
	l       *zap.Logger
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewHTTP(l *zap.Logger, name, addr string, handler http.Handler, middlewares ...middleware.Middleware) *HTTP {
	if l == nil {
		l = log.Logger()
	}
	// enrich the log
	l = log.WithAttributes(l,
		log.KeelServiceTypeKey.String("http"),
		log.KeelServiceNameKey.String(name),
	)

	return &HTTP{
		server: &http.Server{
			Addr:         addr,
			ErrorLog:     zap.NewStdLog(l),
			IdleTimeout:  5 * time.Second,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			Handler:      middleware.Compose(l, name, handler, middlewares...),
		},
		name: name,
		l:    l,
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Getter
// ------------------------------------------------------------------------------------------------

func (s *HTTP) Name() string {
	return s.name
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (s *HTTP) Healthz() error {
	if !s.running.Load() {
		return ErrServiceNotRunning
	}
	return nil
}

func (s *HTTP) String() string {
	return fmt.Sprintf("`%T` on `%s`", s.server.Handler, s.server.Addr)
}

func (s *HTTP) Start(ctx context.Context) error {
	var fields []zap.Field
	if value := strings.Split(s.server.Addr, ":"); len(value) == 2 {
		ip, port := value[0], value[1]
		if ip == "" {
			ip = "0.0.0.0"
		}
		fields = append(fields, log.FNetHostIP(ip), log.FNetHostPort(port))
	}
	s.l.Info("starting keel service", fields...)
	s.server.BaseContext = func(_ net.Listener) context.Context { return ctx }
	s.server.RegisterOnShutdown(func() {
		s.running.Store(false)
	})
	s.running.Store(true)
	if err := s.server.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
		return nil
	} else if err != nil {
		return errors.Wrap(err, "failed to start service")
	}
	return nil
}

func (s *HTTP) Close(ctx context.Context) error {
	s.l.Info("stopping keel service")
	if err := s.server.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "failed to stop service")
	}
	return nil
}
