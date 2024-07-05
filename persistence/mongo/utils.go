package keelmongo

import (
	"context"

	"github.com/foomo/keel/log"
	"go.mongodb.org/mongo-driver/mongo"
)

// CloseCursor with defer
func CloseCursor(ctx context.Context, cursor *mongo.Cursor) {
	if err := cursor.Close(ctx); err != nil {
		log.WithError(nil, err).Error("failed to close cursor")
	}
}
