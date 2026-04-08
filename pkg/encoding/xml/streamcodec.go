package xml

import (
	"encoding/xml"
	"io"
)

// StreamCodec is a StreamCodec[T] backed by encoding/xml.
// It is safe for concurrent use.
type StreamCodec[T any] struct{}

// NewStreamCodec returns an XML stream serializer for T.
func NewStreamCodec[T any]() *StreamCodec[T] { return &StreamCodec[T]{} }

func (StreamCodec[T]) Encode(w io.Writer, v T) error {
	return xml.NewEncoder(w).Encode(v)
}

func (StreamCodec[T]) Decode(r io.Reader, v *T) error {
	return xml.NewDecoder(r).Decode(v)
}
