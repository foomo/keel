package keeltemporal

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.temporal.io/sdk/client"
)

type metricsHandler struct {
	meter metric.MeterMust
	attr  []attribute.KeyValue
}

// 	scope, _ := tally.NewRootScope(opts, time.Second)
func NewMetricsHandler(meter metric.MeterMust) client.MetricsHandler {
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
	c.inst.Add(context.Background(), v, c.attr...)
}

func (m metricsHandler) Counter(name string) client.MetricsCounter {
	return &counter{
		attr: m.attr,
		inst: m.meter.NewInt64Counter(name),
	}
}

type gaugeFloat struct {
	mu sync.RWMutex
	f  float64
}

func (of *gaugeFloat) set(v float64) {
	of.mu.Lock()
	defer of.mu.Unlock()
	of.f = v
}

func (of *gaugeFloat) get() float64 {
	of.mu.RLock()
	defer of.mu.RUnlock()
	return of.f
}

func newObservedFloat(v float64) *gaugeFloat {
	return &gaugeFloat{
		f: v,
	}
}

type gauge struct {
	inst  metric.Float64GaugeObserver
	value *gaugeFloat
}

func (c *gauge) Update(v float64) {
	c.value.set(v)
}

func (m metricsHandler) Gauge(name string) client.MetricsGauge {
	value := newObservedFloat(0)
	return &gauge{
		value: value,
		inst: m.meter.NewFloat64GaugeObserver(name, func(ctx context.Context, result metric.Float64ObserverResult) {
			v := value.get()
			result.Observe(v, m.attr...)
		}),
	}
}

type timer struct {
	inst metric.Int64Histogram
	attr []attribute.KeyValue
}

func (c *timer) Record(time.Duration) {
	c.inst.Record(context.TODO(), int64(time.Millisecond), c.attr...)
}

func (m metricsHandler) Timer(name string) client.MetricsTimer {
	return &timer{
		inst: m.meter.NewInt64Histogram(name),
		attr: m.attr,
	}
}
