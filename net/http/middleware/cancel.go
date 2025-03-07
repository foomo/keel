package middleware

import (
	"net/http"

	httputils "github.com/foomo/keel/utils/net/http"
	"go.uber.org/zap"
)

type (
	CancelOptions struct {
		Code int
	}
	CancelOption func(*CancelOptions)
)

// GetDefaultCancelOptions returns the default options
func GetDefaultCancelOptions() CancelOptions {
	return CancelOptions{
		Code: 499,
	}
}

// CancelWithCode middleware option
func CancelWithCode(v int) CancelOption {
	return func(o *CancelOptions) {
		o.Code = v
	}
}

// Cancel middleware
func Cancel(opts ...CancelOption) Middleware {
	options := GetDefaultCancelOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	return CancelWithOptions(options)
}

// CancelWithOptions middleware
func CancelWithOptions(opts CancelOptions) Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			done := make(chan struct{})
			go func() {
				next.ServeHTTP(w, r)
				close(done)
			}()

			select {
			case <-r.Context().Done():
				httputils.ServerError(l, w, r, opts.Code, r.Context().Err())
			case <-done:
				// If handler completes normally
			}
		})
	}
}
