package roundtripware

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

type (
	LoggerOptions struct {
		Message      string
		ErrorMessage string
		MinWarnCode  int
		MinErrorCode int
	}
	LoggerOption func(*LoggerOptions)
)

// GetDefaultLoggerOptions returns the default options
func GetDefaultLoggerOptions() LoggerOptions {
	return LoggerOptions{
		Message:      "sent request",
		ErrorMessage: "failed to sent request",
		MinWarnCode:  400,
		MinErrorCode: 500,
	}
}

// LoggerWithMessage middleware option
func LoggerWithMessage(v string) LoggerOption {
	return func(o *LoggerOptions) {
		o.Message = v
	}
}

// LoggerWithErrorMessage middleware option
func LoggerWithErrorMessage(v string) LoggerOption {
	return func(o *LoggerOptions) {
		o.ErrorMessage = v
	}
}

// LoggerWithMinWarnCode middleware option
func LoggerWithMinWarnCode(v int) LoggerOption {
	return func(o *LoggerOptions) {
		o.MinWarnCode = v
	}
}

// LoggerWithMinErrorCode middleware option
func LoggerWithMinErrorCode(v int) LoggerOption {
	return func(o *LoggerOptions) {
		o.MinErrorCode = v
	}
}

// Logger returns a RoundTripware which logs all requests
func Logger(opts ...LoggerOption) RoundTripware {
	o := GetDefaultLoggerOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}

	return func(l *zap.Logger, next Handler) Handler {
		return func(req *http.Request) (*http.Response, error) {
			start := time.Now()
			statusCode := http.StatusTeapot

			// extend logger using local instance
			l := log.WithHTTPRequestOut(l, req)

			// execute next handler
			resp, err := next(req)
			if err != nil {
				l = log.WithError(l, err)
			} else if resp != nil {
				l = log.With(l,
					log.FHTTPStatusCode(resp.StatusCode),
					log.FHTTPRequestContentLength(resp.ContentLength),
				)
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
