package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/foomo/keel/log"
	keeltime "github.com/foomo/keel/time"
)

type (
	LoggerOptions struct {
		Message      string
		MinWarnCode  int
		MinErrorCode int
	}
	LoggerOption func(*LoggerOptions)
)

// GetDefaultLoggerOptions returns the default options
func GetDefaultLoggerOptions() LoggerOptions {
	return LoggerOptions{
		Message:      "handled http request",
		MinWarnCode:  400,
		MinErrorCode: 500,
	}
}

// Logger middleware
func Logger(opts ...LoggerOption) Middleware {
	options := GetDefaultLoggerOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	return LoggerWithOptions(options)
}

// LoggerWithMessage middleware option
func LoggerWithMessage(v string) LoggerOption {
	return func(o *LoggerOptions) {
		o.Message = v
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

// LoggerWithOptions middleware
func LoggerWithOptions(opts LoggerOptions) Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := keeltime.Now()

			// wrap response write to get access to status & size
			wr := WrapResponseWriter(w)

			next.ServeHTTP(wr, r)

			l := log.WithHTTPRequest(l, r).With(
				log.FDuration(time.Since(start)),
				log.FHTTPStatusCode(wr.StatusCode()),
				log.FHTTPWroteBytes(int64(wr.Size())),
			)

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
