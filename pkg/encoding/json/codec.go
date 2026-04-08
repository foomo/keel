package json

import (
	"encoding/json"
)

// Codec is a Codec[T] backed by encoding/json.
// It is safe for concurrent use.
type Codec[T any] struct{}

// NewCodec returns a JSON serializer for T.
func NewCodec[T any]() Codec[T] { return Codec[T]{} }

func (Codec[T]) Encode(v T) ([]byte, error) {
	return json.Marshal(v)
}

func (Codec[T]) Decode(b []byte, v *T) error {
	return json.Unmarshal(b, v)
}
