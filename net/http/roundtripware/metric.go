package roundtripware

import (
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

			ctx, labeler := LabelerFromContext(req.Context())

			start := time.Now()
			resp, err := next(req.WithContext(ctx))
			duration := time.Since(start)
			if err != nil {
				return resp, err
			}

			attributes := append(labeler.Get(), attribute.String("method", req.Method))

			if resp != nil {
				attributes = append(labeler.Get(), attribute.Int("status_code", resp.StatusCode))
			}

			histogram.Record(ctx, duration.Seconds(), attributes...)

			return resp, err
		}
	}
}
