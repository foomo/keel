package roundtripware

import (
	"context"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.uber.org/zap"
)

// Metric returns a RoundTripper which prints out the request & response object
func Metric(meter metric.Meter, name, description string) RoundTripware {

	histogram, err := meter.SyncFloat64().Histogram(
		name,
		instrument.WithDescription(description),
	)
	if err != nil {
		panic(err)
	}

	return func(l *zap.Logger, next Handler) Handler {
		return func(req *http.Request) (*http.Response, error) {

			name, hasName := GetRequestName(req)
			if !hasName {
				name = "http.Do"
			}

			start := time.Now()
			resp, err := next(req)
			duration := time.Since(start)

			status := "unknown response"
			if resp != nil {
				status = resp.Status
			}

			histogram.Record(req.Context(), duration.Seconds(),
				// semconv.HTTPClientAttributesFromHTTPRequest(req)[],
				attribute.String("name", name),
				attribute.String("method", req.Method),
				attribute.String("status", status),
			)

			return resp, err
		}
	}
}

func WithNamedRequestContext(r *http.Request, name string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), contextKeyHttpClientRequestName, name))
}

func GetRequestName(r *http.Request) (string, bool) {
	if value := r.Context().Value(contextKeyHttpClientRequestName); value != nil {
		return value.(string), true
	} else {
		return "", false
	}
}

func ClearRequestName(r *http.Request) {
	*r = *r.WithContext(context.WithValue(r.Context(), contextKeyHttpClientRequestName, nil))
}
