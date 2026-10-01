package log

import (
	"go.uber.org/zap"
)

const (
	// PeerServiceKey is the log field key "peer_service".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	PeerServiceKey = "peer_service"
	// ServiceTypeKey is the log field key "service_type".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	ServiceTypeKey = "service_type"
	// ServiceNameKey is the log field key "service_name".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	ServiceNameKey = "service_name"
	// ServiceMethodKey is the log field key "service_method".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	ServiceMethodKey = "service_method"
	// ServiceNamespaceKey is the log field key "service_namespace".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	ServiceNamespaceKey = "service_namespace"
	// ServiceInstanceIDKey is the log field key "service_instance.id".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	ServiceInstanceIDKey = "service_instance.id"
	// ServiceVersionKey is the log field key "service_version".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	ServiceVersionKey = "service_version"
)

// FPeerService returns a field with the given value under [PeerServiceKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FPeerService(name string) zap.Field {
	return zap.String(PeerServiceKey, name)
}

// FServiceType returns a field with the given value under [ServiceTypeKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FServiceType(name string) zap.Field {
	return zap.String(ServiceTypeKey, name)
}

// FServiceName returns a field with the given value under [ServiceNameKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FServiceName(name string) zap.Field {
	return zap.String(ServiceNameKey, name)
}

// FServiceNamespace returns a field with the given value under [ServiceNamespaceKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FServiceNamespace(namespace string) zap.Field {
	return zap.String(ServiceNamespaceKey, namespace)
}

// FServiceInstanceID returns a field with the given value under [ServiceInstanceIDKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FServiceInstanceID(id string) zap.Field {
	return zap.String(ServiceInstanceIDKey, id)
}

// FServiceVersion returns a field with the given value under [ServiceVersionKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FServiceVersion(version string) zap.Field {
	return zap.String(ServiceVersionKey, version)
}

// FServiceMethod returns a field with the given value under [ServiceMethodKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FServiceMethod(method string) zap.Field {
	return zap.String(ServiceMethodKey, method)
}
