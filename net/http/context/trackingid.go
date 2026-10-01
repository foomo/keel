package context

import (
	"context"
)

// ContextKeyTrackingID is the context key under which the tracking ID is stored.
const ContextKeyTrackingID contextKey = "trackingId"

// GetTrackingID returns the tracking ID stored in ctx by [SetTrackingID]. It reports false
// if no tracking ID is present.
func GetTrackingID(ctx context.Context) (string, bool) {
	if value, ok := ctx.Value(ContextKeyTrackingID).(string); ok {
		return value, true
	} else {
		return "", false
	}
}

// SetTrackingID returns a copy of ctx carrying trackingID as the tracking ID.
func SetTrackingID(ctx context.Context, trackingID string) context.Context {
	return context.WithValue(ctx, ContextKeyTrackingID, trackingID)
}
