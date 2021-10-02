package keelmongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

// CloseCursor with defer
func CloseCursor(ctx context.Context, cursor *mongo.Cursor, err *error) {
	cErr := cursor.Close(ctx)
	if *err == nil {
		*err = cErr
	}
}
