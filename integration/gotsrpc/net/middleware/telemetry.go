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
	httplog "github.com/foomo/keel/net/http/log"
	"github.com/foomo/keel/net/http/middleware"
	keelsemconv "github.com/foomo/keel/semconv"
	"github.com/foomo/keel/semconv/gotsrpcconv"
	"github.com/foomo/keel/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Prometheus Metrics
const (
	defaultGOTSRPCFunctionLabel = "gotsrpc_func"
	defaultGOTSRPCServiceLabel  = "gotsrpc_service"
	defaultGOTSRPCPackageLabel  = "gotsrpc_package"
	defaultGOTSRPCErrorCode     = "gotsrpc_error_code"
	defaultGOTSRPCErrorType     = "gotsrpc_error_type"
	defaultGOTSRPCErrorMessage  = "gotsrpc_error_message"
)

type (
	TelemetryOptions struct {
		meter                    metric.Meter
		bucketBoundries          []float64
		PayloadAttributeDisabled bool
	}
	TelemetryOption func(*TelemetryOptions)
)

// DefaultTelemetryOptions returns the default options
func DefaultTelemetryOptions() TelemetryOptions {
	return TelemetryOptions{
		meter:                    telemetry.Meter(),
		bucketBoundries:          []float64{0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10},
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
		o.bucketBoundries = v
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
		metric.WithExplicitBucketBoundaries(opts.bucketBoundries...),
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
			span := trace.SpanFromContext(r.Context())
			if span.IsRecording() {
				span.AddEvent("GOTSRCP Telemetry")
			}

			*r = *gotsrpc.RequestWithStatsContext(r)

			next.ServeHTTP(w, r)

			if stats, ok := gotsrpc.GetStatsForRequest(r); ok {
				if !opts.PayloadAttributeDisabled {
					span.SetAttributes(keelsemconv.GoTSRPCPayload(sanitizePayload(r)))
				}

				var pkg string
				if parts := strings.Split(stats.Package, "/"); len(parts) > 0 {
					pkg = parts[len(parts)-1] + "."
				}

				span.SetName(fmt.Sprintf("GOTSRPC %s%s/%s", pkg, stats.Service, stats.Func))
				span.SetAttributes(
					keelsemconv.GoTSRPCFunc(stats.Func),
					keelsemconv.GoTSRPCService(stats.Service),
					keelsemconv.GoTSRPCPackage(stats.Package),
					keelsemconv.GoTSRPCMarshalling(stats.Marshalling.Milliseconds()),
					keelsemconv.GoTSRPCUnmarshalling(stats.Unmarshalling.Milliseconds()),
				)

				if stats.ErrorCode != 0 {
					span.SetStatus(codes.Error, stats.ErrorMessage)
					span.SetAttributes(keelsemconv.GoTSRPCErrorCode(stats.ErrorCode))
				}

				if stats.ErrorType != "" {
					span.SetAttributes(keelsemconv.GoTSRPCErrorType(stats.ErrorType))
				}

				if stats.ErrorMessage != "" {
					span.SetAttributes(keelsemconv.GoTSRPCErrorMessage(stats.ErrorMessage))
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
					labeler.Add(
						zap.String(defaultGOTSRPCFunctionLabel, stats.Func),
						zap.String(defaultGOTSRPCServiceLabel, stats.Service),
						zap.String(defaultGOTSRPCPackageLabel, stats.Package),
					)

					if stats.ErrorCode != 0 {
						labeler.Add(zap.Int(defaultGOTSRPCErrorCode, stats.ErrorCode))
					}

					if stats.ErrorType != "" {
						labeler.Add(zap.String(defaultGOTSRPCErrorType, stats.ErrorType))
					}

					if stats.ErrorMessage != "" {
						labeler.Add(zap.String(defaultGOTSRPCErrorMessage, stats.ErrorMessage))
					}
				}
			}
		})
	}
}
