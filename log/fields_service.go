package log

import (
	"go.uber.org/zap"
)

const (
	// PeerServiceKey represents the ServiceNameKey name of the remote service.
	// Should equal the actual `service.name` resource attribute of the remote service, if any.
	PeerServiceKey = "peer_service"
)

func FPeerService(name string) zap.Field {
	return zap.String(PeerServiceKey, name)
}

const (
	ServiceTypeKey = "service_type"

	// ServiceNameKey represents the NameKey of the service.
	ServiceNameKey = "service_name"

	// ServiceMethodKey represents the Method of the service.
	ServiceMethodKey = "service_method"

	// ServiceNamespaceKey represents a namespace for `service.name`. This needs to
	// have meaning that helps to distinguish a group of services. For example, the
	// team name that owns a group of services. `service.name` is expected to be
	// unique within the same namespace.
	ServiceNamespaceKey = "service_namespace"

	// ServiceInstanceIDKey represents a unique identifier of the service instance. In conjunction
	// with the `service.name` and `service.namespace` this must be unique.
	ServiceInstanceIDKey = "service_instance.id"

	// ServiceVersionKey represents the version of the service API.
	ServiceVersionKey = "service_version"
)

func FServiceType(name string) zap.Field {
	return zap.String(ServiceTypeKey, name)
}

func FServiceName(name string) zap.Field {
	return zap.String(ServiceNameKey, name)
}

func FServiceNamespace(namespace string) zap.Field {
	return zap.String(ServiceNamespaceKey, namespace)
}

func FServiceInstanceID(id string) zap.Field {
	return zap.String(ServiceInstanceIDKey, id)
}

func FServiceVersion(version string) zap.Field {
	return zap.String(ServiceVersionKey, version)
}

func FServiceMethod(method string) zap.Field {
	return zap.String(ServiceMethodKey, method)
}
