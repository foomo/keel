package encoding_test

import (
	"bytes"
	"encoding/asn1"
	stdpem "encoding/pem"
	"fmt"

	encascii85 "github.com/foomo/keel/pkg/encoding/ascii85"
	encasn1 "github.com/foomo/keel/pkg/encoding/asn1"
	"github.com/foomo/keel/pkg/encoding/base32"
	"github.com/foomo/keel/pkg/encoding/base64"
	"github.com/foomo/keel/pkg/encoding/csv"
	"github.com/foomo/keel/pkg/encoding/gob"
	"github.com/foomo/keel/pkg/encoding/hex"
	"github.com/foomo/keel/pkg/encoding/json"
	"github.com/foomo/keel/pkg/encoding/pem"
	"github.com/foomo/keel/pkg/encoding/xml"
)

func ExampleStreamCodec_json() {
	type Data struct {
		Name string
	}

	c := json.NewStreamCodec[Data]()

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

func ExampleStreamCodec_gob() {
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

func ExampleStreamCodec_xml() {
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

func ExampleStreamCodec_csv() {
	c := csv.NewStreamCodec()

	records := [][]string{
		{"name", "age"},
		{"Alice", "30"},
	}

	var buf bytes.Buffer
	if err := c.Encode(&buf, records); err != nil {
		fmt.Printf("Encode failed: %v\n", err)
		return
	}

	var decoded [][]string
	if err := c.Decode(&buf, &decoded); err != nil {
		fmt.Printf("Decode failed: %v\n", err)
		return
	}

	fmt.Printf("Decoded: %v\n", decoded)
	// Output:
	// Decoded: [[name age] [Alice 30]]
}

func ExampleStreamCodec_base64() {
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

func ExampleStreamCodec_base32() {
	c := base32.NewStreamCodec()

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
	// Encoded: NBSWY3DP
	// Decoded: hello
}

func ExampleStreamCodec_hex() {
	c := hex.NewStreamCodec()

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
	// Encoded: 68656c6c6f
	// Decoded: hello
}

func ExampleStreamCodec_ascii85() {
	c := encascii85.NewStreamCodec()

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

func ExampleStreamCodec_pem() {
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

func ExampleStreamCodec_asn1() {
	c := encasn1.NewStreamCodec[asn1.RawValue]()

	original := asn1.RawValue{
		Tag:   asn1.TagUTF8String,
		Class: asn1.ClassUniversal,
		Bytes: []byte("hello"),
	}

	var buf bytes.Buffer
	if err := c.Encode(&buf, original); err != nil {
		fmt.Printf("Encode failed: %v\n", err)
		return
	}

	var decoded asn1.RawValue
	if err := c.Decode(&buf, &decoded); err != nil {
		fmt.Printf("Decode failed: %v\n", err)
		return
	}

	fmt.Printf("Decoded: %s\n", string(decoded.Bytes))
	// Output:
	// Decoded: hello
}
