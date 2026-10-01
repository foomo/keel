package roundtripware

import (
	"net/http"
	"time"

	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

type (
	// LoggerOptions configures the [Logger] RoundTripware.
	LoggerOptions struct {
		// Message is the log message for completed requests.
		Message string
		// ErrorMessage is the log message for requests that failed with an
		// error.
		ErrorMessage string
		// MinWarnCode is the status code from which requests are logged at
		// warn level. Zero disables it.
		MinWarnCode int
		// MinErrorCode is the status code from which requests are logged at
		// error level. Zero disables it.
		MinErrorCode int
	}
	// LoggerOption configures [LoggerOptions].
	LoggerOption func(*LoggerOptions)
)

// GetDefaultLoggerOptions returns the default options: warn from 400 and
// error from 500.
func GetDefaultLoggerOptions() LoggerOptions {
	return LoggerOptions{
		Message:      "sent request",
		ErrorMessage: "failed to sent request",
		MinWarnCode:  400,
		MinErrorCode: 500,
	}
}

// LoggerWithMessage sets the log message. Defaults to "sent request".
func LoggerWithMessage(v string) LoggerOption {
	return func(o *LoggerOptions) {
		o.Message = v
	}
}

// LoggerWithErrorMessage sets the log message for failed requests. Defaults
// to "failed to sent request".
func LoggerWithErrorMessage(v string) LoggerOption {
	return func(o *LoggerOptions) {
		o.ErrorMessage = v
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

// Logger returns a RoundTripware that logs every outgoing request with its
// duration and, on success, status code and response size. The level depends
// on the error and status code.
func Logger(opts ...LoggerOption) RoundTripware {
	o := GetDefaultLoggerOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}

	return func(l *zap.Logger, next Handler) Handler {
		return func(r *http.Request) (*http.Response, error) {
			span := trace.SpanFromContext(r.Context())
			if span.IsRecording() {
				span.AddEvent("Logger")
			}

			start := time.Now()
			statusCode := http.StatusTeapot

			// extend logger using local instance
			l := log.WithHTTPRequestOut(l, r)

			// execute next handler
			resp, err := next(r)
			if err != nil {
				l = log.WithError(l, err)
			} else if resp != nil {
				l = log.With(l, log.Attributes(
					semconv.HTTPResponseStatusCode(resp.StatusCode),
					semconv.HTTPResponseBodySizeKey.Int64(resp.ContentLength),
				)...)
				statusCode = resp.StatusCode
			}

			l = l.With(log.FDuration(time.Since(start)))

			switch {
			case err != nil:
				l.Error(o.ErrorMessage)
			case o.MinErrorCode > 0 && statusCode >= o.MinErrorCode:
				l.Error(o.Message)
			case o.MinWarnCode > 0 && statusCode >= o.MinWarnCode:
				l.Warn(o.Message)
			default:
				l.Info(o.Message)
			}

			return resp, err
		}
	}
}
