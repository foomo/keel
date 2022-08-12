package context

import (
	"context"
)

const ContextKeySessionID contextKey = "sessionId"

func GetSessionID(ctx context.Context) (string, bool) {
	if value, ok := ctx.Value(ContextKeySessionID).(string); ok {
		return value, true
	} else {
		return "", false
	}
}

func SetSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, ContextKeySessionID, sessionID)
}
