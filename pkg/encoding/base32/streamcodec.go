package base32

import (
	stdbase32 "encoding/base32"
	"io"
)

// StreamCodec is a StreamCodec[[]byte] backed by encoding/base32.
// It is safe for concurrent use.
type StreamCodec struct{}

// NewStreamCodec returns a Base32 stream serializer.
func NewStreamCodec() *StreamCodec { return &StreamCodec{} }

func (StreamCodec) Encode(w io.Writer, v []byte) error {
	enc := stdbase32.NewEncoder(stdbase32.StdEncoding, w)
	if _, err := enc.Write(v); err != nil {
		return err
	}

	return enc.Close()
}

func (StreamCodec) Decode(r io.Reader, v *[]byte) error {
	data, err := io.ReadAll(stdbase32.NewDecoder(stdbase32.StdEncoding, r))
	if err != nil {
		return err
	}

	*v = data

	return nil
}
