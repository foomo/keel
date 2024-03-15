package service_test

import (
	"context"
	"sync"
	"time"

	"github.com/foomo/keel"
	"github.com/foomo/keel/service"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func ExampleNewGoRoutine() {
	var once sync.Once

	svr := keel.NewServer(
		keel.WithLogger(zap.NewExample()),
		keel.WithGracefulTimeout(time.Second),
		keel.WithShutdownTimeout(3*time.Second),
	)

	svr.AddService(
		service.NewGoRoutine(svr.Logger(), "demo", func(ctx context.Context, l *zap.Logger) error {
			for {
				// handle graceful shutdowns
				if err := ctx.Err(); errors.Is(context.Cause(ctx), service.ErrServiceShutdown) {
					l.Info("context has been canceled du to graceful shutdow")
					return nil
				} else if err != nil {
					return errors.Wrap(err, "unexpected context error")
				}

				l.Info("ping")
				time.Sleep(700 * time.Millisecond)
				once.Do(shutdown)
			}
		}),
	)

	svr.Run()

	// Output:
	// {"level":"info","msg":"starting keel server"}
	// {"level":"info","msg":"starting keel service","keel_service_type":"goroutine","keel_service_name":"demo"}
	// {"level":"info","msg":"ping","keel_service_type":"goroutine","keel_service_name":"demo","keel_service_inst":0}
	// {"level":"info","msg":"ping","keel_service_type":"goroutine","keel_service_name":"demo","keel_service_inst":0}
	// {"level":"info","msg":"keel graceful shutdown"}
	// {"level":"info","msg":"keel graceful shutdown timeout","graceful_timeout":"1s","shutdown_timeout":"3s"}
	// {"level":"info","msg":"ping","keel_service_type":"goroutine","keel_service_name":"demo","keel_service_inst":0}
	// {"level":"info","msg":"keel graceful shutdown timeout complete"}
	// {"level":"info","msg":"keel graceful shutdown closers"}
	// {"level":"info","msg":"stopping keel service","keel_service_type":"goroutine","keel_service_name":"demo"}
	// {"level":"info","msg":"context has been canceled du to graceful shutdow","keel_service_type":"goroutine","keel_service_name":"demo","keel_service_inst":0}
	// {"level":"debug","msg":"keel graceful shutdown closer closed","name":"*service.GoRoutine"}
	// {"level":"debug","msg":"keel graceful shutdown closer closed","name":"noop.TracerProvider"}
	// {"level":"debug","msg":"keel graceful shutdown closer closed","name":"noop.MeterProvider"}
	// {"level":"info","msg":"keel graceful shutdown complete"}
	// {"level":"info","msg":"keel server stopped"}
}
