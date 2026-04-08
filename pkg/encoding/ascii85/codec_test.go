package ascii85_test

import (
	"fmt"

	"github.com/foomo/keel/pkg/encoding/ascii85"
)

func ExampleNewCodec() {
	c := ascii85.NewCodec()

	encoded, err := c.Encode([]byte("hello"))
	if err != nil {
		fmt.Printf("Encode failed: %v\n", err)
		return
	}

	fmt.Printf("Encoded: %s\n", string(encoded))

	var decoded []byte
	if err := c.Decode(encoded, &decoded); err != nil {
		fmt.Printf("Decode failed: %v\n", err)
		return
	}

	fmt.Printf("Decoded: %s\n", string(decoded))
	// Output:
	// Encoded: BOu!rDZ
	// Decoded: hello
}
