package roundtripware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/tinylib/msgp/msgp"
)

func readBodyPretty(contentType string, original io.ReadCloser) (io.ReadCloser, string) {
	var (
		bs   bytes.Buffer
		body string
	)

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
		var (
			prettyJSON bytes.Buffer
			out        bytes.Buffer
		)

		if _, err := msgp.UnmarshalAsJSON(&out, bs.Bytes()); err == nil {
			if err := json.Indent(&prettyJSON, out.Bytes(), "", "  "); err == nil {
				body = prettyJSON.String()
			}
		}
	}

	// return copy of the original
	return io.NopCloser(strings.NewReader(bs.String())), body
}

// errNoBody is a sentinel error value used by failureToReadBody so we
// can detect that the lack of body was intentional.
var errNoBody = errors.New("sentinel error value")

// failureToReadBody is an io.ReadCloser that just returns errNoBody on
// Read. It's swapped in when we don't actually want to consume
// the body, but need a non-nil one, and want to distinguish the
// error from reading the dummy body.
type failureToReadBody struct{}

func (failureToReadBody) Read([]byte) (int, error) { return 0, errNoBody }
func (failureToReadBody) Close() error             { return nil }

// emptyBody is an instance of empty reader.
var emptyBody = io.NopCloser(strings.NewReader(""))

func copyRequest(req *http.Request, body bool) (*http.Request, error) {
	// we don't care about the context, since it is only used for the isSuccessful check
	out := req.Clone(context.Background())

	// duplicate the body
	if !body {
		// For content length of zero. Make sure the body is an empty
		// reader, instead of returning error through failureToReadBody{}.
		if req.ContentLength == 0 {
			out.Body = emptyBody
		} else {
			// if it is attempted to read from the body in isSuccessful we actually want the read to fail
			out.Body = failureToReadBody{}
		}
	} else if req.Body == nil {
		req.Body = nil
		out.Body = nil
	} else {
		var err error

		out.Body, req.Body, err = drainBody(req.Body)
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

func copyResponse(resp *http.Response, body bool) (*http.Response, error) {
	// we don't care about the context, since it is only used for the isSuccessful check
	out := new(http.Response)
	*out = *resp

	// duplicate the body
	if !body {
		// For content length of zero. Make sure the body is an empty
		// reader, instead of returning error through failureToReadBody{}.
		if resp.ContentLength == 0 {
			out.Body = emptyBody
		} else {
			// if it is attempted to read from the body in isSuccessful we actually want the read to fail
			out.Body = failureToReadBody{}
		}
	} else if resp.Body == nil {
		out.Body = nil
	} else {
		var err error

		out.Body, resp.Body, err = drainBody(resp.Body)
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

// copied from httputil
func drainBody(b io.ReadCloser) (io.ReadCloser, io.ReadCloser, error) {
	var err error

	if b == nil || b == http.NoBody {
		// No copying needed. Preserve the magic sentinel meaning of NoBody.
		return http.NoBody, http.NoBody, nil
	}

	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, b, err
	}

	if err = b.Close(); err != nil {
		return nil, b, err
	}

	return io.NopCloser(&buf), io.NopCloser(bytes.NewReader(buf.Bytes())), nil
}
