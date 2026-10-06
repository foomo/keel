package keel

import (
	"context"

	"github.com/foomo/keel/interfaces"
	"github.com/foomo/keel/log"
	keelsemconv "github.com/foomo/keel/semconv"
	"go.uber.org/zap"
)

// closeAll calls the matching close method on every given closer, bounded by the
// provided context. Nil closers are skipped. Failures are logged but do not stop
// the remaining closers from being closed. It is shared by Server graceful
// shutdown and Job finalization to keep the closer interface contract in one place.
func closeAll(ctx context.Context, l *zap.Logger, closers []any) {
	for _, closer := range closers {
		if closer == nil {
			continue
		}

		var err error

		cl := log.WithAttributes(l, keelsemconv.KeelCloserType(closer))
		switch c := closer.(type) {
		case interfaces.Closer:
			c.Close()
		case interfaces.ErrorCloser:
			err = c.Close()
		case interfaces.CloserWithContext:
			c.Close(ctx)
		case interfaces.ErrorCloserWithContext:
			err = c.Close(ctx)
		case interfaces.Shutdowner:
			c.Shutdown()
		case interfaces.ErrorShutdowner:
			err = c.Shutdown()
		case interfaces.ShutdownerWithContext:
			c.Shutdown(ctx)
		case interfaces.ErrorShutdownerWithContext:
			err = c.Shutdown(ctx)
		case interfaces.Stopper:
			c.Stop()
		case interfaces.ErrorStopper:
			err = c.Stop()
		case interfaces.StopperWithContext:
			c.Stop(ctx)
		case interfaces.ErrorStopperWithContext:
			err = c.Stop(ctx)
		case interfaces.Unsubscriber:
			c.Unsubscribe()
		case interfaces.ErrorUnsubscriber:
			err = c.Unsubscribe()
		case interfaces.UnsubscriberWithContext:
			c.Unsubscribe(ctx)
		case interfaces.ErrorUnsubscriberWithContext:
			err = c.Unsubscribe(ctx)
		}

		if err != nil {
			log.WithError(cl, err).Warn("keel closer failed")
		} else {
			cl.Debug("keel closer closed")
		}
	}
}

// IsCloser reports whether v implements any of the closer interfaces declared
// in [github.com/foomo/keel/interfaces] (Closer, Shutdowner, Stopper or
// Unsubscriber in any of their error or context variants).
func IsCloser(v any) bool {
	switch v.(type) {
	case interfaces.Closer,
		interfaces.ErrorCloser,
		interfaces.CloserWithContext,
		interfaces.ErrorCloserWithContext,
		interfaces.Shutdowner,
		interfaces.ErrorShutdowner,
		interfaces.ShutdownerWithContext,
		interfaces.ErrorShutdownerWithContext,
		interfaces.Stopper,
		interfaces.ErrorStopper,
		interfaces.StopperWithContext,
		interfaces.ErrorStopperWithContext,
		interfaces.Unsubscriber,
		interfaces.ErrorUnsubscriber,
		interfaces.UnsubscriberWithContext,
		interfaces.ErrorUnsubscriberWithContext:
		return true
	default:
		return false
	}
}
