package middleware

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
	httputils "github.com/foomo/keel/utils/net/http"
)

type (
	RecoverOptions struct {
		DisablePrintStack bool
	}
	RecoverOption func(*RecoverOptions)
)

// GetDefaultRecoverOptions returns the default options
func GetDefaultRecoverOptions() RecoverOptions {
	return RecoverOptions{
		DisablePrintStack: false,
	}
}

// RecoverWithDisablePrintStack middleware option
func RecoverWithDisablePrintStack(v bool) RecoverOption {
	return func(o *RecoverOptions) {
		o.DisablePrintStack = v
	}
}

// Recover middleware
func Recover(opts ...RecoverOption) Middleware {
	options := GetDefaultRecoverOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	return RecoverWithOptions(options)
}

// RecoverWithOptions middleware
func RecoverWithOptions(opts RecoverOptions) Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if e := recover(); e != nil {
					if e == http.ErrAbortHandler {
						panic(e)
					}

					err, ok := e.(error)
					if !ok {
						err = fmt.Errorf("%v", e)
					}

					ll := log.WithError(l, err)
					if !opts.DisablePrintStack {
						ll = ll.With(log.FStackSkip(3))
					}

					httputils.InternalServerError(ll, w, r, errors.Wrap(err, "recovering from panic"))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
