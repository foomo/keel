package roundtripware

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Deprecated: use keelhttp.WithTelemetry instead.
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
			span.AddEvent("Metric")

			ctx, labeler := LabelerFromContext(r.Context())

			start := time.Now()
			resp, err := next(r.WithContext(ctx))
			duration := time.Since(start)

			if err != nil {
				return resp, err
			}

			attributes := append(labeler.Get(), attribute.String("method", r.Method))

			if resp != nil {
				attributes = append(labeler.Get(), attribute.Int("status_code", resp.StatusCode))
			}

			histogram.Record(ctx, duration.Seconds(), metric.WithAttributes(attributes...))

			return resp, err
		}
	}
}
