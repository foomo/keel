package pem_test

import (
	"bytes"
	stdpem "encoding/pem"
	"fmt"

	"github.com/foomo/keel/pkg/encoding/pem"
)

func ExampleStreamCodec() {
	c := pem.NewStreamCodec()

	block := &stdpem.Block{
		Type:  "TEST",
		Bytes: []byte("hello"),
	}

	var buf bytes.Buffer
	if err := c.Encode(&buf, block); err != nil {
		fmt.Printf("Encode failed: %v\n", err)
		return
	}

	var decoded *stdpem.Block
	if err := c.Decode(&buf, &decoded); err != nil {
		fmt.Printf("Decode failed: %v\n", err)
		return
	}

	fmt.Printf("Decoded Type: %s\n", decoded.Type)
	fmt.Printf("Decoded Bytes: %s\n", string(decoded.Bytes))
	// Output:
	// Decoded Type: TEST
	// Decoded Bytes: hello
}
