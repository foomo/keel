package nonrecording

import (
	"context"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/asyncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/asyncint64"
	"go.opentelemetry.io/otel/metric/instrument/syncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
)

// NewNoopMeterProvider creates a MeterProvider that does not record any metrics.
func NewNoopMeterProvider() metric.MeterProvider {
	return noopMeterProvider{}
}

type noopMeterProvider struct{}

var _ metric.MeterProvider = noopMeterProvider{}

func (noopMeterProvider) Meter(instrumentationName string, opts ...metric.MeterOption) metric.Meter {
	return noopMeter{}
}

// NewNoopMeter creates a Meter that does not record any metrics.
func NewNoopMeter() metric.Meter {
	return noopMeter{}
}

type noopMeter struct{}

var _ metric.Meter = noopMeter{}

func (noopMeter) AsyncInt64() asyncint64.InstrumentProvider {
	return nonrecordingAsyncInt64Instrument{}
}
func (noopMeter) AsyncFloat64() asyncfloat64.InstrumentProvider {
	return nonrecordingAsyncFloat64Instrument{}
}
func (noopMeter) SyncInt64() syncint64.InstrumentProvider {
	return nonrecordingSyncInt64Instrument{}
}
func (noopMeter) SyncFloat64() syncfloat64.InstrumentProvider {
	return nonrecordingSyncFloat64Instrument{}
}
func (noopMeter) RegisterCallback([]instrument.Asynchronous, func(context.Context)) error {
	return nil
}
