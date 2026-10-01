package log

import (
	"context"
	"net/http"

	"github.com/foomo/keel/log"
)

// loggerLabelerContextKey is the context key under which the request
// [log.Labeler] is stored.
const loggerLabelerContextKey log.LabelerContextKey = "github.com/foomo/keel/net/log.Labeler"

// LabelerFromContext returns the [log.Labeler] injected into ctx by
// [InjectLabelerIntoContext]. It reports false if none is present.
func LabelerFromContext(ctx context.Context) (*log.Labeler, bool) {
	return log.LabelerFromContext(ctx, loggerLabelerContextKey)
}

// LabelerFromRequest returns the [log.Labeler] injected into the context of r.
// It reports false if none is present.
func LabelerFromRequest(r *http.Request) (*log.Labeler, bool) {
	return log.LabelerFromContext(r.Context(), loggerLabelerContextKey)
}

// InjectLabelerIntoContext returns a copy of ctx carrying a fresh, empty
// [log.Labeler], together with that labeler.
func InjectLabelerIntoContext(ctx context.Context) (context.Context, *log.Labeler) {
	return log.InjectLabeler(ctx, loggerLabelerContextKey)
}

// InjectLabelerIntoRequest returns a shallow copy of r whose context carries a
// fresh, empty [log.Labeler], together with that labeler.
func InjectLabelerIntoRequest(r *http.Request) (*http.Request, *log.Labeler) {
	ctx, labeler := InjectLabelerIntoContext(r.Context())
	return r.WithContext(ctx), labeler
}
