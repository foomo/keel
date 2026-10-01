// Package keelgotsrpcmiddleware provides HTTP middleware that adds gotsrpc
// call telemetry to keel HTTP services.
//
// [Telemetry] renames the active span after the called gotsrpc service and
// function, annotates it and the request logger with gotsrpc attributes and
// records the execution duration metric defined in
// [github.com/foomo/keel/semconv/gotsrpcconv]. The middleware is deprecated
// in favor of gotsrpc v3, which includes OpenTelemetry instrumentation.
package keelgotsrpcmiddleware
