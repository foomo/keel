package semconv

import (
	"go.opentelemetry.io/otel/attribute"
)

const (
	// HTTPXRequestIDKey is the key for http.request.id.
	HTTPXRequestIDKey = attribute.Key("http.request.id")
	// HTTPXRequestRefererKey is the key for http.request.referer.
	HTTPXRequestRefererKey = attribute.Key("http.request.referer")
)

// HTTPXRequestID returns a new attribute.KeyValue for http.request.id.
func HTTPXRequestID(v string) attribute.KeyValue {
	return HTTPXRequestIDKey.String(v)
}

// HTTPXRequestReferer returns a new attribute.KeyValue for http.request.referer.
func HTTPXRequestReferer(v string) attribute.KeyValue {
	return HTTPXRequestRefererKey.String(v)
}
