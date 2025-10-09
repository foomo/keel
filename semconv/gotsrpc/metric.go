package gotsrpc

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

// ServerRequestDuration is an instrument used to record metric values conforming
// to the "http.server.request.duration" semantic conventions. It represents the
// duration of HTTP server requests.
type ServerRequestDuration struct {
	metric.Float64Histogram
}

var newServerRequestDurationOpts = []metric.Float64HistogramOption{
	metric.WithDescription("Duration of GOTSRPC server requests."),
	metric.WithUnit("s"),
}

// NewServerRequestDuration returns a new ServerRequestDuration instrument.
func NewServerRequestDuration(
	m metric.Meter,
	opt ...metric.Float64HistogramOption,
) (ServerRequestDuration, error) {
	// Check if the meter is nil.
	if m == nil {
		return ServerRequestDuration{noop.Float64Histogram{}}, nil
	}

	if len(opt) == 0 {
		opt = newServerRequestDurationOpts
	} else {
		opt = append(opt, newServerRequestDurationOpts...)
	}

	i, err := m.Float64Histogram(
		"gotsrpc.server.request.duration",
		opt...,
	)
	if err != nil {
		return ServerRequestDuration{noop.Float64Histogram{}}, err
	}

	return ServerRequestDuration{i}, nil
}

// Inst returns the underlying metric instrument.
func (m ServerRequestDuration) Inst() metric.Float64Histogram {
	return m.Float64Histogram
}

// Name returns the semantic convention name of the instrument.
func (ServerRequestDuration) Name() string {
	return "gotsrpc.server.request.duration"
}

// Unit returns the semantic convention unit of the instrument
func (ServerRequestDuration) Unit() string {
	return "s"
}

// Description returns the semantic convention description of the instrument
func (ServerRequestDuration) Description() string {
	return "Duration of GOTSRPC server requests."
}

func (m ServerRequestDuration) Record(
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
func (m ServerRequestDuration) RecordSet(ctx context.Context, val float64, set attribute.Set) {
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

func (ServerRequestDuration) AttrError(val bool) attribute.KeyValue {
	return attribute.Bool("gotsprc.error", val)
}
