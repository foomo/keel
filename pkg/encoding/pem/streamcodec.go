package pem

import (
	stdpem "encoding/pem"
	"errors"
	"io"
)

// StreamCodec is a StreamCodec[*pem.Block] backed by encoding/pem.
// It is safe for concurrent use.
type StreamCodec struct{}

// NewStreamCodec returns a PEM stream serializer.
func NewStreamCodec() *StreamCodec { return &StreamCodec{} }

func (StreamCodec) Encode(w io.Writer, v *stdpem.Block) error {
	return stdpem.Encode(w, v)
}

func (StreamCodec) Decode(r io.Reader, v **stdpem.Block) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	block, _ := stdpem.Decode(data)
	if block == nil {
		return errors.New("encoding: no PEM block found")
	}

	*v = block

	return nil
}
