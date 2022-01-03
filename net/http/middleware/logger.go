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
		Message string
	}
	LoggerOption func(*LoggerOptions)
)

// GetDefaultLoggerOptions returns the default options
func GetDefaultLoggerOptions() LoggerOptions {
	return LoggerOptions{
		Message: "handled http request",
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

// LoggerWithOptions middleware
func LoggerWithOptions(opts LoggerOptions) Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := keeltime.Now()

			// wrap response write to get access to status & size
			wr := WrapResponseWriter(w)

			next.ServeHTTP(wr, r)

			log.WithHTTPRequest(l, r).Info(
				opts.Message,
				log.FDuration(time.Since(start)),
				log.FHTTPStatusCode(wr.StatusCode()),
				log.FHTTPWroteBytes(int64(wr.Size())),
			)
		})
	}
}
