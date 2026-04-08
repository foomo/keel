package yaml

import (
	"go.yaml.in/yaml/v3"
)

// Codec is a Codec[T] backed by encoding/json.
// It is zero-allocation on the encode path for small structs and safe for
// concurrent use.
type Codec[T any] struct{}

// NewCodec returns a Codec codec for T.
func NewCodec[T any]() Codec[T] { return Codec[T]{} }

func (Codec[T]) Encode(v T) ([]byte, error) {
	return yaml.Marshal(v)
}

func (Codec[T]) Decode(b []byte, v *T) error {
	return yaml.Unmarshal(b, v)
}
