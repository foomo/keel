package context

import (
	"context"
)

const ContextKeyRequestID contextKey = "requestId"

func GetRequestID(ctx context.Context) (string, bool) {
	if value, ok := ctx.Value(ContextKeyRequestID).(string); ok {
		return value, true
	} else {
		return "", false
	}
}

func SetRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, ContextKeyRequestID, requestID)
}
