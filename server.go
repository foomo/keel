package keel

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/foomo/keel/config"
	"github.com/foomo/keel/log"
	"github.com/foomo/keel/telemetry"
)

// Server struct
type Server struct {
	services        []Service
	shutdownTimeout time.Duration
	closers         []interface{}
	ctx             context.Context
	l               *zap.Logger
	c               *viper.Viper
}

func NewServer(opts ...Option) *Server {
	var (
		defaultShutdownTimeout = 5 * time.Second
		defaultConfig          = config.Config()
		defaultLogger          = log.Logger()
		defaultCtx             = context.Background()
	)

	inst := &Server{
		shutdownTimeout: defaultShutdownTimeout,
		ctx:             defaultCtx,
		c:               defaultConfig,
		l:               defaultLogger,
	}

	for _, opt := range opts {
		opt(inst)
	}

	return inst
}

// Logger returns server logger
func (s *Server) Logger() *zap.Logger {
	return s.l
}

// Config returns server config
func (s *Server) Config() *viper.Viper {
	return s.c
}

// Context returns server context
func (s *Server) Context() context.Context {
	return s.ctx
}

// AddServices adds multiple service
func (s *Server) AddServices(services ...Service) {
	for _, service := range services {
		s.AddService(service)
	}
}

// AddService add a single service
func (s *Server) AddService(service Service) {
	for _, value := range s.services {
		if value == service {
			return
		}
	}
	s.services = append(s.services, service)
}

// AddClosers adds an closer to be called on shutdown
func (s *Server) AddClosers(closers ...interface{}) {
	for _, closer := range closers {
		s.AddCloser(closer)
	}
}

// AddCloser adds a closer to be called on shutdown
func (s *Server) AddCloser(closer interface{}) {
	switch closer.(type) {
	case Closer,
		ErrorCloser,
		CloserWithContext,
		ErrorCloserWithContext,
		Shutdowner,
		ErrorShutdowner,
		ShutdownerWithContext,
		ErrorShutdownerWithContext,
		Unsubscriber,
		ErrorUnsubscriber,
		UnsubscriberWithContext,
		ErrorUnsubscriberWithContext:
		s.closers = append(s.closers, closer)
	default:
		s.l.Warn("unable to add closer")
	}
}

// Run runs the server
func (s *Server) Run() {
	s.l.Info("starting server")

	ctx, stop := signal.NotifyContext(s.ctx, os.Interrupt)
	defer stop()

	g, gctx := errgroup.WithContext(ctx)

	for _, service := range s.services {
		service := service
		g.Go(func() error {
			if err := service.Start(s.ctx); errors.Is(err, http.ErrServerClosed) {
				log.WithError(s.l, err).Debug("server has closed")
			} else if err != nil {
				log.WithError(s.l, err).Error("failed to start service")
				return err
			}
			return nil
		})
		// register started service
		s.AddCloser(service)
	}

	// gracefully shutdown servers
	g.Go(func() error {
		<-gctx.Done()
		s.l.Debug("gracefully stopping closers...")

		timeoutCtx, timeoutCancel := context.WithTimeout(
			context.Background(),
			s.shutdownTimeout,
		)
		defer timeoutCancel()

		// append internal closers
		closers := append(s.closers, telemetry.Exporter(), telemetry.Controller()) //nolint:gocritic

		for _, closer := range closers {
			switch c := closer.(type) {
			case Closer:
				c.Close()
			case ErrorCloser:
				if err := c.Close(); err != nil {
					log.WithError(s.l, err).Error("failed to gracefully stop ErrorCloser")
					continue
				}
			case CloserWithContext:
				c.Close(timeoutCtx)
			case ErrorCloserWithContext:
				if err := c.Close(timeoutCtx); err != nil {
					log.WithError(s.l, err).Error("failed to gracefully stop ErrorCloserWithContext")
					continue
				}
			case Shutdowner:
				c.Shutdown()
			case ErrorShutdowner:
				if err := c.Shutdown(); err != nil {
					log.WithError(s.l, err).Error("failed to gracefully stop ErrorShutdowner")
					continue
				}
			case ShutdownerWithContext:
				c.Shutdown(timeoutCtx)
			case ErrorShutdownerWithContext:
				if err := c.Shutdown(timeoutCtx); err != nil {
					log.WithError(s.l, err).Error("failed to gracefully stop ErrorShutdownerWithContext")
					continue
				}
			case Unsubscriber:
				c.Unsubscribe()
			case ErrorUnsubscriber:
				if err := c.Unsubscribe(); err != nil {
					log.WithError(s.l, err).Error("failed to gracefully stop ErrorUnsubscriber")
					continue
				}
			case UnsubscriberWithContext:
				c.Unsubscribe(timeoutCtx)
			case ErrorUnsubscriberWithContext:
				if err := c.Unsubscribe(timeoutCtx); err != nil {
					log.WithError(s.l, err).Error("failed to gracefully stop ErrorUnsubscriberWithContext")
					continue
				}
			}
			s.l.Info("stopped registered closer", log.FName(fmt.Sprintf("%T", closer)))
		}
		return gctx.Err()
	})

	// wait for shutdown
	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.WithError(s.l, err).Error("service error")
	}

	s.l.Info("graceful shutdown complete")
}
