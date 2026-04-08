package xml_test

import (
	"bytes"
	"fmt"

	"github.com/foomo/keel/pkg/encoding/xml"
)

func ExampleStreamCodec() {
	type Data struct {
		XMLName struct{} `xml:"data"`
		Name    string   `xml:"name"`
	}

	c := xml.NewStreamCodec[Data]()

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
