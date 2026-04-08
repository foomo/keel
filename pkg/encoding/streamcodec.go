package encoding

import (
	"io"
)

// StreamCodec encodes T to an io.Writer and decodes T from an io.Reader.
type StreamCodec[T any] interface {
	Encode(w io.Writer, v T) error
	Decode(r io.Reader, v *T) error
}

type Encoder[T any] interface {
	Encode(v T) error
}

type Decoder[T any] interface {
	Decode(v any) error
}
