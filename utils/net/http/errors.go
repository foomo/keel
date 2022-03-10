package httputils

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

// InternalServerError http response
func InternalServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, err error) {
	ServerError(l, w, r, http.StatusInternalServerError, err)
}

// InternalServiceUnavailable http response
func InternalServiceUnavailable(l *zap.Logger, w http.ResponseWriter, r *http.Request, err error) {
	ServerError(l, w, r, http.StatusServiceUnavailable, err)
}

// InternalServiceTooEarly http response
func InternalServiceTooEarly(l *zap.Logger, w http.ResponseWriter, r *http.Request, err error) {
	ServerError(l, w, r, http.StatusTooEarly, err)
}

// UnauthorizedServerError http response
func UnauthorizedServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, err error) {
	ServerError(l, w, r, http.StatusUnauthorized, err)
}

// BadRequestServerError http response
func BadRequestServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, err error) {
	ServerError(l, w, r, http.StatusBadRequest, err)
}

// NotFoundServerError http response
func NotFoundServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, err error) {
	ServerError(l, w, r, http.StatusNotFound, err)
}

// ServerError http response
func ServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, code int, err error) {
	if err != nil {
		log.WithHTTPRequest(l, r).Error("http server error", log.FError(err), log.FHTTPStatusCode(code))
		// w.Header().Set(keelhttp.HeaderXError, err.Error()) TODO make configurable with better value
		http.Error(w, http.StatusText(code), code)
	}
}
