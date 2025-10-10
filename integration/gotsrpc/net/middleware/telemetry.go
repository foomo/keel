package keelgotsrpcmiddleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/foomo/gotsrpc/v2"
	"github.com/foomo/keel/env"
	"github.com/foomo/keel/log"
	httplog "github.com/foomo/keel/net/http/log"
	"github.com/foomo/keel/net/http/middleware"
	keelsemconv "github.com/foomo/keel/semconv"
	"github.com/foomo/keel/semconv/gotsrpcconv"
	"github.com/foomo/keel/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

type (
	TelemetryOptions struct {
		meter                    metric.Meter
		bucketBoundaries         []float64
		PayloadAttributeDisabled bool
	}
	TelemetryOption func(*TelemetryOptions)
)

// DefaultTelemetryOptions returns the default options
func DefaultTelemetryOptions() TelemetryOptions {
	return TelemetryOptions{
		meter:                    telemetry.Meter(),
		bucketBoundaries:         []float64{0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10},
		PayloadAttributeDisabled: env.GetBool("OTEL_GOTSRPC_PAYLOAD_ATTRIBUTE_DISABLED", true),
	}
}

// Deprecated: TelemetryWithExemplarsDisabled middleware option
func TelemetryWithExemplarsDisabled(v bool) TelemetryOption {
	return func(o *TelemetryOptions) {
	}
}

// Deprecated: TelemetryWithObserveExecution middleware option
func TelemetryWithObserveExecution(v bool) TelemetryOption {
	return func(o *TelemetryOptions) {
	}
}

// Deprecated: TelemetryWithObserveMarshalling middleware option
func TelemetryWithObserveMarshalling(v bool) TelemetryOption {
	return func(o *TelemetryOptions) {
	}
}

// Deprecated: TelemetryWithObserveUnmarshalling middleware option
func TelemetryWithObserveUnmarshalling(v bool) TelemetryOption {
	return func(o *TelemetryOptions) {
	}
}

// TelemetryWithBucketBoundries middleware option
func TelemetryWithBucketBoundries(v []float64) TelemetryOption {
	return func(o *TelemetryOptions) {
		o.bucketBoundaries = v
	}
}

// TelemetryWithPayloadAttributeDisabled middleware option
func TelemetryWithPayloadAttributeDisabled(v bool) TelemetryOption {
	return func(o *TelemetryOptions) {
		o.PayloadAttributeDisabled = v
	}
}

// Telemetry middleware
func Telemetry(opts ...TelemetryOption) middleware.Middleware {
	options := DefaultTelemetryOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return TelemetryWithOptions(options)
}

// TelemetryWithOptions middleware
func TelemetryWithOptions(opts TelemetryOptions) middleware.Middleware {
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

			ctx := telemetry.Ctx(r.Context())
			ctx.AddSpanEvent("GOTSRCP Telemetry")

			ctx.StartProfile(func(ctx telemetry.Context) {
				r = r.WithContext(ctx)

				next.ServeHTTP(w, r)

				if stats, ok := gotsrpc.GetStatsForRequest(r); ok {
					if !opts.PayloadAttributeDisabled {
						ctx.SetSpanAttributes(keelsemconv.GoTSRPCPayload(sanitizePayload(r)))
					}

					var pkg string
					if parts := strings.Split(stats.Package, "/"); len(parts) > 0 {
						pkg = parts[len(parts)-1] + "."
					}

					ctx.SetSpanName(fmt.Sprintf("GOTSRPC %s%s/%s", pkg, stats.Service, stats.Func))
					attrs := []attribute.KeyValue{
						keelsemconv.GoTSRPCFunc(stats.Func),
						keelsemconv.GoTSRPCService(stats.Service),
						keelsemconv.GoTSRPCPackage(stats.Package),
					}
					ctx.SetSpanAttributes(append(attrs,
						keelsemconv.GoTSRPCMarshalling(stats.Marshalling.Milliseconds()),
						keelsemconv.GoTSRPCUnmarshalling(stats.Unmarshalling.Milliseconds()),
					)...)
					ctx.SetProfileAttributes(attrs...)

					if stats.ErrorCode != 0 {
						ctx.SetSpanStatusError(stats.ErrorMessage)
						ctx.SetSpanAttributes(keelsemconv.GoTSRPCErrorCode(stats.ErrorCode))
					}

					if stats.ErrorType != "" {
						ctx.SetSpanAttributes(keelsemconv.GoTSRPCErrorType(stats.ErrorType))
					}

					if stats.ErrorMessage != "" {
						ctx.SetSpanAttributes(keelsemconv.GoTSRPCErrorMessage(stats.ErrorMessage))
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
		})
	}
}
