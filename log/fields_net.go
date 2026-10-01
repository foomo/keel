package log

import (
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.uber.org/zap"
)

const (
	// NetHostIPKey is the log field key "net_host_ip".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	NetHostIPKey = "net_host_ip"
	// NetHostPortKey is the log field key "net_host_port".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	NetHostPortKey = "net_host_port"
)

// FNetHostIP returns the semconv.HostIP attribute as a field.
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FNetHostIP(ip string) zap.Field {
	return Attribute(semconv.HostIP(ip))
}

// FNetHostPort returns a field with the given value under [NetHostPortKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FNetHostPort(port string) zap.Field {
	return zap.String(NetHostPortKey, port)
}
