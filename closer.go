package keel

import (
	"github.com/foomo/keel/interfaces"
)

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
