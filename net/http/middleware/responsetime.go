package middleware

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

type (
	ResponseTimeOptions struct {
		SetHeader          bool
		MaxDuration        time.Duration
		MaxDurationMessage string
	}
	ResponseTimeOption func(*ResponseTimeOptions)
)

// GetDefaultResponseTimeOptions returns the default options
func GetDefaultResponseTimeOptions() ResponseTimeOptions {
	return ResponseTimeOptions{
		SetHeader:          true,
		MaxDurationMessage: "max response time exceeded",
	}
}

// ResponseTimeWithMaxDurationMessage middleware option
func ResponseTimeWithMaxDurationMessage(v string) ResponseTimeOption {
	return func(o *ResponseTimeOptions) {
		o.MaxDurationMessage = v
	}
}

// ResponseTimeWithMaxDuration middleware option
func ResponseTimeWithMaxDuration(v time.Duration) ResponseTimeOption {
	return func(o *ResponseTimeOptions) {
		o.MaxDuration = v
	}
}

// ResponseTimeWithSetHeader middleware option
func ResponseTimeWithSetHeader(v bool) ResponseTimeOption {
	return func(o *ResponseTimeOptions) {
		o.SetHeader = v
	}
}

// ResponseTime middleware
func ResponseTime(opts ...ResponseTimeOption) Middleware {
	options := GetDefaultResponseTimeOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return ResponseTimeWithOptions(options)
}

// ResponseTimeWithOptions middleware
func ResponseTimeWithOptions(opts ResponseTimeOptions) Middleware {
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
