package pem_test

import (
	stdpem "encoding/pem"
	"fmt"

	"github.com/foomo/keel/pkg/encoding/pem"
)

func ExampleCodec() {
	c := pem.NewCodec()

	block := &stdpem.Block{
		Type:  "TEST",
		Bytes: []byte("hello"),
	}

	encoded, err := c.Encode(block)
	if err != nil {
		fmt.Printf("Encode failed: %v\n", err)
		return
	}

	fmt.Printf("Encoded:\n%s", string(encoded))

	var decoded *stdpem.Block
	if err := c.Decode(encoded, &decoded); err != nil {
		fmt.Printf("Decode failed: %v\n", err)
		return
	}

	fmt.Printf("Decoded Type: %s\n", decoded.Type)
	fmt.Printf("Decoded Bytes: %s\n", string(decoded.Bytes))
	// Output:
	// Encoded:
	// -----BEGIN TEST-----
	// aGVsbG8=
	// -----END TEST-----
	// Decoded Type: TEST
	// Decoded Bytes: hello
}
