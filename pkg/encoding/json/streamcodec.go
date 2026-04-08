package json

import (
	"encoding/json"
	"io"
)

// StreamCodec is a StreamCodec[T] backed by encoding/json.
// It is safe for concurrent use.
type StreamCodec[T any] struct{}

// NewStreamCodec returns a JSON stream serializer for T.
func NewStreamCodec[T any]() *StreamCodec[T] { return &StreamCodec[T]{} }

func (StreamCodec[T]) Encode(w io.Writer, v T) error {
	return json.NewEncoder(w).Encode(v)
}

func (StreamCodec[T]) Decode(r io.Reader, v *T) error {
	return json.NewDecoder(r).Decode(v)
}
