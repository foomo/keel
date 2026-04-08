package yaml_test

import (
	"fmt"

	"github.com/foomo/keel/pkg/encoding/yaml/v3"
)

func ExampleCodec() {
	type Data struct {
		Name string
	}

	c := yaml.NewCodec[Data]()

	encoded, err := c.Encode(Data{Name: "example-123"})
	if err != nil {
		fmt.Printf("Encode failed: %v\n", err)
		return
	}

	fmt.Printf("Encoded: %s\n", string(encoded))

	var decoded Data
	if err := c.Decode(encoded, &decoded); err != nil {
		fmt.Printf("Decode failed: %v\n", err)
		return
	}

	fmt.Printf("Decoded Name: %s\n", decoded.Name)
	// Output:
	// Encoded: name: example-123
	//
	// Decoded Name: example-123
}
