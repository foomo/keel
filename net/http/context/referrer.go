package context

import (
	"context"
)

// ContextKeyReferer is the context key under which the referer is stored.
const ContextKeyReferer contextKey = "referer"

// GetReferer returns the referer stored in ctx by [SetReferer]. It reports false
// if no referer is present.
func GetReferer(ctx context.Context) (string, bool) {
	if value, ok := ctx.Value(ContextKeyReferer).(string); ok {
		return value, true
	} else {
		return "", false
	}
}

// SetReferer returns a copy of ctx carrying referer as the referer.
func SetReferer(ctx context.Context, referer string) context.Context {
	return context.WithValue(ctx, ContextKeyReferer, referer)
}
