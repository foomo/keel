package roundtripware

import (
	"net/http"

	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	keelhttpcontext "github.com/foomo/keel/net/http/context"
)

type (
	// SessionIDOptions configures the [SessionID] RoundTripware.
	SessionIDOptions struct {
		// Header is the name of the request header to set.
		Header string
	}
	// SessionIDOption configures [SessionIDOptions].
	SessionIDOption func(*SessionIDOptions)
	// SessionIDGenerator returns a new session ID. It is not used by this
	// package.
	SessionIDGenerator func() string
)

// GetDefaultSessionIDOptions returns the default options using the
// X-Session-ID header.
func GetDefaultSessionIDOptions() SessionIDOptions {
	return SessionIDOptions{
		Header: "X-Session-ID",
	}
}

// SessionIDWithHeader sets the request header name. Defaults to X-Session-ID.
func SessionIDWithHeader(v string) SessionIDOption {
	return func(o *SessionIDOptions) {
		o.Header = v
	}
}

// SessionID returns a RoundTripware that sets the session ID stored in the
// request context as request header, unless the header is already present.
func SessionID(opts ...SessionIDOption) RoundTripware {
	o := GetDefaultSessionIDOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}

	return func(l *zap.Logger, next Handler) Handler {
		return func(r *http.Request) (*http.Response, error) {
			span := trace.SpanFromContext(r.Context())
			if span.IsRecording() {
				span.AddEvent("SessionID")
			}

			if value := r.Header.Get(o.Header); value == "" {
				if value, ok := keelhttpcontext.GetSessionID(r.Context()); ok && value != "" {
					if span.IsRecording() {
						span.SetAttributes(semconv.SessionID(value))
					}

					r.Header.Set(o.Header, value)
				}
			}

			return next(r)
		}
	}
}
