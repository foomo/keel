package csv_test

import (
	"bytes"
	"fmt"

	"github.com/foomo/keel/pkg/encoding/csv"
)

func ExampleStreamCodec() {
	c := csv.NewStreamCodec()

	records := [][]string{
		{"name", "age"},
		{"Alice", "30"},
	}

	var buf bytes.Buffer
	if err := c.Encode(&buf, records); err != nil {
		fmt.Printf("Encode failed: %v\n", err)
		return
	}

	var decoded [][]string
	if err := c.Decode(&buf, &decoded); err != nil {
		fmt.Printf("Decode failed: %v\n", err)
		return
	}

	fmt.Printf("Decoded: %v\n", decoded)
	// Output:
	// Decoded: [[name age] [Alice 30]]
}
