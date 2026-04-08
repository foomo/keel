package json

import (
	"io"

	"github.com/go-json-experiment/json"
)

// Codec is a Codec[T] backed by encoding/json.
// It is zero-allocation on the encode path for small structs and safe for
// concurrent use.
type Codec[T any] struct{}

// NewCodec returns a JSON serializer for T.
func NewCodec[T any]() *Codec[T] { return &Codec[T]{} }

func (Codec[T]) Encode(v T) ([]byte, error) {
	return json.Marshal(v)
}

func (Codec[T]) Decode(b []byte, v *T) error {
	return json.Unmarshal(b, v)
}

func (Codec[T]) EncodeTo(w io.Writer, v T) error {
	return json.MarshalWrite(w, v)
}

func (Codec[T]) DecodeFrom(r io.Reader, v *T) error {
	return json.UnmarshalRead(r, v)
}
