package nats

import (
	semconvmsg "go.opentelemetry.io/otel/semconv/v1.40.0/messagingconv"
)

// system is the messaging.system attribute value for this transport.
var system = semconvmsg.SystemAttr("nats")
