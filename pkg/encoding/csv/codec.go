package csv

import (
	"bytes"
	stdcsv "encoding/csv"

	"github.com/foomo/keel/internal/sync"
)

// Codec is a Codec[[][]string] backed by encoding/csv.
// It is safe for concurrent use.
type Codec struct{}

// NewCodec returns a CSV serializer.
func NewCodec() *Codec { return &Codec{} }

func (Codec) Encode(v [][]string) ([]byte, error) {
	buf := sync.Get()
	defer sync.Put(buf)

	cw := stdcsv.NewWriter(buf)
	if err := cw.WriteAll(v); err != nil {
		return nil, err
	}

	cw.Flush()

	if err := cw.Error(); err != nil {
		return nil, err
	}

	out := make([]byte, buf.Len())
	copy(out, buf.Bytes())

	return out, nil
}

func (Codec) Decode(b []byte, v *[][]string) error {
	records, err := stdcsv.NewReader(bytes.NewReader(b)).ReadAll()
	if err != nil {
		return err
	}

	*v = records

	return nil
}
