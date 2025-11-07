package roundtripware

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Dump returns a RoundTripper which prints out the request & response object
func Dump() RoundTripware {
	return func(l *zap.Logger, next Handler) Handler {
		return func(r *http.Request) (*http.Response, error) {
			dumpRequest(r)
			resp, err := next(r)
			dumpResponse(r, resp)

			return resp, err
		}
	}
}

// DumpRequest returns a RoundTripper which prints out the request object
func DumpRequest() RoundTripware {
	return func(l *zap.Logger, next Handler) Handler {
		return func(r *http.Request) (*http.Response, error) {
			span := trace.SpanFromContext(r.Context())
			if span.IsRecording() {
				span.AddEvent("DumpRequest")
			}

			dumpRequest(r)

			return next(r)
		}
	}
}

// DumpResponse returns a RoundTripper which prints out the response object
func DumpResponse() RoundTripware {
	return func(l *zap.Logger, next Handler) Handler {
		return func(r *http.Request) (*http.Response, error) {
			span := trace.SpanFromContext(r.Context())
			if span.IsRecording() {
				span.AddEvent("DumpResponse")
			}

			resp, err := next(r)
			dumpResponse(r, resp)

			return resp, err
		}
	}
}

func dumpRequest(req *http.Request) {
	if req.Header != nil && req.Header.Get("Content-Type") != "" {
		var body string
		if req.Body, body = readBodyPretty(req.Header.Get("Content-Type"), req.Body); body != "" {
			fmt.Printf("Request %s:\n%s\n", req.URL, body) //nolint:forbidigo
		}
	}
}

func dumpResponse(req *http.Request, resp *http.Response) {
	if resp == nil {
		fmt.Println("Response is nil") //nolint:forbidigo
		return
	}

	if resp.Header != nil && resp.Header.Get("Content-Type") != "" {
		var body string
		if resp.Body, body = readBodyPretty(resp.Header.Get("Content-Type"), resp.Body); body != "" {
			fmt.Printf("Response %s:\n%s\n", req.URL, body) //nolint:forbidigo
		}
	}
}
