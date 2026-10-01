package middleware

import (
	"net/http"
	"time"

	keelhttp "github.com/foomo/keel/net/http"
	httplog "github.com/foomo/keel/net/http/log"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

type (
	// LoggerOptions configures the [Logger] middleware.
	LoggerOptions struct {
		// Message is the log message.
		Message string
		// MinWarnCode is the status code from which requests are logged at
		// warn level. Zero disables it.
		MinWarnCode int
		// MinErrorCode is the status code from which requests are logged at
		// error level. Zero disables it.
		MinErrorCode int
		// InjectLabeler reports whether a log labeler is injected into the
		// request so downstream handlers can add fields to the log entry.
		InjectLabeler bool
	}
	// LoggerOption configures [LoggerOptions].
	LoggerOption func(*LoggerOptions)
)

// GetDefaultLoggerOptions returns the default options: message "handled http
// request", warn from 400, error from 500 and labeler injection enabled.
func GetDefaultLoggerOptions() LoggerOptions {
	return LoggerOptions{
		Message:       "handled http request",
		MinWarnCode:   400,
		MinErrorCode:  500,
		InjectLabeler: true,
	}
}

// Logger returns a middleware that logs one entry per request once the
// handler returns. The entry carries request fields, duration, status code,
// response size and any labeler fields; its level depends on the status code.
func Logger(opts ...LoggerOption) keelhttp.Middleware {
	options := GetDefaultLoggerOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return LoggerWithOptions(options)
}

// LoggerWithMessage sets the log message.
func LoggerWithMessage(v string) LoggerOption {
	return func(o *LoggerOptions) {
		o.Message = v
	}
}

// LoggerWithMinWarnCode sets the status code from which requests are logged
// at warn level. Defaults to 400.
func LoggerWithMinWarnCode(v int) LoggerOption {
	return func(o *LoggerOptions) {
		o.MinWarnCode = v
	}
}

// LoggerWithMinErrorCode sets the status code from which requests are logged
// at error level. Defaults to 500.
func LoggerWithMinErrorCode(v int) LoggerOption {
	return func(o *LoggerOptions) {
		o.MinErrorCode = v
	}
}

// LoggerWithInjectLabeler sets [LoggerOptions.InjectLabeler]. Defaults to true.
func LoggerWithInjectLabeler(v bool) LoggerOption {
	return func(o *LoggerOptions) {
		o.InjectLabeler = v
	}
}

// LoggerWithOptions is like [Logger] but takes fully populated options.
func LoggerWithOptions(opts LoggerOptions) keelhttp.Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			span := trace.SpanFromContext(r.Context())
			if span.IsRecording() {
				span.AddEvent("Logger")
			}

			start := time.Now()

			// wrap response write to get access to status & size
			wr := WrapResponseWriter(w)

			l := log.WithHTTPRequest(l, r)

			var labeler *log.Labeler

			if labeler == nil && opts.InjectLabeler {
				r, labeler = httplog.InjectLabelerIntoRequest(r)
			}

			next.ServeHTTP(wr, r)

			l = l.With(
				log.FDuration(time.Since(start)),
				log.Attribute(semconv.HTTPResponseStatusCode(wr.StatusCode())),
				log.Attribute(semconv.HTTPResponseSize(wr.Size())),
			)

			if labeler != nil {
				l = l.With(labeler.Get()...)
			}

			if err := r.Context().Err(); err != nil {
				l = l.With(zap.String("error_context", err.Error()))
			}

			switch {
			case opts.MinErrorCode > 0 && wr.statusCode >= opts.MinErrorCode:
				l.Error(opts.Message)
			case opts.MinWarnCode > 0 && wr.statusCode >= opts.MinWarnCode:
				l.Warn(opts.Message)
			default:
				l.Info(opts.Message)
			}
		})
	}
}
