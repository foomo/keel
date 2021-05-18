package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

type LoggerConfig struct{}

var DefaultLoggerConfig = LoggerConfig{}

func Logger() Middleware {
	return LoggerWithConfig(DefaultLoggerConfig)
}

func LoggerWithConfig(config LoggerConfig) Middleware {
	return func(l *zap.Logger, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// wrap response write to get access to status & size
			wr := wrapResponseWriter(w)

			next.ServeHTTP(wr, r)

			log.WithHTTPRequest(l, r).Info(
				"handled http request",
				log.FDuration(time.Since(start)),
				log.FHTTPStatusCode(wr.StatusCode()),
				log.FHTTPWroteBytes(int64(wr.Size())),
			)
		})
	}
}
