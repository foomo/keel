package middleware

import (
	"net/http"
	"time"

	httplog "github.com/foomo/keel/net/http/log"
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
	keeltime "github.com/foomo/keel/time"
)

type (
	LoggerOptions struct {
		Message       string
		MinWarnCode   int
		MinErrorCode  int
		InjectLabeler bool
	}
	LoggerOption func(*LoggerOptions)
)

// GetDefaultLoggerOptions returns the default options
func GetDefaultLoggerOptions() LoggerOptions {
	return LoggerOptions{
		Message:       "handled http request",
		MinWarnCode:   400,
		MinErrorCode:  500,
		InjectLabeler: true,
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

// LoggerWithInjectLabeler middleware option
func LoggerWithInjectLabeler(v bool) LoggerOption {
	return func(o *LoggerOptions) {
		o.InjectLabeler = v
	}
}

// LoggerWithOptions middleware
func LoggerWithOptions(opts LoggerOptions) Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := keeltime.Now()

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
				log.FHTTPStatusCode(wr.StatusCode()),
				log.FHTTPWroteBytes(int64(wr.Size())),
			)

			if labeler != nil {
				l = l.With(labeler.Get()...)
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
