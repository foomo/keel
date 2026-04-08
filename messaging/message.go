package messaging

// Message is the unit passed to every Handler. Subject carries the routing
// key (e.g. a NATS subject or Redis channel); Payload holds the decoded value.
type Message[T any] struct {
	Subject string `json:"subject"`
	Payload T      `json:"payload"`
}

// NewMessage creates a new Message.
func NewMessage[T any](subject string, payload T) Message[T] {
	return Message[T]{
		Subject: subject,
		Payload: payload,
	}
}
