package keeltemporal

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument/asyncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	"go.temporal.io/sdk/client"
)

type metricsHandler struct {
	meter metric.Meter
	attr  []attribute.KeyValue
}

// 	scope, _ := tally.NewRootScope(opts, time.Second)
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
	inst syncint64.Counter
	attr []attribute.KeyValue
}

func (c *counter) Inc(v int64) {
	c.inst.Add(context.Background(), v, c.attr...)
}

func (m metricsHandler) Counter(name string) client.MetricsCounter {
	c, err := m.meter.SyncInt64().Counter(name)
	if err != nil {
		otel.Handle(err)
	}
	return &counter{
		attr: m.attr,
		inst: c,
	}
}

type gauge struct {
	inst asyncfloat64.Gauge
	attr []attribute.KeyValue
}

func (c *gauge) Update(v float64) {
	c.inst.Observe(context.TODO(), v, c.attr...)
}

func (m metricsHandler) Gauge(name string) client.MetricsGauge {
	c, err := m.meter.AsyncFloat64().Gauge(name)
	if err != nil {
		otel.Handle(err)
	}
	return &gauge{
		inst: c,
		attr: m.attr,
	}
}

type timer struct {
	inst syncint64.Histogram
	attr []attribute.KeyValue
}

func (c *timer) Record(v time.Duration) {
	c.inst.Record(context.TODO(), v.Milliseconds(), c.attr...)
}

func (m metricsHandler) Timer(name string) client.MetricsTimer {
	c, err := m.meter.SyncInt64().Histogram(name)
	if err != nil {
		otel.Handle(err)
	}
	return &timer{
		inst: c,
		attr: m.attr,
	}
}
