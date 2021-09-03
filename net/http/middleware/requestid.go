package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"

	httputils "github.com/foomo/keel/net/http"
)

type RequestIDConfig struct {
	Generator func() string
}

var DefaultRequestIDConfig = RequestIDConfig{}

func DefaultRequestIDGenerator() string {
	return uuid.New().String()
}

func RequestID() Middleware {
	return RequestIDWithConfig(DefaultRequestIDConfig)
}

func RequestIDWithConfig(config RequestIDConfig) Middleware {
	if config.Generator == nil {
		config.Generator = DefaultRequestIDGenerator
	}

	return func(l *zap.Logger, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if id := r.Header.Get(httputils.HeaderXRequestID); id == "" {
				r.Header.Set(httputils.HeaderXRequestID, config.Generator())
			}

			next.ServeHTTP(w, r)
		})
	}
}
