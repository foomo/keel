package service_test

import (
	"context"
	"time"

	"github.com/foomo/keel"
	"github.com/foomo/keel/service"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func ExampleNewGoRoutine() {
	shutdownAfter(3 * time.Second)

	svr := keel.NewServer(
		keel.WithLogger(zap.NewExample()),
	)

	svr.AddService(
		service.NewGoRoutine(svr.Logger(), "demo", func(ctx context.Context, l *zap.Logger) error {
			for {
				if err := ctx.Err(); errors.Is(context.Cause(ctx), service.ErrServiceShutdown) {
					l.Info("context has been canceled du to graceful shutdow")
					return nil
				} else if err != nil {
					return errors.Wrap(err, "unexpected context error")
				}
				l.Info("ping")
				time.Sleep(time.Second)
			}
		}),
	)

	svr.Run()

	// Output:
	// {"level":"info","msg":"starting keel server"}
	// {"level":"info","msg":"starting keel service","keel_service_type":"goroutine","keel_service_name":"demo"}
	// {"level":"info","msg":"ping","keel_service_type":"goroutine","keel_service_name":"demo","keel_service_inst":0}
	// {"level":"info","msg":"ping","keel_service_type":"goroutine","keel_service_name":"demo","keel_service_inst":0}
	// {"level":"info","msg":"ping","keel_service_type":"goroutine","keel_service_name":"demo","keel_service_inst":0}
	// {"level":"debug","msg":"keel graceful shutdown"}
	// {"level":"info","msg":"stopping keel service","keel_service_type":"goroutine","keel_service_name":"demo"}
	// {"level":"info","msg":"context has been canceled du to graceful shutdow","keel_service_type":"goroutine","keel_service_name":"demo","keel_service_inst":0}
	// {"level":"info","msg":"keel server stopped"}
}
