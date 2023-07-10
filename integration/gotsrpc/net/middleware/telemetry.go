package keelgotsrpcmiddleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/foomo/gotsrpc/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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
	defaultGOTSRPCErrorMessage     = "gotsrpc_error_message"
)

var (
	gotsrpcRequestDurationSummary = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "gotsrpc_request_duration_seconds",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		Help:       "Specifies the duration of gotsrpc request in seconds",
	}, []string{defaultGOTSRPCFunctionLabel, defaultGOTSRPCServiceLabel, defaultGOTSRPCPackageLabel, defaultGOTSRPCPackageOperation, defaultGOTSRPCError})
)

// Telemetry middleware
func Telemetry() middleware.Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*r = *gotsrpc.RequestWithStatsContext(r)
			next.ServeHTTP(w, r)
			if stats, ok := gotsrpc.GetStatsForRequest(r); ok {
				_, labeler := middleware.LoggerLabelerFromContext(r.Context())
				labeler.Add(
					zap.String(defaultGOTSRPCFunctionLabel, stats.Func),
					zap.String(defaultGOTSRPCServiceLabel, stats.Service),
					zap.String(defaultGOTSRPCPackageLabel, stats.Package),
				)
				if stats.ErrorCode != 0 {
					labeler.Add(zap.Int(defaultGOTSRPCErrorCode, stats.ErrorCode))
					if stats.ErrorMessage != "" {
						labeler.Add(zap.String(defaultGOTSRPCErrorMessage, stats.ErrorMessage))
					}
				}

				observe := func(operation string, duration time.Duration) {
					gotsrpcRequestDurationSummary.WithLabelValues(
						stats.Func,
						stats.Service,
						stats.Package,
						operation,
						strconv.FormatBool(stats.ErrorCode != 0),
					).Observe(duration.Seconds())
				}

				observe("marshalling", stats.Marshalling)
				observe("unmarshalling", stats.Unmarshalling)
				observe("execution", stats.Execution)
			}
		})
	}
}
