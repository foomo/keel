package context

import (
	"context"
)

const ContextKeyReferer contextKey = "referer"

func GetReferer(ctx context.Context) (string, bool) {
	if value, ok := ctx.Value(ContextKeyReferer).(string); ok {
		return value, true
	} else {
		return "", false
	}
}

func SetReferer(ctx context.Context, referer string) context.Context {
	return context.WithValue(ctx, ContextKeyReferer, referer)
}
