package keel

import (
	"context"
	"fmt"

	"github.com/foomo/keel/interfaces"
	"github.com/foomo/keel/log"
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

		cl := l.With(log.FName(fmt.Sprintf("%T", closer)))
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
			cl.Warn("keel closer failed", zap.Error(err))
		} else {
			cl.Debug("keel closer closed")
		}
	}
}

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
