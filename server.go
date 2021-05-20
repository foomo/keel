package keel

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/foomo/keel/config"
	"github.com/foomo/keel/log"
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
		defaultContext         = context.Background() // TODO context.WithCancel
		defaultConfig          = config.Config()
		defaultLogger          = log.Logger()
	)

	inst := &Server{
		shutdownTimeout: defaultShutdownTimeout,
		ctx:             defaultContext,
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

// AddCloser adds an closer to be called on shutdown
func (s *Server) AddCloser(closer interface{}) {
	if closer != nil {
		switch closer.(type) {
		case Closer,
			CloserWithContext,
			Syncer,
			SyncerWithContext,
			Shutdowner,
			ShutdownerWithContext:
			s.closers = append(s.closers, closer)
		default:
			panic("unsupported closer")
		}
	}
}

// Run runs the server
func (s *Server) Run() {
	s.l.Info("starting server")

	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(
			signalChan,
			syscall.SIGHUP,  // kill -SIGHUP XXXX
			syscall.SIGINT,  // kill -SIGINT XXXX or Ctrl+c
			syscall.SIGTERM, // kill -SIGTERM XXXX
			syscall.SIGQUIT, // kill -SIGQUIT XXXX
		)

		// handle termination
		select {
		case sig := <-signalChan:
			s.l.Debug(fmt.Sprintf("go signal: %s", sig))
			cancel()
			return nil
		case <-gctx.Done():
			return gctx.Err()
		}
	})

	for _, service := range s.services {
		service := service
		g.Go(func() error {
			// TODO handle other 'positive' errors
			if err := service.Start(gctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

		for _, closer := range s.closers {
			switch c := closer.(type) {
			case Closer:
				if err := c.Close(); err != nil {
					log.WithError(s.l, err).Error("failed to gracefully stop Closer")
				}
			case CloserWithContext:
				if err := c.Close(timeoutCtx); err != nil {
					log.WithError(s.l, err).Error("failed to gracefully stop CloserWithContext")
				}
			case Syncer:
				if err := c.Sync(); err != nil {
					log.WithError(s.l, err).Error("failed to gracefully stop Syncer")
				}
			case SyncerWithContext:
				if err := c.Sync(timeoutCtx); err != nil {
					log.WithError(s.l, err).Error("failed to gracefully stop SyncerWithContext")
				}
			case Shutdowner:
				if err := c.Shutdown(); err != nil {
					log.WithError(s.l, err).Error("failed to gracefully stop Shutdowner")
				}
			case ShutdownerWithContext:
				if err := c.Shutdown(timeoutCtx); err != nil {
					log.WithError(s.l, err).Error("failed to gracefully stop ShutdownerWithContext")
				}
			}
		}
		return gctx.Err()
	})

	// wait for shutdown
	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.WithError(s.l, err).Error("service error")
	}

	s.l.Info("graceful shutdown complete")
}
