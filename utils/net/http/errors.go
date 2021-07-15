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
		log.Configure(l).HTTPRequest(r).Error(err).Logger().Error("http server error", zap.Int("code", code))
		http.Error(w, http.StatusText(code), code) // TODO enrich headers
	}
}
