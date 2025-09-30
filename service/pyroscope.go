//go:build !pprof

package service

import (
	"context"
	"sync/atomic"

	otelpyroscope "github.com/grafana/otel-profiling-go"
	"github.com/grafana/pyroscope-go"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type Pyroscope struct {
	l        *zap.Logger
	name     string
	cfg      pyroscope.Config
	running  atomic.Bool
	profiler *pyroscope.Profiler
}

func NewPyroscope(l *zap.Logger, cfg pyroscope.Config) *GoRoutine {
	otel.SetTracerProvider(otelpyroscope.NewTracerProvider(otel.GetTracerProvider()))
	return NewGoRoutine(l, "pyroscope", func(ctx context.Context, l *zap.Logger) error {
		p, err := pyroscope.Start(cfg)
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return p.Stop()
		}
	})
}
