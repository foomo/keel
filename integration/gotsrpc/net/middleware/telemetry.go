package keelgotsrpcmiddleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/foomo/gotsrpc/v2"
	"github.com/foomo/keel/env"
	httplog "github.com/foomo/keel/net/http/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/foomo/keel/net/http/middleware"
)

// Prometheus Metrics
const (
	defaultGOTSRPCFunctionLabel    = "gotsrpc_func"
	defaultGOTSRPCServiceLabel     = "gotsrpc_service"
	defaultGOTSRPCPackageLabel     = "gotsrpc_package"
	defaultGOTSRPCPackageOperation = "gotsrpc_operation"
	defaultGOTSRPCError            = "gotsrpc_error"
	defaultGOTSRPCErrorCode        = "gotsrpc_error_code"
	defaultGOTSRPCErrorType        = "gotsrpc_error_type"
	defaultGOTSRPCErrorMessage     = "gotsrpc_error_message"
)

type (
	TelemetryOptions struct {
		ExemplarsDisabled        bool
		ObserveExecution         bool
		ObserveMarshalling       bool
		ObserveUnmarshalling     bool
		PayloadAttributeDisabled bool
	}
	TelemetryOption func(*TelemetryOptions)
)

var (
	gotsrpcRequestDurationHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   "gotsrpc",
		Name:        "process_duration_seconds",
		Help:        "Specifies the duration of the gotsrpc process in seconds",
		ConstLabels: nil,
		Buckets:     []float64{0.01, 0.05, 0.1, 0.5, 1, 5, 10},
	}, []string{
		defaultGOTSRPCFunctionLabel,
		defaultGOTSRPCServiceLabel,
		defaultGOTSRPCPackageLabel,
		defaultGOTSRPCPackageOperation,
		defaultGOTSRPCError,
	})
)

// DefaultTelemetryOptions returns the default options
func DefaultTelemetryOptions() TelemetryOptions {
	return TelemetryOptions{
		ObserveExecution:         true,
		ObserveMarshalling:       false,
		ObserveUnmarshalling:     false,
		PayloadAttributeDisabled: env.GetBool("OTEL_GOTSRPC_PAYLOAD_ATTRIBUTE_DISABLED", true),
		ExemplarsDisabled:        env.GetBool("OTEL_GOTSRPC_EXEMPLARS_DISABLED", false),
	}
}

// TelemetryWithExemplarsDisabled middleware option
func TelemetryWithExemplarsDisabled(v bool) TelemetryOption {
	return func(o *TelemetryOptions) {
		o.ExemplarsDisabled = v
	}
}

// TelemetryWithObserveExecution middleware option
func TelemetryWithObserveExecution(v bool) TelemetryOption {
	return func(o *TelemetryOptions) {
		o.ObserveExecution = v
	}
}

// TelemetryWithObserveMarshalling middleware option
func TelemetryWithObserveMarshalling(v bool) TelemetryOption {
	return func(o *TelemetryOptions) {
		o.ObserveMarshalling = v
	}
}

// TelemetryWithObserveUnmarshalling middleware option
func TelemetryWithObserveUnmarshalling(v bool) TelemetryOption {
	return func(o *TelemetryOptions) {
		o.ObserveUnmarshalling = v
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
	observe := func(spanCtx trace.SpanContext, metric prometheus.ObserverVec, stats *gotsrpc.CallStats, operation string) {
		observer := metric.WithLabelValues(
			stats.Func,
			stats.Service,
			stats.Package,
			operation,
			strconv.FormatBool(stats.ErrorCode != 0),
		)

		var duration time.Duration

		switch operation {
		case "marshalling":
			duration = stats.Marshalling
		case "unmarshalling":
			duration = stats.Unmarshalling
		case "execution":
			duration = stats.Execution
		}

		if exemplarObserver, ok := observer.(prometheus.ExemplarObserver); ok && opts.ExemplarsDisabled && spanCtx.HasTraceID() && spanCtx.IsSampled() {
			exemplarObserver.ObserveWithExemplar(duration.Seconds(), prometheus.Labels{
				"traceID": spanCtx.TraceID().String(),
			})

			return
		}

		observer.Observe(duration.Seconds())
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
					span.SetAttributes(attribute.String("gotsprc.payload", sanitizePayload(r)))
				}

				var pkg string
				if parts := strings.Split(stats.Package, "/"); len(parts) > 0 {
					pkg = parts[len(parts)-1] + "."
				}

				span.SetName(fmt.Sprintf("GOTSRPC %s%s/%s", pkg, stats.Service, stats.Func))
				span.SetAttributes(
					attribute.String("gotsrpc.func", stats.Func),
					attribute.String("gotsrpc.service", stats.Service),
					attribute.String("gotsrpc.package", stats.Package),
					attribute.Int64("gotsrpc.marshalling", stats.Marshalling.Milliseconds()),
					attribute.Int64("gotsrpc.unmarshalling", stats.Unmarshalling.Milliseconds()),
				)

				if stats.ErrorCode != 0 {
					span.SetStatus(codes.Error, fmt.Sprintf("%s: %s", stats.ErrorType, stats.ErrorMessage))
					span.SetAttributes(attribute.Int("gotsrpc.error.code", stats.ErrorCode))
				} else {
					span.SetStatus(codes.Ok, "")
				}

				if stats.ErrorType != "" {
					span.SetAttributes(attribute.String("gotsrpc.error.type", stats.ErrorType))
				}

				if stats.ErrorMessage != "" {
					span.SetAttributes(attribute.String("gotsrpc.error.message", stats.ErrorMessage))
				}

				// create custom metics
				if opts.ObserveMarshalling {
					observe(span.SpanContext(), gotsrpcRequestDurationHistogram, stats, "marshalling")
				}

				if opts.ObserveUnmarshalling {
					observe(span.SpanContext(), gotsrpcRequestDurationHistogram, stats, "unmarshalling")
				}

				if opts.ObserveExecution {
					observe(span.SpanContext(), gotsrpcRequestDurationHistogram, stats, "execution")
				}

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
