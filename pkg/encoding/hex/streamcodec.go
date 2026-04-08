package hex

import (
	stdhex "encoding/hex"
	"io"
)

// StreamCodec is a StreamCodec[[]byte] backed by encoding/hex.
// It is safe for concurrent use.
type StreamCodec struct{}

// NewStreamCodec returns a Hex stream serializer.
func NewStreamCodec() *StreamCodec { return &StreamCodec{} }

func (StreamCodec) Encode(w io.Writer, v []byte) error {
	_, err := stdhex.NewEncoder(w).Write(v)

	return err
}

func (StreamCodec) Decode(r io.Reader, v *[]byte) error {
	data, err := io.ReadAll(stdhex.NewDecoder(r))
	if err != nil {
		return err
	}

	*v = data

	return nil
}
