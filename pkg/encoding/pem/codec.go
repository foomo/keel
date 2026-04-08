package pem

import (
	stdpem "encoding/pem"
	"errors"
)

// Codec is a Codec[*pem.Block] backed by encoding/pem.
// It is safe for concurrent use.
type Codec struct{}

// NewCodec returns a PEM serializer.
func NewCodec() *Codec { return &Codec{} }

func (Codec) Encode(v *stdpem.Block) ([]byte, error) {
	return stdpem.EncodeToMemory(v), nil
}

func (Codec) Decode(b []byte, v **stdpem.Block) error {
	block, _ := stdpem.Decode(b)
	if block == nil {
		return errors.New("encoding: no PEM block found")
	}

	*v = block

	return nil
}
