package http

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// NewServer creates and configures an HTTP server with the provided logger, name, address, handler, and middlewares.
func NewServer(l *zap.Logger, name, addr string, handler http.Handler, middlewares ...Middleware) *http.Server {
	return &http.Server{
		Addr:           addr,
		Handler:        Compose(l, name, handler, middlewares...),
		ErrorLog:       zap.NewStdLog(l),
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   60 * time.Second,
		IdleTimeout:    30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}
