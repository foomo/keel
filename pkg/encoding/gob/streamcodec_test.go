package gob_test

import (
	"bytes"
	"fmt"

	"github.com/foomo/keel/pkg/encoding/gob"
)

func ExampleStreamCodec() {
	type Data struct {
		Name string
	}

	c := gob.NewStreamCodec[Data]()

	var buf bytes.Buffer
	if err := c.Encode(&buf, Data{Name: "example-123"}); err != nil {
		fmt.Printf("Encode failed: %v\n", err)
		return
	}

	var decoded Data
	if err := c.Decode(&buf, &decoded); err != nil {
		fmt.Printf("Decode failed: %v\n", err)
		return
	}

	fmt.Printf("Decoded Name: %s\n", decoded.Name)
	// Output:
	// Decoded Name: example-123
}
