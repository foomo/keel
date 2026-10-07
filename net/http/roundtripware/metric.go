package roundtripware

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Metric returns a RoundTripware that records the request duration in
// seconds in a Float64Histogram named name. Failed requests are not recorded.
// It panics if the histogram cannot be created.
//
// Deprecated: Use keelhttp.WithTelemetry instead.
func Metric(meter metric.Meter, name, description string) RoundTripware {
	histogram, err := meter.Float64Histogram(
		name,
		metric.WithDescription(description),
	)
	if err != nil {
		panic(err)
	}

	return func(l *zap.Logger, next Handler) Handler {
		return func(r *http.Request) (*http.Response, error) {
			span := trace.SpanFromContext(r.Context())
			if span.IsRecording() {
				span.AddEvent("Metric")
			}

			ctx, labeler := LabelerFromContext(r.Context())

			start := time.Now()
			resp, err := next(r.WithContext(ctx))
			duration := time.Since(start)

			if err != nil {
				return resp, err
			}

			attributes := append(labeler.Get(), semconv.HTTPRequestMethodKey.String(r.Method))

			if resp != nil {
				attributes = append(attributes, semconv.HTTPResponseStatusCode(resp.StatusCode))
			}

			histogram.Record(ctx, duration.Seconds(), metric.WithAttributes(attributes...))

			return resp, err
		}
	}
}
