package keel

// ProbeType type
// https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/
type ProbeType string

const probesServiceName = "probes"

const (
	// ProbeTypeAny will run on any checks
	ProbeTypeAny ProbeType = "any"
	// ProbeTypeStartup will run on /healthz/startup checks
	// > The kubelet uses startup probes to know when a container application has started. If such a probe is configured,
	// > it disables liveness and readiness checks until it succeeds, making sure those probes don't interfere with the
	// > application startup. This can be used to adopt liveness checks on slow starting containers, avoiding them getting
	// > killed by the kubelet before they are up and running.
	ProbeTypeStartup ProbeType = "startup"
	// ProbeTypeReadiness will run on /healthz/readiness checks
	// > The kubelet uses readiness probes to know when a container is ready to start accepting traffic.
	// > A Pod is considered ready when all of its containers are ready. One use of this signal is to control
	// > which Pods are used as backends for Services. When a Pod is not ready, it is removed from Service load balancers.
	ProbeTypeReadiness ProbeType = "readiness"
	// ProbeTypeLiveness  will run on /healthz/liveness checks
	// > The kubelet uses liveness probes to know when to restart a container. For example, liveness probes could catch
	// > a deadlock, where an application is running, but unable to make progress. Restarting a container in such a state
	// > can help to make the application more available despite bugs.
	ProbeTypeLiveness ProbeType = "liveness"
)

// String interface
func (t ProbeType) String() string {
	return string(t)
}
