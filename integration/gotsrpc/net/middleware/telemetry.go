package keelgotsrpcmiddleware

import (
	"net/http"
	"time"

	"github.com/foomo/gotsrpc/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	"github.com/foomo/keel/net/http/middleware"
)

//Prometheus Metrics
const (
	defaultGOTSRPCFunctionLabel    = "gotsrpc_func"
	defaultGOTSRPCServiceLabel     = "gotsrpc_service"
	defaultGOTSRPCPackageLabel     = "gotsrpc_package"
	defaultGOTSRPCPackageOperation = "gotsrpc_operation"
)

var (
	gotsrpcRequestDurationSummary = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name: "gotsrpc_request_duration_seconds",
		Help: "Specifies the duration of gotsrpc request in seconds",
	}, []string{defaultGOTSRPCFunctionLabel, defaultGOTSRPCServiceLabel, defaultGOTSRPCPackageLabel, defaultGOTSRPCPackageOperation})
)

// Telemetry middleware
func Telemetry() middleware.Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*r = *gotsrpc.RequestWithStatsContext(r)
			next.ServeHTTP(w, r)
			if labeler, ok := otelhttp.LabelerFromContext(r.Context()); ok {
				if stats, ok := gotsrpc.GetStatsForRequest(r); ok {
					labeler.Add(attribute.String(defaultGOTSRPCFunctionLabel, stats.Func))
					labeler.Add(attribute.String(defaultGOTSRPCServiceLabel, stats.Service))
					labeler.Add(attribute.String(defaultGOTSRPCPackageLabel, stats.Package))

					observe := func(operation string, duration time.Duration) {
						gotsrpcRequestDurationSummary.WithLabelValues(
							stats.Func,
							stats.Service,
							stats.Package,
							operation,
						).Observe(duration.Seconds())
					}

					observe("marshalling", stats.Marshalling)
					observe("unmarshalling", stats.Unmarshalling)
					observe("execution", stats.Execution)
				}
			}
		})
	}
}
