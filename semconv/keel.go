package semconv

import (
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

const (
	// KeelServiceTypeKey is the key for keel.service.type.
	KeelServiceTypeKey = attribute.Key("keel.service.type")
	// KeelServiceNameKey is the key for keel.service.name.
	KeelServiceNameKey = attribute.Key("keel.service.name")
	// KeelServiceInstKey is the key for keel.service.inst.
	KeelServiceInstKey = attribute.Key("keel.service.inst")
	// KeelGracefulPeriodKey is the key for keel.graceful_period, in seconds.
	KeelGracefulPeriodKey = attribute.Key("keel.graceful_period")
	// KeelJobStepKey is the key for keel.job.step.
	KeelJobStepKey = attribute.Key("keel.job.step")
	// KeelCloserTypeKey is the key for keel.closer.type.
	KeelCloserTypeKey = attribute.Key("keel.closer.type")
	// KeelHealthzProbeTypeKey is the key for keel.healthz.probe.type.
	KeelHealthzProbeTypeKey = attribute.Key("keel.healthz.probe.type")
)

// KeelGracefulPeriod returns a new attribute.KeyValue for keel.graceful_period in seconds.
func KeelGracefulPeriod(v time.Duration) attribute.KeyValue {
	return KeelGracefulPeriodKey.Float64(v.Seconds())
}

// KeelJobStep returns a new attribute.KeyValue for keel.job.step.
func KeelJobStep(v string) attribute.KeyValue {
	return KeelJobStepKey.String(v)
}

// KeelCloserType returns a new attribute.KeyValue for keel.closer.type with
// the Go type name of closer.
func KeelCloserType(closer any) attribute.KeyValue {
	return KeelCloserTypeKey.String(fmt.Sprintf("%T", closer))
}

// KeelHealthzProbeType returns a new attribute.KeyValue for
// keel.healthz.probe.type with the Go type name of probe.
func KeelHealthzProbeType(probe any) attribute.KeyValue {
	return KeelHealthzProbeTypeKey.String(fmt.Sprintf("%T", probe))
}

// KeelServiceTypeAttr is an enum value of keel.service.type.
type KeelServiceTypeAttr string

const (
	// KeelServiceTypeHTTP is the keel.service.type of an HTTP service.
	KeelServiceTypeHTTP KeelServiceTypeAttr = "http"
	// KeelServiceTypeGoRoutine is the keel.service.type of a goroutine service.
	KeelServiceTypeGoRoutine KeelServiceTypeAttr = "goroutine"
	// KeelServiceTypeSubscription is the keel.service.type of a subscription service.
	KeelServiceTypeSubscription KeelServiceTypeAttr = "sub"
	// KeelServiceTypeJob is the keel.service.type of a job.
	KeelServiceTypeJob KeelServiceTypeAttr = "job"
)

// KeelServiceType returns a new attribute.KeyValue for keel.service.type.
func KeelServiceType(v KeelServiceTypeAttr) attribute.KeyValue {
	return KeelServiceTypeKey.String(string(v))
}

// KeelServiceName returns a new attribute.KeyValue for keel.service.name.
func KeelServiceName(v string) attribute.KeyValue {
	return KeelServiceNameKey.String(v)
}

// KeelServiceInst returns a new attribute.KeyValue for keel.service.inst.
func KeelServiceInst(v int) attribute.KeyValue {
	return KeelServiceInstKey.Int(v)
}
