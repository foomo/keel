package context

import (
	"context"
)

// ContextKeyRequestID is the context key under which the request ID is stored.
const ContextKeyRequestID contextKey = "requestId"

// GetRequestID returns the request ID stored in ctx by [SetRequestID]. It reports false
// if no request ID is present.
func GetRequestID(ctx context.Context) (string, bool) {
	if value, ok := ctx.Value(ContextKeyRequestID).(string); ok {
		return value, true
	} else {
		return "", false
	}
}

// SetRequestID returns a copy of ctx carrying requestID as the request ID.
func SetRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, ContextKeyRequestID, requestID)
}
