package csv

import (
	stdcsv "encoding/csv"
	"io"
)

// StreamCodec is a StreamCodec[[][]string] backed by encoding/csv.
// It is safe for concurrent use.
type StreamCodec struct{}

// NewStreamCodec returns a CSV stream serializer.
func NewStreamCodec() *StreamCodec { return &StreamCodec{} }

func (StreamCodec) Encode(w io.Writer, v [][]string) error {
	cw := stdcsv.NewWriter(w)
	if err := cw.WriteAll(v); err != nil {
		return err
	}

	cw.Flush()

	return cw.Error()
}

func (StreamCodec) Decode(r io.Reader, v *[][]string) error {
	records, err := stdcsv.NewReader(r).ReadAll()
	if err != nil {
		return err
	}

	*v = records

	return nil
}
