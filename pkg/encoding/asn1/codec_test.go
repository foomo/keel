package asn1_test

import (
	"fmt"

	"github.com/foomo/keel/pkg/encoding/asn1"
)

func ExampleCodec() {
	c := asn1.NewCodec[int]()

	encoded, err := c.Encode(42)
	if err != nil {
		fmt.Printf("Encode failed: %v\n", err)
		return
	}

	var decoded int
	if err := c.Decode(encoded, &decoded); err != nil {
		fmt.Printf("Decode failed: %v\n", err)
		return
	}

	fmt.Printf("Decoded: %d\n", decoded)
	// Output:
	// Decoded: 42
}
