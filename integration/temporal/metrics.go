package keeltemporal

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.temporal.io/sdk/client"
)

type metricsHandler struct {
	meter metric.Meter
	attr  []attribute.KeyValue
}

func NewMetricsHandler(meter metric.Meter) client.MetricsHandler {
	return metricsHandler{meter: meter}
}

func (m metricsHandler) WithTags(tags map[string]string) client.MetricsHandler {
	attr := make([]attribute.KeyValue, 0, len(tags))
	for k, v := range tags {
		attr = append(attr, attribute.String(k, v))
	}
	return metricsHandler{meter: m.meter, attr: attr}
}

type counter struct {
	inst metric.Int64Counter
	attr []attribute.KeyValue
}

func (c *counter) Inc(v int64) {
	c.inst.Add(context.Background(), v, metric.WithAttributes(c.attr...))
}

func (m metricsHandler) Counter(name string) client.MetricsCounter {
	c, err := m.meter.Int64Counter(name)
	if err != nil {
		otel.Handle(err)
	}
	return &counter{
		attr: m.attr,
		inst: c,
	}
}

type gauge struct {
	inst  metric.Float64ObservableGauge
	value float64
}

func (c *gauge) Update(v float64) {
	c.value = v
}

func (m metricsHandler) Gauge(name string) client.MetricsGauge {
	c, err := m.meter.Float64ObservableGauge(name)
	if err != nil {
		otel.Handle(err)
	}
	inst := &gauge{
		inst: c,
	}
	_, err = m.meter.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		o.ObserveFloat64(c, inst.value, metric.WithAttributes(m.attr...))
		return nil
	})
	if err != nil {
		otel.Handle(err)
	}
	return inst
}

type timer struct {
	inst metric.Int64Histogram
	attr []attribute.KeyValue
}

func (c *timer) Record(v time.Duration) {
	c.inst.Record(context.TODO(), v.Milliseconds(), metric.WithAttributes(c.attr...))
}

func (m metricsHandler) Timer(name string) client.MetricsTimer {
	c, err := m.meter.Int64Histogram(name)
	if err != nil {
		otel.Handle(err)
	}
	return &timer{
		inst: c,
		attr: m.attr,
	}
}
