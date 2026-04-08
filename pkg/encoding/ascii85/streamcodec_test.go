package ascii85_test

import (
	"bytes"
	"fmt"

	"github.com/foomo/keel/pkg/encoding/ascii85"
)

func ExampleNewStreamCodec() {
	c := ascii85.NewStreamCodec()

	var buf bytes.Buffer
	if err := c.Encode(&buf, []byte("hello")); err != nil {
		fmt.Printf("Encode failed: %v\n", err)
		return
	}

	fmt.Printf("Encoded: %s\n", buf.String())

	var decoded []byte
	if err := c.Decode(bytes.NewReader(buf.Bytes()), &decoded); err != nil {
		fmt.Printf("Decode failed: %v\n", err)
		return
	}

	fmt.Printf("Decoded: %s\n", string(decoded))
	// Output:
	// Encoded: BOu!rDZ
	// Decoded: hello
}
