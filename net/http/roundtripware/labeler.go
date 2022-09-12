package roundtripware

import (
	"context"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type labelerContextKeyType int

const labelerContextKey labelerContextKeyType = 0

func injectLabeler(ctx context.Context, l *otelhttp.Labeler) context.Context {
	return context.WithValue(ctx, labelerContextKey, l)
}

// LabelerFromContext retrieves a Labeler instance from the provided context if
// one is available.  If no Labeler was found in the provided context a new, empty
// Labeler is returned and the second return value is false.  In this case it is
// safe to use the Labeler but any attributes added to it will not be used.
func LabelerFromContext(ctx context.Context) (context.Context, *otelhttp.Labeler) {
	l, ok := ctx.Value(labelerContextKey).(*otelhttp.Labeler)
	if !ok {
		l = &otelhttp.Labeler{}
		ctx = injectLabeler(ctx, l)
	}
	return ctx, l
}
