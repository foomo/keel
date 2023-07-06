package context

import (
	"context"
)

const ContextKeyReferrer contextKey = "referrer"

func GetReferrer(ctx context.Context) (string, bool) {
	if value, ok := ctx.Value(ContextKeyReferrer).(string); ok {
		return value, true
	} else {
		return "", false
	}
}

func SetReferrer(ctx context.Context, referer string) context.Context {
	return context.WithValue(ctx, ContextKeyReferrer, referer)
}
