package middleware

import (
	"net/http"

	"github.com/foomo/keel/log"
	keelhttp "github.com/foomo/keel/net/http"
	httplog "github.com/foomo/keel/net/http/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type (
	// TelemetryOptions configures the [Telemetry] middleware.
	TelemetryOptions struct {
		// Name is the operation name passed to otelhttp. When empty, the
		// service name is used.
		Name string
		// OtelOpts are passed to [otelhttp.NewHandler].
		OtelOpts []otelhttp.Option
		// InjectPropagationHeader reports whether the trace context is
		// injected into the response headers.
		InjectPropagationHeader bool
	}
	// TelemetryOption configures [TelemetryOptions].
	TelemetryOption func(*TelemetryOptions)
)

// GetDefaultTelemetryOptions returns the default options, which inject the
// propagation header.
func GetDefaultTelemetryOptions() TelemetryOptions {
	return TelemetryOptions{
		InjectPropagationHeader: true,
	}
}

// Telemetry returns a middleware that instruments the handler with
// [otelhttp.NewHandler]. It optionally injects the trace context into the
// response headers and, for sampled spans, adds the trace and span IDs to the
// request's log labeler.
func Telemetry(opts ...TelemetryOption) keelhttp.Middleware {
	options := GetDefaultTelemetryOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return TelemetryWithOptions(options)
}

// TelemetryWithName sets [TelemetryOptions.Name]. Defaults to the service
// name.
func TelemetryWithName(v string) TelemetryOption {
	return func(o *TelemetryOptions) {
		o.Name = v
	}
}

// TelemetryWithInjectPropagationHeader sets
// [TelemetryOptions.InjectPropagationHeader]. Defaults to true.
func TelemetryWithInjectPropagationHeader(v bool) TelemetryOption {
	return func(o *TelemetryOptions) {
		o.InjectPropagationHeader = v
	}
}

// TelemetryWithOtelOpts appends options passed to [otelhttp.NewHandler].
func TelemetryWithOtelOpts(v ...otelhttp.Option) TelemetryOption {
	return func(o *TelemetryOptions) {
		o.OtelOpts = append(o.OtelOpts, v...)
	}
}

// TelemetryWithOptions is like [Telemetry] but takes fully populated options.
func TelemetryWithOptions(opts TelemetryOptions) keelhttp.Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		if opts.Name != "" {
			name = opts.Name
		}

		return otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			span := trace.SpanFromContext(r.Context())
			if span.IsRecording() {
				span.AddEvent("Telemetry")
			}

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
