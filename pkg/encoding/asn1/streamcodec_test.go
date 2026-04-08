package asn1_test

import (
	"bytes"
	"fmt"

	"github.com/foomo/keel/pkg/encoding/asn1"
)

func ExampleStreamCodec() {
	c := asn1.NewStreamCodec[int]()

	var buf bytes.Buffer
	if err := c.Encode(&buf, 42); err != nil {
		fmt.Printf("Encode failed: %v\n", err)
		return
	}

	var decoded int
	if err := c.Decode(&buf, &decoded); err != nil {
		fmt.Printf("Decode failed: %v\n", err)
		return
	}

	fmt.Printf("Decoded: %d\n", decoded)
	// Output:
	// Decoded: 42
}
