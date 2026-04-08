package hex_test

import (
	"fmt"

	"github.com/foomo/keel/pkg/encoding/hex"
)

func ExampleCodec() {
	c := hex.NewCodec()

	encoded, err := c.Encode([]byte("hello"))
	if err != nil {
		fmt.Printf("Encode failed: %v\n", err)
		return
	}

	fmt.Printf("Encoded: %s\n", string(encoded))
	// Output:
	// Encoded: 68656c6c6f
}
