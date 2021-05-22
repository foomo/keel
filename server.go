package keel

import (
	"context"
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
	"github.com/foomo/keel/telemetry"
)

// Server struct
type Server struct {
	services        []Service
	shutdownTimeout time.Duration
	closers         []interface{}
	ctx             context.Context
	cancelFn        context.CancelFunc
	l               *zap.Logger
	c               *viper.Viper
}

func NewServer(opts ...Option) *Server {

	// context
	defaultCtx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt, // kill -SIGINT XXXX or Ctrl+c
		// os.Kill,				// kill -SIGINT XXXX or Ctrl+c // no graceful shutdown if system needs to kill process // @TODO optional interrupts
		syscall.SIGHUP,  // kill -SIGHUP XXXX
		syscall.SIGTERM, // kill -SIGTERM XXXX
		syscall.SIGQUIT, // kill -SIGQUIT XXXX
	)

	var (
		defaultShutdownTimeout = 5 * time.Second
		defaultConfig          = config.Config()
		defaultLogger          = log.Logger()
	)

	inst := &Server{
		shutdownTimeout: defaultShutdownTimeout,
		ctx:             defaultCtx,
		cancelFn:        cancel,
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

// ContextCancel returns server context cancel func
func (s *Server) ContextCancel() context.CancelFunc {
	return s.cancelFn
}

func (s *Server) Shutdown() {
	s.cancelFn()
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

	g, gctx := errgroup.WithContext(s.ctx)

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

		// append internal closers
		closers := append(s.closers, log.Logger())
		if provider := telemetry.Provider(); provider != nil {
			closers = append(closers, provider)
		}

		for _, closer := range closers {
			if closer != nil {
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
		}
		return gctx.Err()
	})

	// wait for shutdown
	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.WithError(s.l, err).Error("service error")
	}

	s.l.Info("graceful shutdown complete")
}
