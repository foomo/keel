package base64

import (
	stdbase64 "encoding/base64"
)

// Codec is a Codec[[]byte] backed by encoding/base64.
// It is safe for concurrent use.
type Codec struct{}

// NewCodec returns a Base64 serializer.
func NewCodec() *Codec { return &Codec{} }

func (Codec) Encode(v []byte) ([]byte, error) {
	dst := make([]byte, stdbase64.StdEncoding.EncodedLen(len(v)))
	stdbase64.StdEncoding.Encode(dst, v)

	return dst, nil
}

func (Codec) Decode(b []byte, v *[]byte) error {
	dst := make([]byte, stdbase64.StdEncoding.DecodedLen(len(b)))

	n, err := stdbase64.StdEncoding.Decode(dst, b)
	if err != nil {
		return err
	}

	*v = dst[:n]

	return nil
}
