package gob_test

import (
	"fmt"

	"github.com/foomo/keel/pkg/encoding/gob"
)

func ExampleCodec() {
	type Data struct {
		Name string
	}

	c := gob.NewCodec[Data]()

	encoded, err := c.Encode(Data{Name: "example-123"})
	if err != nil {
		fmt.Printf("Encode failed: %v\n", err)
		return
	}

	var decoded Data
	if err := c.Decode(encoded, &decoded); err != nil {
		fmt.Printf("Decode failed: %v\n", err)
		return
	}

	fmt.Printf("Decoded Name: %s\n", decoded.Name)
	// Output:
	// Decoded Name: example-123
}
