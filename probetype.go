package keel

type ProbeType string

const probesServiceName = "probes"

const (
	ProbeTypeAll ProbeType = "all"
	// https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/
	// Startup probe disables liveness and readiness checks until it succeeds
	ProbeTypeStartup ProbeType = "startup"
	// Kubernetes uses readiness probes to know when a container is ready to start accepting traffic
	ProbeTypeReadiness ProbeType = "readiness"
	// Kubernetes uses liveness probes to know when to restart a container
	ProbeTypeLiveliness ProbeType = "liveliness"
)

func (t ProbeType) String() string {
	return string(t)
}
