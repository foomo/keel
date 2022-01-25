package roundtripware

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"

	"github.com/tinylib/msgp/msgp"
)

func readBodyPretty(contentType string, original io.ReadCloser) (io.ReadCloser, string) {
	var bs bytes.Buffer
	var body string
	defer func() {
		_ = original.Close()
	}()

	// read in body
	if _, err := io.Copy(&bs, original); err != nil {
		return original, ""
	} else {
		body = bs.String()
	}

	switch {
	case strings.HasPrefix(contentType, "application/json"):
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, bs.Bytes(), "", "  "); err == nil {
			body = prettyJSON.String()
		}
	case strings.HasPrefix(contentType, "application/msgpack"):
		var prettyJSON bytes.Buffer
		var out bytes.Buffer
		if _, err := msgp.UnmarshalAsJSON(&out, bs.Bytes()); err == nil {
			if err := json.Indent(&prettyJSON, out.Bytes(), "", "  "); err == nil {
				body = prettyJSON.String()
			}
		}
	}

	// return copy of the original
	return ioutil.NopCloser(strings.NewReader(bs.String())), body
}
