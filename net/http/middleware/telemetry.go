package middleware

import (
	"fmt"
	"net/http"

	"github.com/foomo/keel/log"
	httplog "github.com/foomo/keel/net/http/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type (
	TelemetryOptions struct {
		Name                    string
		OtelOpts                []otelhttp.Option
		InjectPropagationHeader bool
	}
	TelemetryOption func(*TelemetryOptions)
)

// GetDefaultTelemetryOptions returns the default options
func GetDefaultTelemetryOptions() TelemetryOptions {
	return TelemetryOptions{
		OtelOpts: []otelhttp.Option{
			otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
				return fmt.Sprintf("HTTP %s", operation)
			}),
		},
		InjectPropagationHeader: true,
	}
}

// Telemetry middleware
func Telemetry(opts ...TelemetryOption) Middleware {
	options := GetDefaultTelemetryOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	return TelemetryWithOptions(options)
}

func TelemetryWithName(v string) TelemetryOption {
	return func(o *TelemetryOptions) {
		o.Name = v
	}
}

func TelemetryWithInjectPropagationHeader(v bool) TelemetryOption {
	return func(o *TelemetryOptions) {
		o.InjectPropagationHeader = v
	}
}

// TelemetryWithOtelOpts middleware options
func TelemetryWithOtelOpts(v ...otelhttp.Option) TelemetryOption {
	return func(o *TelemetryOptions) {
		o.OtelOpts = append(o.OtelOpts, v...)
	}
}

// TelemetryWithOptions middleware
func TelemetryWithOptions(opts TelemetryOptions) Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		if opts.Name != "" {
			name = opts.Name
		}

		return otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if opts.InjectPropagationHeader {
				otel.GetTextMapPropagator().Inject(r.Context(), propagation.HeaderCarrier(w.Header()))
			}

			if labeler, ok := httplog.LabelerFromRequest(r); ok {
				if spanCtx := trace.SpanContextFromContext(r.Context()); spanCtx.IsValid() && spanCtx.IsSampled() {
					labeler.Add(log.FTraceID(spanCtx.TraceID().String()))
					labeler.Add(log.FSpanID(spanCtx.SpanID().String()))
				}
			}

			// wrap response write to get access to status & size
			wr := WrapResponseWriter(w)

			next.ServeHTTP(wr, r)
		}), name, opts.OtelOpts...)
	}
}
