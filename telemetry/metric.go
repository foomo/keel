package telemetry

import (
	"context"

	pkgsemconv "github.com/foomo/keel/semconv"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

// NewIntHistogram creates and returns a Int64Histogram metric instrument with the specified name and optional settings.
func NewIntHistogram(name string, opts ...any) metric.Int64Histogram {
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
			Ctx(context.TODO()).LogWarn("invalid Histogram option", pkgsemconv.RefectType(v))
		}
	}

	m, err := Meter(meterOptions...).Int64Histogram(name, metricOptions...)
	if err != nil {
		otel.Handle(err)
		return noop.Int64Histogram{}
	}

	return m
}

// NewFloatHistogram creates and returns a Float64Histogram metric instrument with the specified name and optional settings.
func NewFloatHistogram(name string, opts ...any) metric.Float64Histogram {
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
			Ctx(context.TODO()).LogWarn("invalid Histogram option", pkgsemconv.RefectType(v))
		}
	}

	m, err := Meter(meterOptions...).Float64Histogram(name, metricOptions...)
	if err != nil {
		otel.Handle(err)
		return noop.Float64Histogram{}
	}

	return m
}

// NewIntGauge creates and returns a Int64Gauge metric instrument with the specified name and optional settings.
func NewIntGauge(name string, opts ...any) metric.Int64Gauge {
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
			Ctx(context.TODO()).LogWarn("invalid Gauge option", pkgsemconv.RefectType(v))
		}
	}

	m, err := Meter(meterOptions...).Int64Gauge(name, metricOptions...)
	if err != nil {
		Ctx(context.TODO()).LogWarn("failed to create Gauge", semconv.ErrorType(err), semconv.ErrorMessage(err.Error()))
		return noop.Int64Gauge{}
	}

	return m
}

// NewFloatGauge creates and returns a Float64Gauge metric with the specified name and optional configurations.
func NewFloatGauge(name string, opts ...any) metric.Float64Gauge {
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
			Ctx(context.TODO()).LogWarn("invalid Gauge option", pkgsemconv.RefectType(v))
		}
	}

	m, err := Meter(meterOptions...).Float64Gauge(name, metricOptions...)
	if err != nil {
		Ctx(context.TODO()).LogWarn("failed to create Gauge", semconv.ErrorType(err), semconv.ErrorMessage(err.Error()))
		return noop.Float64Gauge{}
	}

	return m
}

// NewIntCounter creates and returns a Int64Counter metric instrument with the specified name and optional settings.
func NewIntCounter(name string, opts ...any) metric.Int64Counter {
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
			Ctx(context.TODO()).LogWarn("invalid Counter option", pkgsemconv.RefectType(v))
		}
	}

	m, err := Meter(meterOptions...).Int64Counter(name, metricOptions...)
	if err != nil {
		Ctx(context.TODO()).LogWarn("failed to create Counter", semconv.ErrorType(err), semconv.ErrorMessage(err.Error()))
		return noop.Int64Counter{}
	}

	return m
}

// NewFloatCounter creates and returns a Float64Counter metric instrument with the specified name and optional settings.
func NewFloatCounter(name string, opts ...any) metric.Float64Counter {
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
			Ctx(context.TODO()).LogWarn("invalid Counter option", pkgsemconv.RefectType(v))
		}
	}

	m, err := Meter(meterOptions...).Float64Counter(name, metricOptions...)
	if err != nil {
		Ctx(context.TODO()).LogWarn("failed to create Counter", semconv.ErrorType(err), semconv.ErrorMessage(err.Error()))
		return noop.Float64Counter{}
	}

	return m
}

// NewIntUpDownCounter creates and returns a Int64UpDownCounter metric instrument with the specified name and optional settings.
func NewIntUpDownCounter(name string, opts ...any) metric.Int64UpDownCounter {
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
			Ctx(context.TODO()).LogWarn("invalid UpDownCounter option", pkgsemconv.RefectType(v))
		}
	}

	m, err := Meter(meterOptions...).Int64UpDownCounter(name, metricOptions...)
	if err != nil {
		Ctx(context.TODO()).LogWarn("failed to create UpDownCounter", semconv.ErrorType(err), semconv.ErrorMessage(err.Error()))
		return noop.Int64UpDownCounter{}
	}

	return m
}

// NewFloatUpDownCounter creates and returns a Float64UpDownCounter metric instrument with the specified name and optional settings.
func NewFloatUpDownCounter(name string, opts ...any) metric.Float64UpDownCounter {
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
			Ctx(context.TODO()).LogWarn("invalid UpDownCounter option", pkgsemconv.RefectType(v))
		}
	}

	m, err := Meter(meterOptions...).Float64UpDownCounter(name, metricOptions...)
	if err != nil {
		Ctx(context.TODO()).LogWarn("failed to create UpDownCounter", semconv.ErrorType(err), semconv.ErrorMessage(err.Error()))
		return noop.Float64UpDownCounter{}
	}

	return m
}
