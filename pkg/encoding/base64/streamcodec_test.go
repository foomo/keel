package base64_test

import (
	"bytes"
	"fmt"

	"github.com/foomo/keel/pkg/encoding/base64"
)

func ExampleStreamCodec() {
	c := base64.NewStreamCodec()

	var buf bytes.Buffer
	if err := c.Encode(&buf, []byte("hello")); err != nil {
		fmt.Printf("Encode failed: %v\n", err)
		return
	}

	fmt.Printf("Encoded: %s\n", buf.String())

	var decoded []byte
	if err := c.Decode(&buf, &decoded); err != nil {
		fmt.Printf("Decode failed: %v\n", err)
		return
	}

	fmt.Printf("Decoded: %s\n", string(decoded))
	// Output:
	// Encoded: aGVsbG8=
	// Decoded: hello
}
