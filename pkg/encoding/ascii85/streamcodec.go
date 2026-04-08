package ascii85

import (
	stdascii85 "encoding/ascii85"
	"io"
)

// StreamCodec is a StreamCodec[[]byte] backed by encoding/ascii85.
// It is safe for concurrent use.
type StreamCodec struct{}

// NewStreamCodec returns an ASCII85 stream serializer.
func NewStreamCodec() StreamCodec { return StreamCodec{} }

func (StreamCodec) Encode(w io.Writer, v []byte) error {
	dst := make([]byte, stdascii85.MaxEncodedLen(len(v)))
	n := stdascii85.Encode(dst, v)

	_, err := w.Write(dst[:n])

	return err
}

func (StreamCodec) Decode(r io.Reader, v *[]byte) error {
	data, err := io.ReadAll(stdascii85.NewDecoder(r))
	if err != nil {
		return err
	}

	*v = data

	return nil
}
