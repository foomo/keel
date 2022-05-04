package keelgotsrpcmiddleware

import (
	"net/http"
	"time"

	"github.com/foomo/gotsrpc/v2"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	"github.com/foomo/keel/net/http/middleware"
)

// Telemetry middleware
func Telemetry() middleware.Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*r = *gotsrpc.RequestWithStatsContext(r)
			next.ServeHTTP(w, r)
			if labeler, ok := otelhttp.LabelerFromContext(r.Context()); ok {
				if stats, ok := gotsrpc.GetStatsForRequest(r); ok {
					// Use floating point division here for higher precision (instead of Millisecond method).
					labeler.Add(attribute.String("gotsrpc_func", stats.Func))
					labeler.Add(attribute.String("gotsrpc_service", stats.Service))
					labeler.Add(attribute.String("gotsrpc_package", stats.Package))
					labeler.Add(attribute.Float64("gotsrpc_execution", float64(stats.Execution)/float64(time.Millisecond)))
					labeler.Add(attribute.Float64("gotsrpc_marshalling", float64(stats.Marshalling)/float64(time.Millisecond)))
					labeler.Add(attribute.Float64("gotsrpc_unmarshalling", float64(stats.Unmarshalling)/float64(time.Millisecond)))
				}
			}
		})
	}
}
