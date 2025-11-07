package gotsrpcconv

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

var (
	// addOptPool = &sync.Pool{New: func() any { return &[]metric.AddOption{} }}
	recOptPool = &sync.Pool{New: func() any { return &[]metric.RecordOption{} }}
)

// ExecutionDuration is an instrument used to record metric values conforming
// to the "gotsrpc.execution.duration" semantic conventions. It represents the
// duration of HTTP server requests.
type ExecutionDuration struct {
	metric.Float64Histogram
}

var newExecutionDurationOpts = []metric.Float64HistogramOption{
	metric.WithDescription("Duration of GOTSRPC execution."),
	metric.WithUnit("s"),
}

// NewExecutionDuration returns a new ExecutionDuration instrument.
func NewExecutionDuration(
	m metric.Meter,
	opt ...metric.Float64HistogramOption,
) (ExecutionDuration, error) {
	// Check if the meter is nil.
	if m == nil {
		return ExecutionDuration{noop.Float64Histogram{}}, nil
	}

	if len(opt) == 0 {
		opt = newExecutionDurationOpts
	} else {
		opt = append(opt, newExecutionDurationOpts...)
	}

	i, err := m.Float64Histogram(
		"gotsrpc.execution.duration",
		opt...,
	)
	if err != nil {
		return ExecutionDuration{noop.Float64Histogram{}}, err
	}

	return ExecutionDuration{i}, nil
}

// Inst returns the underlying metric instrument.
func (m ExecutionDuration) Inst() metric.Float64Histogram {
	return m.Float64Histogram
}

// Name returns the semantic convention name of the instrument.
func (ExecutionDuration) Name() string {
	return "gotsrpc.execution.duration"
}

// Unit returns the semantic convention unit of the instrument
func (ExecutionDuration) Unit() string {
	return "s"
}

// Description returns the semantic convention description of the instrument
func (ExecutionDuration) Description() string {
	return "Duration of GOTSRPC execution."
}

func (m ExecutionDuration) Record(
	ctx context.Context,
	val float64,
	pkg string,
	svs string,
	fnc string,
	attrs ...attribute.KeyValue,
) {
	if len(attrs) == 0 {
		m.Float64Histogram.Record(ctx, val)
		return
	}

	o := recOptPool.Get().(*[]metric.RecordOption) //nolint:forcetypeassert

	defer func() {
		*o = (*o)[:0]
		recOptPool.Put(o)
	}()

	*o = append(
		*o,
		metric.WithAttributes(
			append(
				attrs,
				attribute.String("gotsrpc.package", pkg),
				attribute.String("gotsrpc.service", svs),
				attribute.String("gotsrpc.func", fnc),
			)...,
		),
	)

	m.Float64Histogram.Record(ctx, val, *o...)
}

// RecordSet records val to the current distribution for set.
func (m ExecutionDuration) RecordSet(ctx context.Context, val float64, set attribute.Set) {
	if set.Len() == 0 {
		m.Float64Histogram.Record(ctx, val)
	}

	o := recOptPool.Get().(*[]metric.RecordOption) //nolint:forcetypeassert

	defer func() {
		*o = (*o)[:0]
		recOptPool.Put(o)
	}()

	*o = append(*o, metric.WithAttributeSet(set))
	m.Float64Histogram.Record(ctx, val, *o...)
}

func (ExecutionDuration) AttrError(val bool) attribute.KeyValue {
	return attribute.Bool("gotsprc.error", val)
}
