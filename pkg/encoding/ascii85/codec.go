package ascii85

import (
	"bytes"
	stdascii85 "encoding/ascii85"
)

// Codec is a Codec[[]byte] backed by encoding/ascii85.
// It is safe for concurrent use.
type Codec struct{}

// NewCodec returns an ASCII85 serializer.
func NewCodec() *Codec { return &Codec{} }

func (Codec) Encode(v []byte) ([]byte, error) {
	dst := make([]byte, stdascii85.MaxEncodedLen(len(v)))
	n := stdascii85.Encode(dst, v)

	return dst[:n], nil
}

func (Codec) Decode(b []byte, v *[]byte) error {
	buf := bytes.NewBuffer(make([]byte, 0, len(b)))
	r := stdascii85.NewDecoder(bytes.NewReader(b))

	if _, err := buf.ReadFrom(r); err != nil {
		return err
	}

	*v = buf.Bytes()

	return nil
}
