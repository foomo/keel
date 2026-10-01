package keeltemporal

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.temporal.io/sdk/client"
)

// metricsHandler implements [client.MetricsHandler] on top of an
// OpenTelemetry meter.
type metricsHandler struct {
	meter metric.Meter
	attr  []attribute.KeyValue
}

// NewMetricsHandler returns a Temporal metrics handler that records counters,
// gauges and timers with meter. Instrument creation errors are reported via
// [otel.Handle].
func NewMetricsHandler(meter metric.Meter) client.MetricsHandler {
	return metricsHandler{meter: meter}
}

// WithTags returns a handler whose instruments record tags as string
// attributes. Tags replace any previously set tags.
func (m metricsHandler) WithTags(tags map[string]string) client.MetricsHandler {
	attr := make([]attribute.KeyValue, 0, len(tags))
	for k, v := range tags {
		attr = append(attr, attribute.String(k, v))
	}

	return metricsHandler{meter: m.meter, attr: attr}
}

// counter adapts an Int64Counter to [client.MetricsCounter].
type counter struct {
	inst metric.Int64Counter
	attr []attribute.KeyValue
}

// Inc adds v to the counter.
func (c *counter) Inc(v int64) {
	c.inst.Add(context.Background(), v, metric.WithAttributes(c.attr...))
}

// Counter returns an Int64Counter instrument with the given name.
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

// gauge adapts a Float64ObservableGauge to [client.MetricsGauge]; the last
// updated value is reported on each collection.
type gauge struct {
	inst  metric.Float64ObservableGauge
	value float64
}

// Update sets the value reported by the gauge.
func (c *gauge) Update(v float64) {
	c.value = v
}

// Gauge returns a Float64ObservableGauge instrument with the given name.
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

// timer adapts an Int64Histogram to [client.MetricsTimer].
type timer struct {
	inst metric.Int64Histogram
	attr []attribute.KeyValue
}

// Record records v in milliseconds.
func (c *timer) Record(v time.Duration) {
	c.inst.Record(context.TODO(), v.Milliseconds(), metric.WithAttributes(c.attr...))
}

// Timer returns an Int64Histogram instrument with the given name that records
// durations in milliseconds.
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
