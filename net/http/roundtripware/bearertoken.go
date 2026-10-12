package roundtripware

import (
	"net/http"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type (
	// BearerTokenOptions configures the [BearerToken] RoundTripware.
	BearerTokenOptions struct {
		Header string
		Scheme string
	}
	// BearerTokenOption configures [BearerTokenOptions].
	BearerTokenOption func(*BearerTokenOptions)
)

// GetDefaultBearerTokenOptions uses Authorization with the Bearer scheme.
func GetDefaultBearerTokenOptions() BearerTokenOptions {
	return BearerTokenOptions{Header: "Authorization", Scheme: "Bearer "}
}

// BearerTokenWithHeader sets the request header name.
func BearerTokenWithHeader(v string) BearerTokenOption {
	return func(o *BearerTokenOptions) { o.Header = v }
}

// BearerTokenWithScheme sets the token prefix, including any trailing space.
// An empty scheme writes the token without a prefix.
func BearerTokenWithScheme(v string) BearerTokenOption {
	return func(o *BearerTokenOptions) { o.Scheme = v }
}

// BearerToken authenticates outbound requests with a static token, unless the
// header is already set. An empty token leaves the request unchanged.
func BearerToken(token string, opts ...BearerTokenOption) RoundTripware {
	o := GetDefaultBearerTokenOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}

	return func(_ *zap.Logger, next Handler) Handler {
		return func(r *http.Request) (*http.Response, error) {
			// Never attach the credential to logs or span attributes.
			if span := trace.SpanFromContext(r.Context()); span.IsRecording() {
				span.AddEvent("BearerToken")
			}

			if token != "" && r.Header.Get(o.Header) == "" {
				r.Header.Set(o.Header, o.Scheme+token)
			}

			return next(r)
		}
	}
}
