//go:build !pprof

package service

import (
	"context"

	otelpyroscope "github.com/grafana/otel-profiling-go"
	"github.com/grafana/pyroscope-go"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

func NewPyroscope(l *zap.Logger, cfg pyroscope.Config) *GoRoutine {
	otel.SetTracerProvider(otelpyroscope.NewTracerProvider(otel.GetTracerProvider()))
	return NewGoRoutine(l, "pyroscope", func(ctx context.Context, l *zap.Logger) error {
		p, err := pyroscope.Start(cfg)
		if err != nil {
			return err
		}
		<-ctx.Done()
		return p.Stop()
	})
}
