package messaging

import (
	"context"

	"github.com/foomo/keel/log"
	"go.uber.org/zap"
)

type pipeConfig[T any] struct {
	filter     Filter[T]
	deadLetter DeadLetterFunc[T]
}

// PipeOption configures a Pipe or PipeMap call.
type PipeOption[T any] func(*pipeConfig[T])

// WithFilter registers a filter that runs before publish (or before map in
// PipeMap). Messages for which the filter returns false are silently dropped
// and logged. Filter errors are treated as false.
func WithFilter[T any](f Filter[T]) PipeOption[T] {
	return func(c *pipeConfig[T]) { c.filter = f }
}

// WithDeadLetter registers a dead-letter handler called when MapFunc returns
// an error or when the publisher fails after all retries are exhausted.
// The original Message[T] and the terminal error are passed to the handler.
func WithDeadLetter[T any](fn DeadLetterFunc[T]) PipeOption[T] {
	return func(c *pipeConfig[T]) { c.deadLetter = fn }
}

// ---------------------------------------------------------------------------
// Pipe — same-type wiring
// ---------------------------------------------------------------------------

// Pipe returns a Handler[T] that forwards every accepted message to pub.
// Filters run first; a dropped message never reaches pub.
// A publish error is returned to the subscriber as-is — wrap pub with
// NewRetryPublisher to add retry/backoff before that error surfaces.
func Pipe[T any](pub Publisher[T], opts ...PipeOption[T]) Handler[T] {
	cfg := buildConfig(opts)
	l := log.Logger()

	return func(ctx context.Context, msg Message[T]) error {
		if dropped := applyFilter(ctx, cfg.filter, msg, l); dropped {
			return nil
		}

		if err := pub.Publish(ctx, msg.Subject, msg.Payload); err != nil {
			if cfg.deadLetter != nil {
				cfg.deadLetter(ctx, msg, err)
			}
			return err
		}
		return nil
	}
}

// ---------------------------------------------------------------------------
// PipeMap — type-changing wiring
// ---------------------------------------------------------------------------

// PipeMap returns a Handler[T] that maps each message from T to U before
// publishing. Filters run on T before the map. A map error routes the original
// Message[T] to the dead-letter handler (if set) and drops the message.
// A publish error after a successful map is also dead-lettered with the
// original T message.
func PipeMap[T, U any](pub Publisher[U], mapFn MapFunc[T, U], opts ...PipeOption[T]) Handler[T] {
	cfg := buildConfig(opts)
	l := log.Logger()

	return func(ctx context.Context, msg Message[T]) error {
		if dropped := applyFilter(ctx, cfg.filter, msg, l); dropped {
			return nil
		}

		mapped, err := mapFn(ctx, msg)
		if err != nil {
			l.Error("pipemap: map failed, dropping message",
				zap.String("subject", msg.Subject),
				zap.Error(err),
			)
			if cfg.deadLetter != nil {
				cfg.deadLetter(ctx, msg, err)
			}
			return nil // map errors are non-fatal to the subscriber
		}

		if err := pub.Publish(ctx, mapped.Subject, mapped.Payload); err != nil {
			if cfg.deadLetter != nil {
				cfg.deadLetter(ctx, msg, err)
			}
			return err
		}
		return nil
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func buildConfig[T any](opts []PipeOption[T]) *pipeConfig[T] {
	cfg := &pipeConfig[T]{}
	for _, o := range opts {
		o(cfg)
	}
	return cfg
}

// applyFilter returns true if the message should be dropped.
func applyFilter[T any](ctx context.Context, f Filter[T], msg Message[T], l *zap.Logger) bool {
	if f == nil {
		return false
	}
	ok, err := f(ctx, msg)
	if err != nil {
		l.Warn("pipe: filter error, dropping message",
			zap.String("subject", msg.Subject),
			zap.Error(err),
		)
		return true
	}
	if !ok {
		l.Debug("pipe: message filtered",
			zap.String("subject", msg.Subject),
		)
		return true
	}
	return false
}
