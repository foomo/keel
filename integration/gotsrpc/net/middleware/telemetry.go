package keelgotsrpcmiddleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/foomo/gotsrpc/v2"
	httplog "github.com/foomo/keel/net/http/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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
		Exemplars     bool
		Execution     bool
		Marshalling   bool
		Unmarshalling bool
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
		Exemplars:     false,
		Execution:     true,
		Marshalling:   false,
		Unmarshalling: false,
	}
}

// TelemetryWithExemplars middleware option
func TelemetryWithExemplars(v bool) TelemetryOption {
	return func(o *TelemetryOptions) {
		o.Exemplars = v
	}
}

// TelemetryWithExecution middleware option
func TelemetryWithExecution(v bool) TelemetryOption {
	return func(o *TelemetryOptions) {
		o.Execution = v
	}
}

// TelemetryWithMarshalling middleware option
func TelemetryWithMarshalling(v bool) TelemetryOption {
	return func(o *TelemetryOptions) {
		o.Marshalling = v
	}
}

// TelemetryWithUnmarshalling middleware option
func TelemetryWithUnmarshalling(v bool) TelemetryOption {
	return func(o *TelemetryOptions) {
		o.Unmarshalling = v
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
	observe := func(l *zap.Logger, r *http.Request, observer prometheus.ObserverVec, stats *gotsrpc.CallStats, operation string) {
		observer.WithLabelValues(
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
		spanCtx := trace.SpanContextFromContext(r.Context())
		if v, ok := observer.(prometheus.ExemplarObserver); ok && opts.Exemplars && spanCtx.HasTraceID() {
			v.ObserveWithExemplar(duration.Seconds(), prometheus.Labels{
				"traceID": spanCtx.TraceID().String(),
			})
		} else if v, ok2 := observer.(prometheus.Observer); ok2 {
			if !ok {
				l.Info("==> not ok")
			}
			if !opts.Exemplars {
				l.Info("==> not exemplar")
			}
			if !spanCtx.HasSpanID() {
				l.Info("==> not traceID")
			}
			if !spanCtx.IsSampled() {
				l.Info("==> not sampled")
			}
			v.Observe(duration.Seconds())
		}
	}
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*r = *gotsrpc.RequestWithStatsContext(r)

			next.ServeHTTP(w, r)

			if stats, ok := gotsrpc.GetStatsForRequest(r); ok {
				// create custom metics
				if opts.Marshalling {
					observe(l, r, gotsrpcRequestDurationSummary, stats, "marshalling")
					observe(l, r, gotsrpcRequestDurationHistogram, stats, "marshalling")
				}
				if opts.Unmarshalling {
					observe(l, r, gotsrpcRequestDurationSummary, stats, "unmarshalling")
					observe(l, r, gotsrpcRequestDurationHistogram, stats, "unmarshalling")
				}
				if opts.Execution {
					observe(l, r, gotsrpcRequestDurationSummary, stats, "execution")
					observe(l, r, gotsrpcRequestDurationHistogram, stats, "execution")
				}

				// enrich logger
				if labeler, ok := httplog.LabelerFromRequest(r); ok {
					labeler.Add(
						zap.String(defaultGOTSRPCFunctionLabel, stats.Func),
						zap.String(defaultGOTSRPCServiceLabel, stats.Service),
						zap.String(defaultGOTSRPCPackageLabel, stats.Package),
					)
					if stats.ErrorType != "" {
						labeler.Add(zap.String(defaultGOTSRPCErrorType, stats.ErrorType))
					}
					if stats.ErrorCode != 0 {
						labeler.Add(zap.Int(defaultGOTSRPCErrorCode, stats.ErrorCode))
						if stats.ErrorMessage != "" {
							labeler.Add(zap.String(defaultGOTSRPCErrorMessage, stats.ErrorMessage))
						}
					}
				}
			}
		})
	}
}
