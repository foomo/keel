package roundtripware

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// Dump returns a RoundTripper which prints out the request & response object
func Dump() RoundTripware {
	return func(l *zap.Logger, next Handler) Handler {
		return func(req *http.Request) (*http.Response, error) {
			dumpRequest(req)
			resp, err := next(req)
			dumpResponse(req, resp)
			return resp, err
		}
	}
}

// DumpRequest returns a RoundTripper which prints out the request object
func DumpRequest() RoundTripware {
	return func(l *zap.Logger, next Handler) Handler {
		return func(req *http.Request) (*http.Response, error) {
			dumpRequest(req)
			return next(req)
		}
	}
}

// DumpResponse returns a RoundTripper which prints out the response object
func DumpResponse() RoundTripware {
	return func(l *zap.Logger, next Handler) Handler {
		return func(req *http.Request) (*http.Response, error) {
			resp, err := next(req)
			dumpResponse(req, resp)
			return resp, err
		}
	}
}

func dumpRequest(req *http.Request) {
	if req.Header != nil && req.Header.Get("Content-Type") != "" {
		var body string
		if req.Body, body = readBodyPretty(req.Header.Get("Content-Type"), req.Body); body != "" {
			fmt.Printf("Request %s:\n%s\n", req.URL, body)
		}
	}
}

func dumpResponse(req *http.Request, resp *http.Response) {
	if resp.Header != nil && resp.Header.Get("Content-Type") != "" {
		var body string
		if resp.Body, body = readBodyPretty(resp.Header.Get("Content-Type"), resp.Body); body != "" {
			fmt.Printf("Response %s:\n%s\n", req.URL, body)
		}
	}
}
