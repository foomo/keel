package roundtripware

import (
	"context"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// labelerContextKeyType is the type of the context key holding the labeler.
type labelerContextKeyType int

const labelerContextKey labelerContextKeyType = 0

func injectLabeler(ctx context.Context, l *otelhttp.Labeler) context.Context {
	return context.WithValue(ctx, labelerContextKey, l)
}

// LabelerFromContext returns the [otelhttp.Labeler] stored in ctx. If there is
// none, a new, empty Labeler is created and returned together with a derived
// context that holds it.
func LabelerFromContext(ctx context.Context) (context.Context, *otelhttp.Labeler) {
	l, ok := ctx.Value(labelerContextKey).(*otelhttp.Labeler)
	if !ok {
		l = &otelhttp.Labeler{}
		ctx = injectLabeler(ctx, l)
	}

	return ctx, l
}
