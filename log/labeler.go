package log

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

type LabelerContextKey string

type Labeler struct {
	mu     sync.Mutex
	fields []zap.Field
}

// Add attributes to a Labeler.
func (l *Labeler) Add(fields ...zap.Field) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.fields = append(l.fields, fields...)
}

// Get returns a copy of the attributes added to the Labeler.
func (l *Labeler) Get() []zap.Field {
	l.mu.Lock()
	defer l.mu.Unlock()
	ret := make([]zap.Field, len(l.fields))
	copy(ret, l.fields)
	return ret
}

func InjectLabeler(ctx context.Context, key LabelerContextKey) (context.Context, *Labeler) {
	l := &Labeler{}
	return context.WithValue(ctx, key, l), l
}

func LabelerFromContext(ctx context.Context, key LabelerContextKey) (*Labeler, bool) {
	if l, ok := ctx.Value(key).(*Labeler); ok {
		return l, true
	}
	return nil, false
}
