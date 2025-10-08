package log

import (
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.uber.org/zap"
)

const (
	// Deprecated: use semconv messaging attributes instead.
	NetHostIPKey = "net_host_ip"
	// Deprecated: use semconv messaging attributes instead.
	NetHostPortKey = "net_host_port"
)

// Deprecated: use semconv messaging attributes instead.
func FNetHostIP(ip string) zap.Field {
	return Attribute(semconv.HostIP(ip))
}

// Deprecated: use semconv messaging attributes instead.
func FNetHostPort(port string) zap.Field {
	return zap.String(NetHostPortKey, port)
}
