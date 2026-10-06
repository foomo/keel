package semconv

import (
	"go.opentelemetry.io/otel/attribute"
)

const (
	// CircuitBreakerPreviousStateKey is the key for circuit_breaker.previous_state.
	CircuitBreakerPreviousStateKey = attribute.Key("circuit_breaker.previous_state")
	// CircuitBreakerStateChangeKey is the key for circuit_breaker.state_change.
	CircuitBreakerStateChangeKey = attribute.Key("circuit_breaker.state_change")
)

// KeelCircuitBreakerPreviousState returns a new attribute.KeyValue for circuit_breaker.previous_state.
func KeelCircuitBreakerPreviousState(v string) attribute.KeyValue {
	return CircuitBreakerPreviousStateKey.String(v)
}

// KeelCircuitBreakerStateChange returns a new attribute.KeyValue for circuit_breaker.state_change.
func KeelCircuitBreakerStateChange(v bool) attribute.KeyValue {
	return CircuitBreakerStateChangeKey.Bool(v)
}
