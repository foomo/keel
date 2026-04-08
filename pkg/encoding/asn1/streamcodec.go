package asn1

import (
	stdasn1 "encoding/asn1"
	"io"
)

// StreamCodec is a StreamCodec[T] backed by encoding/asn1.
// It is safe for concurrent use.
type StreamCodec[T any] struct{}

// NewStreamCodec returns an ASN1 stream serializer for T.
func NewStreamCodec[T any]() *StreamCodec[T] { return &StreamCodec[T]{} }

func (StreamCodec[T]) Encode(w io.Writer, v T) error {
	data, err := stdasn1.Marshal(v)
	if err != nil {
		return err
	}

	_, err = w.Write(data)

	return err
}

func (StreamCodec[T]) Decode(r io.Reader, v *T) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	_, err = stdasn1.Unmarshal(data, v)

	return err
}
