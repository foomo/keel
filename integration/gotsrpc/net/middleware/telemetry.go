package keelgotsrpcmiddleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/foomo/gotsrpc/v2"
	httplog "github.com/foomo/keel/net/http/log"
	"github.com/foomo/keel/telemetry"
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
		Exemplars                bool
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
		Buckets:     []float64{0.05, 0.1, 0.5, 1, 3, 5, 10},
	}, []string{
		defaultGOTSRPCFunctionLabel,
		defaultGOTSRPCServiceLabel,
		defaultGOTSRPCPackageLabel,
		defaultGOTSRPCPackageOperation,
		defaultGOTSRPCError,
	})
	// Deprecated: use gotsrpc_process_duration_seconds
	gotsrpcRequestDurationSummary = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:  "gotsrpc",
		Name:       "request_duration_seconds",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		Help:       "Specifies the duration of gotsrpc request in seconds summary",
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
		Exemplars:                false,
		ObserveExecution:         true,
		ObserveMarshalling:       false,
		ObserveUnmarshalling:     false,
		PayloadAttributeDisabled: true,
	}
}

// TelemetryWithExemplars middleware option
func TelemetryWithExemplars(v bool) TelemetryOption {
	return func(o *TelemetryOptions) {
		o.Exemplars = v
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
		if exemplarObserver, ok := observer.(prometheus.ExemplarObserver); ok && opts.Exemplars && spanCtx.HasTraceID() && spanCtx.IsSampled() {
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
			ctx, span := telemetry.Start(r.Context(), "GOTSRPC")
			*r = *gotsrpc.RequestWithStatsContext(r.WithContext(ctx))

			next.ServeHTTP(w, r)

			if stats, ok := gotsrpc.GetStatsForRequest(r); ok {
				span.SetName(fmt.Sprintf("GOTSRPC %s.%s", stats.Service, stats.Func))
				span.SetAttributes(
					attribute.String("gotsrpc.func", stats.Func),
					attribute.String("gotsrpc.service", stats.Service),
					attribute.String("gotsrpc.package", stats.Package),
					attribute.Int64("gotsrpc.marshalling", stats.Marshalling.Milliseconds()),
					attribute.Int64("gotsrpc.unmarshalling", stats.Unmarshalling.Milliseconds()),
				)
				if !opts.PayloadAttributeDisabled {
					span.SetAttributes(attribute.String("gotsprc.payload", sanitizePayload(r)))
				}
				if stats.ErrorCode != 0 {
					span.SetStatus(codes.Error, fmt.Sprintf("%s: %s", stats.ErrorType, stats.ErrorMessage))
					span.SetAttributes(attribute.Int("gotsrpc.error.code", stats.ErrorCode))
				}
				if stats.ErrorType != "" {
					span.SetAttributes(attribute.String("gotsrpc.error.type", stats.ErrorType))
				}
				if stats.ErrorMessage != "" {
					span.SetAttributes(attribute.String("gotsrpc.error.message", stats.ErrorMessage))
				}

				// create custom metics
				if opts.ObserveMarshalling {
					observe(span.SpanContext(), gotsrpcRequestDurationSummary, stats, "marshalling")
					observe(span.SpanContext(), gotsrpcRequestDurationHistogram, stats, "marshalling")
				}
				if opts.ObserveUnmarshalling {
					observe(span.SpanContext(), gotsrpcRequestDurationSummary, stats, "unmarshalling")
					observe(span.SpanContext(), gotsrpcRequestDurationHistogram, stats, "unmarshalling")
				}
				if opts.ObserveExecution {
					observe(span.SpanContext(), gotsrpcRequestDurationSummary, stats, "execution")
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
			telemetry.End(span, nil)
		})
	}
}
