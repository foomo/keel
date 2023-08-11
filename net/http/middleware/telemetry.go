package middleware

import (
	"net/http"

	"github.com/foomo/keel/log"
	httplog "github.com/foomo/keel/net/http/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
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
		OtelOpts:                []otelhttp.Option{},
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
		o.OtelOpts = v
	}
}

// TelemetryWithOptions middleware
func TelemetryWithOptions(opts TelemetryOptions) Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		if opts.Name != "" {
			name = opts.Name
		}
		// TODO remove once https://github.com/open-telemetry/opentelemetry-go-contrib/pull/771 is merged
		m := otel.Meter(
			"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp",
			metric.WithInstrumentationVersion(otelhttp.Version()),
		)
		c, err := m.Int64Counter(
			"foo_"+otelhttp.RequestCount,
			metric.WithDescription("counts number of requests withs specific status code"),
		)
		if err != nil {
			otel.Handle(err)
		}

		return otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if opts.InjectPropagationHeader {
				otel.GetTextMapPropagator().Inject(r.Context(), propagation.HeaderCarrier(w.Header()))
			}

			if labeler, ok := httplog.LabelerFromRequest(r); ok {
				if spanCtx := trace.SpanContextFromContext(r.Context()); spanCtx.IsValid() {
					labeler.Add(log.FTraceID(spanCtx.TraceID().String()))
				}
			}

			// wrap response write to get access to status & size
			wr := WrapResponseWriter(w)

			next.ServeHTTP(wr, r)

			if labeler, ok := otelhttp.LabelerFromContext(r.Context()); ok {
				c.Add(r.Context(), 1, metric.WithAttributes(append(labeler.Get(), semconv.HTTPStatusCodeKey.Int(wr.StatusCode()))...))
			}
		}), name, opts.OtelOpts...)
	}
}
