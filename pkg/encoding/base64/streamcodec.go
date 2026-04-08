package base64

import (
	stdbase64 "encoding/base64"
	"io"
)

// StreamCodec is a StreamCodec[[]byte] backed by encoding/base64.
// It is safe for concurrent use.
type StreamCodec struct{}

// NewStreamCodec returns a Base64 stream serializer.
func NewStreamCodec() StreamCodec { return StreamCodec{} }

func (StreamCodec) Encode(w io.Writer, v []byte) error {
	enc := stdbase64.NewEncoder(stdbase64.StdEncoding, w)
	if _, err := enc.Write(v); err != nil {
		_ = enc.Close()
		return err
	}

	return enc.Close()
}

func (StreamCodec) Decode(r io.Reader, v *[]byte) error {
	data, err := io.ReadAll(stdbase64.NewDecoder(stdbase64.StdEncoding, r))
	if err != nil {
		return err
	}

	*v = data

	return nil
}
