package middleware

import (
	"net/http"
	"time"

	keelhttp "github.com/foomo/keel/net/http"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

type (
	// ResponseTimeOptions configures the [ResponseTime] middleware.
	ResponseTimeOptions struct {
		// SetHeader reports whether the X-Response-Time header, the time in
		// microseconds until the response header is written, is set.
		SetHeader bool
		// MaxDuration is the duration above which a warning is logged. Zero
		// disables the warning.
		MaxDuration time.Duration
		// MaxDurationMessage is the message of that warning.
		MaxDurationMessage string
	}
	// ResponseTimeOption configures [ResponseTimeOptions].
	ResponseTimeOption func(*ResponseTimeOptions)
)

// GetDefaultResponseTimeOptions returns the default options, which set the
// header and log no warnings.
func GetDefaultResponseTimeOptions() ResponseTimeOptions {
	return ResponseTimeOptions{
		SetHeader:          true,
		MaxDurationMessage: "max response time exceeded",
	}
}

// ResponseTimeWithMaxDurationMessage sets the warning message. Defaults to
// "max response time exceeded".
func ResponseTimeWithMaxDurationMessage(v string) ResponseTimeOption {
	return func(o *ResponseTimeOptions) {
		o.MaxDurationMessage = v
	}
}

// ResponseTimeWithMaxDuration sets the duration above which a warning is
// logged. Defaults to 0, which disables the warning.
func ResponseTimeWithMaxDuration(v time.Duration) ResponseTimeOption {
	return func(o *ResponseTimeOptions) {
		o.MaxDuration = v
	}
}

// ResponseTimeWithSetHeader sets [ResponseTimeOptions.SetHeader]. Defaults to
// true.
func ResponseTimeWithSetHeader(v bool) ResponseTimeOption {
	return func(o *ResponseTimeOptions) {
		o.SetHeader = v
	}
}

// ResponseTime returns a middleware that measures the handler duration, sets
// the X-Response-Time header and logs a warning when the duration exceeds
// the configured maximum.
func ResponseTime(opts ...ResponseTimeOption) keelhttp.Middleware {
	options := GetDefaultResponseTimeOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return ResponseTimeWithOptions(options)
}

// ResponseTimeWithOptions is like [ResponseTime] but takes fully populated
// options.
func ResponseTimeWithOptions(opts ResponseTimeOptions) keelhttp.Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			span := trace.SpanFromContext(r.Context())
			if span.IsRecording() {
				span.AddEvent("ResponseTime")
			}

			start := time.Now()
			rw := WrapResponseWriter(w)
			rw.SetWriteResponseTimeHeader(opts.SetHeader)
			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			if opts.MaxDuration > 0 && duration > opts.MaxDuration {
				l.Warn(opts.MaxDurationMessage, log.FDuration(opts.MaxDuration), log.FValue(duration.Microseconds()))
			}
		})
	}
}
