package log

import (
	"context"
	"net/http"

	"github.com/foomo/keel/log"
)

const loggerLabelerContextKey log.LabelerContextKey = "github.com/foomo/keel/net/log.Labeler"

func LabelerFromContext(ctx context.Context) (*log.Labeler, bool) {
	return log.LabelerFromContext(ctx, loggerLabelerContextKey)
}

func LabelerFromRequest(r *http.Request) (*log.Labeler, bool) {
	return log.LabelerFromContext(r.Context(), loggerLabelerContextKey)
}

func InjectLabelerIntoContext(ctx context.Context) (context.Context, *log.Labeler) {
	return log.InjectLabeler(ctx, loggerLabelerContextKey)
}

func InjectLabelerIntoRequest(r *http.Request) (*http.Request, *log.Labeler) {
	ctx, labeler := InjectLabelerIntoContext(r.Context())
	return r.WithContext(ctx), labeler
}
