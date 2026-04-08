package gob

import (
	"bytes"
	stdgob "encoding/gob"

	"github.com/foomo/keel/internal/sync"
)

// Codec is a Codec[T] backed by encoding/gob.
// It is safe for concurrent use.
type Codec[T any] struct{}

// NewCodec returns a GOB serializer for T.
func NewCodec[T any]() *Codec[T] { return &Codec[T]{} }

func (Codec[T]) Encode(v T) ([]byte, error) {
	buf := sync.Get()
	defer sync.Put(buf)

	if err := stdgob.NewEncoder(buf).Encode(v); err != nil {
		return nil, err
	}

	out := make([]byte, buf.Len())
	copy(out, buf.Bytes())

	return out, nil
}

func (Codec[T]) Decode(b []byte, v *T) error {
	return stdgob.NewDecoder(bytes.NewReader(b)).Decode(v)
}
