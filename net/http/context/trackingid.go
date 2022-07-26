package context

import (
	"context"
)

const ContextKeyTrackingID contextKey = "trackingId"

func GetTrackingID(ctx context.Context) (string, bool) {
	if value, ok := ctx.Value(ContextKeyTrackingID).(string); ok {
		return value, true
	} else {
		return "", false
	}
}

func SetTrackingID(ctx context.Context, trackingID string) context.Context {
	return context.WithValue(ctx, ContextKeyTrackingID, trackingID)
}
