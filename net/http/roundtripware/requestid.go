package roundtripware

import (
	"net/http"
	"strings"

	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	keelhttpcontext "github.com/foomo/keel/net/http/context"
	"github.com/foomo/keel/net/http/provider"
)

type (
	// RequestIDOptions configures the [RequestID] RoundTripware.
	RequestIDOptions struct {
		// Header is the name of the request header to set.
		Header string
		// Provider generates a request ID when the context holds none.
		Provider provider.RequestID
		// SetHeader is not used.
		SetHeader bool
	}
	// RequestIDOption configures [RequestIDOptions].
	RequestIDOption func(*RequestIDOptions)
)

// GetDefaultRequestIDOptions returns the default options using the
// X-Request-ID header and [provider.DefaultRequestID].
func GetDefaultRequestIDOptions() RequestIDOptions {
	return RequestIDOptions{
		Header:   "X-Request-ID",
		Provider: provider.DefaultRequestID,
	}
}

// RequestIDWithHeader sets the request header name. Defaults to X-Request-ID.
func RequestIDWithHeader(v string) RequestIDOption {
	return func(o *RequestIDOptions) {
		o.Header = v
	}
}

// RequestIDWithProvider sets the request ID provider. Defaults to
// [provider.DefaultRequestID].
func RequestIDWithProvider(v provider.RequestID) RequestIDOption {
	return func(o *RequestIDOptions) {
		o.Provider = v
	}
}

// RequestID returns a RoundTripware that sets a request ID header, unless
// already present, using the ID from the request context or a generated one.
func RequestID(opts ...RequestIDOption) RoundTripware {
	o := GetDefaultRequestIDOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}

	return func(l *zap.Logger, next Handler) Handler {
		return func(r *http.Request) (*http.Response, error) {
			span := trace.SpanFromContext(r.Context())
			if span.IsRecording() {
				span.AddEvent("RequestID")
			}

			if value := r.Header.Get(o.Header); value == "" {
				var requestID string
				if value, ok := keelhttpcontext.GetRequestID(r.Context()); ok && value != "" {
					requestID = value
				}

				if requestID == "" {
					requestID = o.Provider()
				}

				if requestID != "" {
					if span.IsRecording() {
						span.SetAttributes(semconv.HTTPRequestHeader(strings.ToLower(o.Header), requestID))
					}

					r.Header.Set(o.Header, requestID)
				}
			}

			return next(r)
		}
	}
}
