package middleware

import (
	"fmt"
	"net/http"

	keelhttp "github.com/foomo/keel/net/http"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"runtime/debug"

	"github.com/foomo/keel/log"
	httputils "github.com/foomo/keel/utils/net/http"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

type (
	// RecoverOptions configures the [Recover] middleware.
	RecoverOptions struct {
		// DisablePrintStack reports whether the stack trace is omitted from
		// the log entry.
		DisablePrintStack bool
	}
	// RecoverOption configures [RecoverOptions].
	RecoverOption func(*RecoverOptions)
)

// GetDefaultRecoverOptions returns the default options, which include the
// stack trace.
func GetDefaultRecoverOptions() RecoverOptions {
	return RecoverOptions{
		DisablePrintStack: false,
	}
}

// RecoverWithDisablePrintStack sets [RecoverOptions.DisablePrintStack].
func RecoverWithDisablePrintStack(v bool) RecoverOption {
	return func(o *RecoverOptions) {
		o.DisablePrintStack = v
	}
}

// Recover returns a middleware that recovers panics from the next handler,
// logs them and responds with 500 Internal Server Error. Panics with
// [http.ErrAbortHandler] are re-raised.
func Recover(opts ...RecoverOption) keelhttp.Middleware {
	options := GetDefaultRecoverOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return RecoverWithOptions(options)
}

// RecoverWithOptions is like [Recover] but takes fully populated options.
func RecoverWithOptions(opts RecoverOptions) keelhttp.Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			span := trace.SpanFromContext(r.Context())
			span.AddEvent("Recover")

			defer func() {
				if e := recover(); e != nil {
					err, ok := e.(error)
					if !ok {
						err = fmt.Errorf("%v", e)
					}

					if errors.Is(err, http.ErrAbortHandler) {
						panic(e)
					}

					ll := log.WithError(l, err)
					if !opts.DisablePrintStack {
						ll = ll.With(log.Attribute(semconv.ExceptionStacktrace(string(debug.Stack()))))
					}

					httputils.InternalServerError(ll, w, r, errors.Wrap(err, "recovering from panic"))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
