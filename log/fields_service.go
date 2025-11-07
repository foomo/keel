package log

import (
	"go.uber.org/zap"
)

const (
	// Deprecated: use semconv messaging attributes instead.
	PeerServiceKey = "peer_service"
	// Deprecated: use semconv messaging attributes instead.
	ServiceTypeKey = "service_type"
	// Deprecated: use semconv messaging attributes instead.
	ServiceNameKey = "service_name"
	// Deprecated: use semconv messaging attributes instead.
	ServiceMethodKey = "service_method"
	// Deprecated: use semconv messaging attributes instead.
	ServiceNamespaceKey = "service_namespace"
	// Deprecated: use semconv messaging attributes instead.
	ServiceInstanceIDKey = "service_instance.id"
	// Deprecated: use semconv messaging attributes instead.
	ServiceVersionKey = "service_version"
)

// Deprecated: use semconv messaging attributes instead.
func FPeerService(name string) zap.Field {
	return zap.String(PeerServiceKey, name)
}

// Deprecated: use semconv messaging attributes instead.
func FServiceType(name string) zap.Field {
	return zap.String(ServiceTypeKey, name)
}

// Deprecated: use semconv messaging attributes instead.
func FServiceName(name string) zap.Field {
	return zap.String(ServiceNameKey, name)
}

// Deprecated: use semconv messaging attributes instead.
func FServiceNamespace(namespace string) zap.Field {
	return zap.String(ServiceNamespaceKey, namespace)
}

// Deprecated: use semconv messaging attributes instead.
func FServiceInstanceID(id string) zap.Field {
	return zap.String(ServiceInstanceIDKey, id)
}

// Deprecated: use semconv messaging attributes instead.
func FServiceVersion(version string) zap.Field {
	return zap.String(ServiceVersionKey, version)
}

// Deprecated: use semconv messaging attributes instead.
func FServiceMethod(method string) zap.Field {
	return zap.String(ServiceMethodKey, method)
}
