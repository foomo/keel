package roundtripware

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

type (
	RecoverOptions struct {
		DisablePrintStack bool
	}
	RecoverOption func(options *RecoverOptions)
)

// GetDefaultRecoverOptions returns the default options
func GetDefaultRecoverOptions() RecoverOptions {
	return RecoverOptions{
		DisablePrintStack: false,
	}
}

// RecoverWithDisablePrintStack roundTripware option
func RecoverWithDisablePrintStack(v bool) RecoverOption {
	return func(o *RecoverOptions) {
		o.DisablePrintStack = v
	}
}

// Recover returns a RoundTripper which catches any panics
func Recover(opts ...RecoverOption) RoundTripware {
	options := GetDefaultRecoverOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	return RecoverWithOptions(options)
}

// RecoverWithOptions returns a RoundTripper which catches any panics
func RecoverWithOptions(opts RecoverOptions) RoundTripware {
	return func(l *zap.Logger, next Handler) Handler {
		return func(req *http.Request) (*http.Response, error) {
			defer func() {
				if e := recover(); e != nil {
					err, ok := e.(error)
					if !ok {
						err = fmt.Errorf("%v", e)
					}
					ll := log.WithError(l, err)
					if !opts.DisablePrintStack {
						ll = ll.With(log.FStackSkip(3))
					}
					ll.Error("recovering from panic")
				}
			}()
			return next(req)
		}
	}
}
