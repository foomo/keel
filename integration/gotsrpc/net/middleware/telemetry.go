package keelgotsrpcmiddleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/foomo/gotsrpc/v2"
	"github.com/foomo/keel/env"
	"github.com/foomo/keel/log"
	keelhttp "github.com/foomo/keel/net/http"
	httplog "github.com/foomo/keel/net/http/log"
	keelsemconv "github.com/foomo/keel/semconv"
	"github.com/foomo/keel/semconv/gotsrpcconv"
	"github.com/foomo/keel/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.uber.org/zap"
)

type (
	// TelemetryOptions configures the [Telemetry] middleware.
	TelemetryOptions struct {
		meter            metric.Meter
		bucketBoundaries []float64
		// PayloadAttributeDisabled reports whether the request payload is
		// omitted from the span attributes.
		PayloadAttributeDisabled bool
	}
	// TelemetryOption modifies [TelemetryOptions].
	TelemetryOption func(*TelemetryOptions)
)

// DefaultTelemetryOptions returns the default options: the keel meter,
// histogram buckets from 5ms to 10s and the payload attribute disabled unless
// OTEL_GOTSRPC_PAYLOAD_ATTRIBUTE_DISABLED is false.
func DefaultTelemetryOptions() TelemetryOptions {
	return TelemetryOptions{
		meter:                    telemetry.Meter(),
		bucketBoundaries:         []float64{0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10},
		PayloadAttributeDisabled: env.GetBool("OTEL_GOTSRPC_PAYLOAD_ATTRIBUTE_DISABLED", true),
	}
}

// TelemetryWithExemplarsDisabled has no effect.
//
// Deprecated: Exemplars are not configurable.
func TelemetryWithExemplarsDisabled(v bool) TelemetryOption {
	return func(o *TelemetryOptions) {
	}
}

// TelemetryWithObserveExecution has no effect.
//
// Deprecated: Execution is always observed.
func TelemetryWithObserveExecution(v bool) TelemetryOption {
	return func(o *TelemetryOptions) {
	}
}

// TelemetryWithObserveMarshalling has no effect.
//
// Deprecated: Marshalling is not observed as a separate metric.
func TelemetryWithObserveMarshalling(v bool) TelemetryOption {
	return func(o *TelemetryOptions) {
	}
}

// TelemetryWithObserveUnmarshalling has no effect.
//
// Deprecated: Unmarshalling is not observed as a separate metric.
func TelemetryWithObserveUnmarshalling(v bool) TelemetryOption {
	return func(o *TelemetryOptions) {
	}
}

// TelemetryWithBucketBoundries sets the explicit bucket boundaries, in
// seconds, of the execution duration histogram.
func TelemetryWithBucketBoundries(v []float64) TelemetryOption {
	return func(o *TelemetryOptions) {
		o.bucketBoundaries = v
	}
}

// TelemetryWithPayloadAttributeDisabled sets whether the request payload is
// omitted from the span attributes. Defaults to true.
func TelemetryWithPayloadAttributeDisabled(v bool) TelemetryOption {
	return func(o *TelemetryOptions) {
		o.PayloadAttributeDisabled = v
	}
}

// Telemetry returns a middleware configured by [DefaultTelemetryOptions] and
// opts; see [TelemetryWithOptions]. Nil options are skipped.
//
// Deprecated: Use gotsrpc v3, which includes OpenTelemetry instrumentation.
func Telemetry(opts ...TelemetryOption) keelhttp.Middleware {
	options := DefaultTelemetryOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return TelemetryWithOptions(options)
}

// TelemetryWithOptions returns a middleware that, after a gotsrpc call has
// been handled, renames the active span to "<service>/<func>", adds gotsrpc
// attributes and error details to the span and the request log labeler, and
// records the execution duration.
//
// Deprecated: Use gotsrpc v3, which includes OpenTelemetry instrumentation.
func TelemetryWithOptions(opts TelemetryOptions) keelhttp.Middleware {
	m, err := gotsrpcconv.NewExecutionDuration(
		opts.meter,
		metric.WithExplicitBucketBoundaries(opts.bucketBoundaries...),
	)
	if err != nil {
		otel.Handle(err)
	}

	sanitizePayload := func(r *http.Request) string {
		if r.Method != http.MethodPost {
			return ""
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			return ""
		}

		if err := r.Body.Close(); err != nil {
			return ""
		}

		r.Body = io.NopCloser(bytes.NewBuffer(body))

		var out bytes.Buffer
		if err = json.Indent(&out, body, "", "  "); err != nil {
			return ""
		}

		return out.String()
	}

	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*r = *gotsrpc.RequestWithStatsContext(r)
			sp := telemetry.SpanFromContext(r.Context())

			sp.AddEvent("GOTSRPC Telemetry")

			next.ServeHTTP(w, r)

			if stats, ok := gotsrpc.GetStatsForRequest(r); ok {
				if !opts.PayloadAttributeDisabled {
					sp.SetAttributes(keelsemconv.GoTSRPCPayload(sanitizePayload(r)))
				}

				// override span name
				sp.SetName(fmt.Sprintf("%s/%s", stats.Service, stats.Func))

				// define default attributes
				attrs := []attribute.KeyValue{
					semconv.RPCSystemNameKey.String("gotsrpc"),
					keelsemconv.GoTSRPCFunc(stats.Func),
					keelsemconv.GoTSRPCService(stats.Service),
					keelsemconv.GoTSRPCPackage(stats.Package),
				}

				// add trace attributes
				sp.SetAttributes(append(attrs,
					keelsemconv.GoTSRPCMarshalling(stats.Marshalling.Milliseconds()),
					keelsemconv.GoTSRPCUnmarshalling(stats.Unmarshalling.Milliseconds()),
				)...)

				if stats.ErrorCode != 0 {
					sp.SetStatus(codes.Error, stats.ErrorMessage)
					sp.SetAttributes(keelsemconv.GoTSRPCErrorCode(stats.ErrorCode))
				}

				if stats.ErrorType != "" {
					sp.SetAttributes(keelsemconv.GoTSRPCErrorType(stats.ErrorType))
				}

				if stats.ErrorMessage != "" {
					sp.SetAttributes(keelsemconv.GoTSRPCErrorMessage(stats.ErrorMessage))
				}

				m.Record(r.Context(),
					stats.Execution.Seconds(),
					stats.Package,
					stats.Service,
					stats.Func,
					m.AttrError(stats.ErrorCode != 0),
				)

				// enrich logger
				if labeler, ok := httplog.LabelerFromRequest(r); ok {
					labeler.Add(log.Attributes(attrs...)...)

					if stats.ErrorCode != 0 {
						labeler.Add(log.Attribute(keelsemconv.GoTSRPCErrorCode(stats.ErrorCode)))
					}

					if stats.ErrorType != "" {
						labeler.Add(log.Attribute(keelsemconv.GoTSRPCErrorType(stats.ErrorType)))
					}

					if stats.ErrorMessage != "" {
						labeler.Add(log.Attribute(keelsemconv.GoTSRPCErrorMessage(stats.ErrorMessage)))
					}
				}
			}
		})
	}
}
