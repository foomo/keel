package telemetry

import (
	"context"
	"reflect"
	"runtime"
	"runtime/pprof"
	"strings"

	"github.com/foomo/keel/log"
	"github.com/grafana/pyroscope-go"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Context struct {
	context.Context
}

func Ctx(ctx context.Context) Context {
	return Context{ctx}
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

// Log returns the logger from the context.
func (c Context) Log() *zap.Logger {
	return Log(c.Context)
}

// LogDebug logs a message at debug level.
func (c Context) LogDebug(msg string, kv ...attribute.KeyValue) {
	Log(c.Context).Debug(msg, log.Attributes(kv...)...)
}

// LogInfo logs a message at info level.
func (c Context) LogInfo(msg string, kv ...attribute.KeyValue) {
	Log(c.Context).Info(msg, log.Attributes(kv...)...)
}

// LogWarn logs a message at warn level.
func (c Context) LogWarn(msg string, kv ...attribute.KeyValue) {
	Log(c.Context).Warn(msg, log.Attributes(kv...)...)
}

// LogError logs a message at error level.
func (c Context) LogError(msg string, kv ...attribute.KeyValue) {
	Log(c.Context).Error(msg, log.Attributes(kv...)...)
}

// Span returns the span from the context.
func (c Context) Span() trace.Span {
	return trace.SpanFromContext(c.Context)
}

// EndSpan ends the span.
func (c Context) EndSpan(err error, opts ...trace.SpanEndOption) {
	sp := c.Span()
	if sp.IsRecording() {
		if err != nil {
			c.RecordError(err)
		} else {
			c.SetSpanStatusOK()
		}

		sp.End(opts...)
	}
}

// SetSpanStatusOK sets the status of the span to ok.
func (c Context) SetSpanStatusOK() {
	sp := c.Span()
	if sp.IsRecording() {
		sp.SetStatus(codes.Ok, "")
	}
}

// SetSpanStatusError sets the status of the span to error.
func (c Context) SetSpanStatusError(description string) {
	sp := c.Span()
	if sp.IsRecording() {
		sp.SetStatus(codes.Error, description)
	}
}

// SetSpanName sets the name of the span.
func (c Context) SetSpanName(name string) {
	sp := c.Span()
	if sp.IsRecording() {
		sp.SetName(name)
	}
}

// SetSpanAttributes sets the attributes of the span.
func (c Context) SetSpanAttributes(kv ...attribute.KeyValue) {
	sp := c.Span()
	if sp.IsRecording() {
		sp.SetAttributes(kv...)
	}
}

// RecordError records an error on the span and logs it.
func (c Context) RecordError(err error, kv ...attribute.KeyValue) {
	c.RecordSpanError(err, trace.WithAttributes(kv...))
	c.SetSpanStatusError(errors.Cause(err).Error())
	c.Log().With(append(log.Attributes(kv...), zap.Error(err))...).Error(errors.Cause(err).Error())
}

// RecordSpanError records an error on the span.
func (c Context) RecordSpanError(err error, opts ...trace.EventOption) {
	sp := c.Span()
	if sp.IsRecording() {
		sp.RecordError(err, append(opts, trace.WithStackTrace(true))...)
	}
}

// AddSpanEvent adds an event to the span.
func (c Context) AddSpanEvent(name string, opts ...trace.EventOption) {
	sp := c.Span()
	if sp.IsRecording() {
		sp.AddEvent(name, opts...)
	}
}

// StartSpan starts a span.
func (c Context) StartSpan(opts ...trace.SpanStartOption) Context {
	ctx, _ := c.startSpan("CODE", 2, opts...)
	return ctx
}

// StartSpanWithProfile starts a span and profiles the handler.
func (c Context) StartSpanWithProfile(handler func(ctx Context), kv ...attribute.KeyValue) {
	ctx, span := c.startSpan("PROFILE", 2, trace.WithAttributes(kv...))
	defer span.End()

	ctx.StartProfile(handler, kv...)
}

// StartProfile starts a profile for the handler.
func (c Context) StartProfile(handler func(ctx Context), kv ...attribute.KeyValue) {
	pyroscope.TagWrapper(c.Context, PyroscopeLabels(kv...), func(ctx context.Context) {
		handler(Ctx(ctx))
	})
}

// SetProfileAttributes sets the labels for the profile.
func (c Context) SetProfileAttributes(kv ...attribute.KeyValue) Context {
	ctx := pprof.WithLabels(c.Context, PyroscopeLabels(kv...))
	pprof.SetGoroutineLabels(ctx)

	return Ctx(ctx)
}

// IntHistogram creates and returns a Int64Histogram metric instrument with the specified name and optional settings.
func (c Context) IntHistogram(name string, opts ...any) metric.Int64Histogram {
	var (
		meterOptions  []metric.MeterOption
		metricOptions []metric.Int64HistogramOption
	)

	for _, v := range opts {
		switch t := v.(type) {
		case metric.MeterOption:
			meterOptions = append(meterOptions, t)
		case metric.HistogramOption:
			metricOptions = append(metricOptions, t)
		case metric.Int64HistogramOption:
			metricOptions = append(metricOptions, t)
		default:
			c.LogWarn("invalid Histogram option", attribute.String("type", reflect.TypeOf(v).String()))
		}
	}

	m, err := Meter(meterOptions...).Int64Histogram(name, metricOptions...)
	if err != nil {
		otel.Handle(err)
		return noop.Int64Histogram{}
	}

	return m
}

// FloatHistogram creates and returns a Float64Histogram metric instrument with the specified name and optional settings.
func (c Context) FloatHistogram(name string, opts ...any) metric.Float64Histogram {
	var (
		meterOptions  []metric.MeterOption
		metricOptions []metric.Float64HistogramOption
	)

	for _, v := range opts {
		switch t := v.(type) {
		case metric.MeterOption:
			meterOptions = append(meterOptions, t)
		case metric.HistogramOption:
			metricOptions = append(metricOptions, t)
		case metric.Float64HistogramOption:
			metricOptions = append(metricOptions, t)
		default:
			c.LogWarn("invalid Histogram option", attribute.String("type", reflect.TypeOf(v).String()))
		}
	}

	m, err := Meter(meterOptions...).Float64Histogram(name, metricOptions...)
	if err != nil {
		otel.Handle(err)
		return noop.Float64Histogram{}
	}

	return m
}

// IntGauge creates and returns a Int64Gauge metric instrument with the specified name and optional settings.
func (c Context) IntGauge(name string, opts ...any) metric.Int64Gauge {
	var (
		meterOptions  []metric.MeterOption
		metricOptions []metric.Int64GaugeOption
	)

	for _, v := range opts {
		switch t := v.(type) {
		case metric.MeterOption:
			meterOptions = append(meterOptions, t)
		case metric.Int64GaugeOption:
			metricOptions = append(metricOptions, t)
		default:
			c.LogWarn("invalid Gauge option", attribute.String("type", reflect.TypeOf(v).String()))
		}
	}

	m, err := Meter(meterOptions...).Int64Gauge(name, metricOptions...)
	if err != nil {
		c.LogWarn("failed to create Gauge", semconv.ErrorType(err), semconv.ErrorMessage(err.Error()))
		return noop.Int64Gauge{}
	}

	return m
}

// FloatGauge creates and returns a Float64Gauge metric with the specified name and optional configurations.
func (c Context) FloatGauge(name string, opts ...any) metric.Float64Gauge {
	var (
		meterOptions  []metric.MeterOption
		metricOptions []metric.Float64GaugeOption
	)

	for _, v := range opts {
		switch t := v.(type) {
		case metric.MeterOption:
			meterOptions = append(meterOptions, t)
		case metric.Float64GaugeOption:
			metricOptions = append(metricOptions, t)
		default:
			c.LogWarn("invalid Gauge option", attribute.String("type", reflect.TypeOf(v).String()))
		}
	}

	m, err := Meter(meterOptions...).Float64Gauge(name, metricOptions...)
	if err != nil {
		c.LogWarn("failed to create Gauge", semconv.ErrorType(err), semconv.ErrorMessage(err.Error()))
		return noop.Float64Gauge{}
	}

	return m
}

// IntCounter creates and returns a Int64Counter metric instrument with the specified name and optional settings.
func (c Context) IntCounter(name string, opts ...any) metric.Int64Counter {
	var (
		meterOptions  []metric.MeterOption
		metricOptions []metric.Int64CounterOption
	)

	for _, v := range opts {
		switch t := v.(type) {
		case metric.MeterOption:
			meterOptions = append(meterOptions, t)
		case metric.Int64CounterOption:
			metricOptions = append(metricOptions, t)
		default:
			c.LogWarn("invalid Counter option", attribute.String("type", reflect.TypeOf(v).String()))
		}
	}

	m, err := Meter(meterOptions...).Int64Counter(name, metricOptions...)
	if err != nil {
		c.LogWarn("failed to create Counter", semconv.ErrorType(err), semconv.ErrorMessage(err.Error()))
		return noop.Int64Counter{}
	}

	return m
}

// FloatCounter creates and returns a Float64Counter metric instrument with the specified name and optional settings.
func (c Context) FloatCounter(name string, opts ...any) metric.Float64Counter {
	var (
		meterOptions  []metric.MeterOption
		metricOptions []metric.Float64CounterOption
	)

	for _, v := range opts {
		switch t := v.(type) {
		case metric.MeterOption:
			meterOptions = append(meterOptions, t)
		case metric.Float64CounterOption:
			metricOptions = append(metricOptions, t)
		default:
			c.LogWarn("invalid Counter option", attribute.String("type", reflect.TypeOf(v).String()))
		}
	}

	m, err := Meter(meterOptions...).Float64Counter(name, metricOptions...)
	if err != nil {
		c.LogWarn("failed to create Counter", semconv.ErrorType(err), semconv.ErrorMessage(err.Error()))
		return noop.Float64Counter{}
	}

	return m
}

// IntUpDownCounter creates and returns a Int64UpDownCounter metric instrument with the specified name and optional settings.
func (c Context) IntUpDownCounter(name string, opts ...any) metric.Int64UpDownCounter {
	var (
		meterOptions  []metric.MeterOption
		metricOptions []metric.Int64UpDownCounterOption
	)

	for _, v := range opts {
		switch t := v.(type) {
		case metric.MeterOption:
			meterOptions = append(meterOptions, t)
		case metric.Int64UpDownCounterOption:
			metricOptions = append(metricOptions, t)
		default:
			c.LogWarn("invalid UpDownCounter option", attribute.String("type", reflect.TypeOf(v).String()))
		}
	}

	m, err := Meter(meterOptions...).Int64UpDownCounter(name, metricOptions...)
	if err != nil {
		c.LogWarn("failed to create UpDownCounter", semconv.ErrorType(err), semconv.ErrorMessage(err.Error()))
		return noop.Int64UpDownCounter{}
	}

	return m
}

// FloatUpDownCounter creates and returns a Float64UpDownCounter metric instrument with the specified name and optional settings.
func (c Context) FloatUpDownCounter(name string, opts ...any) metric.Float64UpDownCounter {
	var (
		meterOptions  []metric.MeterOption
		metricOptions []metric.Float64UpDownCounterOption
	)

	for _, v := range opts {
		switch t := v.(type) {
		case metric.MeterOption:
			meterOptions = append(meterOptions, t)
		case metric.Float64UpDownCounterOption:
			metricOptions = append(metricOptions, t)
		default:
			c.LogWarn("invalid UpDownCounter option", attribute.String("type", reflect.TypeOf(v).String()))
		}
	}

	m, err := Meter(meterOptions...).Float64UpDownCounter(name, metricOptions...)
	if err != nil {
		c.LogWarn("failed to create UpDownCounter", semconv.ErrorType(err), semconv.ErrorMessage(err.Error()))
		return noop.Float64UpDownCounter{}
	}

	return m
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c Context) startSpan(prefix string, skip int, opts ...trace.SpanStartOption) (Context, trace.Span) {
	name := prefix

	var attrs []attribute.KeyValue
	if pc, file, line, ok := runtime.Caller(skip); ok {
		attrs = append(attrs,
			semconv.CodeLineNumber(line),
			semconv.CodeFilePath(file),
		)

		if details := runtime.FuncForPC(pc); details != nil {
			funcName := details.Name()

			lastSlash := strings.LastIndexByte(funcName, '/')
			if lastSlash < 0 {
				lastSlash = 0
			}

			lastDot := strings.LastIndexByte(funcName[lastSlash:], '.') + lastSlash
			name += " " + funcName[lastDot+1:]
		}
	}

	ctx, span := Tracer().Start(c.Context, name, append(opts, trace.WithAttributes(attrs...))...) //nolint:spancheck

	return Ctx(ctx), span //nolint:spancheck
}
