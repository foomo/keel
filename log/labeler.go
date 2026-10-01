package log

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

// LabelerContextKey is the type of context keys under which a [Labeler] is
// stored.
type LabelerContextKey string

// Labeler collects log fields that are added while a request is processed.
// It is safe for concurrent use. The zero value is ready to use.
type Labeler struct {
	mu     sync.Mutex
	fields []zap.Field
}

// Add appends fields to the Labeler.
func (l *Labeler) Add(fields ...zap.Field) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.fields = append(l.fields, fields...)
}

// Get returns a copy of the fields added to the Labeler.
func (l *Labeler) Get() []zap.Field {
	l.mu.Lock()
	defer l.mu.Unlock()

	ret := make([]zap.Field, len(l.fields))
	copy(ret, l.fields)

	return ret
}

// InjectLabeler returns a copy of ctx carrying a new [Labeler] under key,
// together with that Labeler.
func InjectLabeler(ctx context.Context, key LabelerContextKey) (context.Context, *Labeler) {
	l := &Labeler{}
	return context.WithValue(ctx, key, l), l
}

// LabelerFromContext returns the [Labeler] stored in ctx under key and
// reports whether one was found.
func LabelerFromContext(ctx context.Context, key LabelerContextKey) (*Labeler, bool) {
	if l, ok := ctx.Value(key).(*Labeler); ok {
		return l, true
	}

	return nil, false
}
