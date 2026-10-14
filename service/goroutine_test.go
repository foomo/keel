package service_test

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/foomo/keel"
	"github.com/foomo/keel/service"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func ExampleNewGoRoutine() {
	var once sync.Once
	// unset for github actions
	_ = os.Setenv("GITHUB_REPOSITORY", "")

	svr := keel.NewServer(
		keel.WithLogger(zap.NewExample()),
		keel.WithGracefulPeriod(3*time.Second),
	)

	svr.AddService(
		service.NewGoRoutine(svr.Logger(), "demo", func(ctx context.Context, l *zap.Logger) error {
			for {
				// handle graceful shutdowns
				if err := ctx.Err(); errors.Is(context.Cause(ctx), service.ErrServiceShutdown) {
					l.Info("context has been canceled du to graceful shutdown")
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
	// {"level":"info","msg":"keel closer closed","keel_graceful_period":3}
	// {"level":"info","msg":"keel closer closed: closers"}
	// {"level":"info","msg":"stopping keel service","keel_service_type":"goroutine","keel_service_name":"demo"}
	// {"level":"info","msg":"context has been canceled du to graceful shutdown","keel_service_type":"goroutine","keel_service_name":"demo","keel_service_inst":0}
	// {"level":"debug","msg":"keel closer closed","keel_closer_type":"*service.GoRoutine"}
	// {"level":"debug","msg":"keel closer closed","keel_closer_type":"noop.TracerProvider"}
	// {"level":"debug","msg":"keel closer closed","keel_closer_type":"noop.MeterProvider"}
	// {"level":"debug","msg":"keel closer closed","keel_closer_type":"noop.LoggerProvider"}
	// {"level":"info","msg":"keel closer closed: complete"}
	// {"level":"info","msg":"keel server stopped"}
}

func TestGoRoutine_CloseBeforeStart(t *testing.T) {
	t.Parallel()

	s := service.NewGoRoutine(zap.NewNop(), "test", func(ctx context.Context, l *zap.Logger) error {
		<-ctx.Done()
		return context.Cause(ctx)
	})

	require.NoError(t, s.Close(t.Context()))
	require.ErrorIs(t, s.Start(t.Context()), service.ErrServiceShutdown)
}
