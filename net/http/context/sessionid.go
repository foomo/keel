package context

import (
	"context"
)

// ContextKeySessionID is the context key under which the session ID is stored.
const ContextKeySessionID contextKey = "sessionId"

// GetSessionID returns the session ID stored in ctx by [SetSessionID]. It reports false
// if no session ID is present.
func GetSessionID(ctx context.Context) (string, bool) {
	if value, ok := ctx.Value(ContextKeySessionID).(string); ok {
		return value, true
	} else {
		return "", false
	}
}

// SetSessionID returns a copy of ctx carrying sessionID as the session ID.
func SetSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, ContextKeySessionID, sessionID)
}
