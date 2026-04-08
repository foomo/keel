package hex

import (
	stdhex "encoding/hex"
)

// Codec is a Codec[[]byte] backed by encoding/hex.
// It is safe for concurrent use.
type Codec struct{}

// NewCodec returns a Hex serializer.
func NewCodec() *Codec { return &Codec{} }

func (Codec) Encode(v []byte) ([]byte, error) {
	dst := make([]byte, stdhex.EncodedLen(len(v)))
	stdhex.Encode(dst, v)

	return dst, nil
}

func (Codec) Decode(b []byte, v *[]byte) error {
	dst := make([]byte, stdhex.DecodedLen(len(b)))

	n, err := stdhex.Decode(dst, b)
	if err != nil {
		return err
	}

	*v = dst[:n]

	return nil
}
