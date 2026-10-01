package otel

import (
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

// ErrorHandler is an [otel.ErrorHandler] that logs OpenTelemetry errors to
// a zap logger. Errors with the message "not implemented yet" are logged at
// warn level, all others at error level; nil errors are ignored.
//
// [otel.ErrorHandler]: https://pkg.go.dev/go.opentelemetry.io/otel#ErrorHandler
type ErrorHandler struct {
	l *zap.Logger
}

// NewErrorHandler returns an [ErrorHandler] logging to l.
func NewErrorHandler(l *zap.Logger) *ErrorHandler {
	return &ErrorHandler{l}
}

// Handle logs err.
func (h *ErrorHandler) Handle(err error) {
	l := log.WithError(h.l, err)
	if err != nil && err.Error() == "not implemented yet" {
		l.Warn("otel error")
	} else if err != nil {
		l.Error("otel error")
	}
}

// SetLogger replaces the logger. It is not safe for concurrent use with
// [ErrorHandler.Handle].
func (h *ErrorHandler) SetLogger(l *zap.Logger) {
	h.l = l
}
