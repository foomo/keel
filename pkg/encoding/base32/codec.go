package base32

import (
	stdbase32 "encoding/base32"
)

// Codec is a Codec[[]byte] backed by encoding/base32.
// It is safe for concurrent use.
type Codec struct{}

// NewCodec returns a Base32 serializer.
func NewCodec() *Codec { return &Codec{} }

func (Codec) Encode(v []byte) ([]byte, error) {
	dst := make([]byte, stdbase32.StdEncoding.EncodedLen(len(v)))
	stdbase32.StdEncoding.Encode(dst, v)

	return dst, nil
}

func (Codec) Decode(b []byte, v *[]byte) error {
	dst := make([]byte, stdbase32.StdEncoding.DecodedLen(len(b)))

	n, err := stdbase32.StdEncoding.Decode(dst, b)
	if err != nil {
		return err
	}

	*v = dst[:n]

	return nil
}
