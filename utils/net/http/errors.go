package httputils

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

// InternalServerError http response
func InternalServerError(l *zap.Logger, w http.ResponseWriter, r *http.Request, err error) {
	if err != nil {
		log.Configure(l).HTTPRequest(r).Error(err).Logger().Error("http internal server error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError) // TODO enrich headers
	}
}
