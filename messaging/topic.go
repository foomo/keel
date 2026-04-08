package messaging

// Topic bundles a Publisher and Subscriber sharing the same Codec. Use it
// when a service needs to both produce and consume the same message type.
type Topic[T any] struct {
	Publisher[T]
	Subscriber[T]
}
