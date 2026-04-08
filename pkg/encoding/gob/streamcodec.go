package gob

import (
	"encoding/gob"
	"io"
)

// StreamCodec is a StreamCodec[T] backed by encoding/gob.
// It is safe for concurrent use.
type StreamCodec[T any] struct{}

// NewStreamCodec returns a GOB stream serializer for T.
func NewStreamCodec[T any]() *StreamCodec[T] { return &StreamCodec[T]{} }

func (StreamCodec[T]) Encode(w io.Writer, v T) error {
	return gob.NewEncoder(w).Encode(v)
}

func (StreamCodec[T]) Decode(r io.Reader, v *T) error {
	return gob.NewDecoder(r).Decode(v)
}
