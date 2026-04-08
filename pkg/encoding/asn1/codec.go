package asn1

import (
	stdasn1 "encoding/asn1"
)

// Codec is a Codec[T] backed by encoding/asn1.
// It is safe for concurrent use.
type Codec[T any] struct{}

// NewCodec returns an ASN1 serializer for T.
func NewCodec[T any]() *Codec[T] { return &Codec[T]{} }

func (Codec[T]) Encode(v T) ([]byte, error) {
	return stdasn1.Marshal(v)
}

func (Codec[T]) Decode(b []byte, v *T) error {
	_, err := stdasn1.Unmarshal(b, v)

	return err
}
