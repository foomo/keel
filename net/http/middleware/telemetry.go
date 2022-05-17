package middleware

import (
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
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
		m := global.MeterProvider().Meter(
			"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp",
			metric.WithInstrumentationVersion(otelhttp.SemVersion()),
		)
		c, err := m.SyncInt64().Counter(
			otelhttp.RequestCount,
			instrument.WithDescription("counts number of requests withs specific status code"),
		)
		if err != nil {
			otel.Handle(err)
		}

		serverLatency, err := m.SyncFloat64().Histogram(
			otelhttp.ServerLatency,
			instrument.WithDescription("histogram for request duration in seconds"),
		)
		if err != nil {
			otel.Handle(err)
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ctx := r.Context()
			if opts.InjectPropagationHeader {
				otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(w.Header()))
			}

			// wrap response write to get access to status & size
			wr := WrapResponseWriter(w)

			labeler, _ := otelhttp.LabelerFromContext(ctx)
			attributes := append(labeler.Get(), semconv.HTTPServerMetricAttributesFromHTTPRequest(name, r)...)

			next.ServeHTTP(wr, r)
			elapsedTime := time.Since(start).Seconds()

			serverLatency.Record(ctx, elapsedTime, attributes...)
			c.Add(ctx, 1, append(attributes, semconv.HTTPStatusCodeKey.Int(wr.StatusCode()))...)
		})
	}
}
