// Package gotsrpcconv provides typed OpenTelemetry metric instruments for
// gotsrpc semantic conventions.
//
// [NewExecutionDuration] creates the "gotsrpc.execution.duration" histogram,
// recorded in seconds and labeled with the gotsrpc package, service and
// function. A nil meter yields a no-op instrument.
package gotsrpcconv
