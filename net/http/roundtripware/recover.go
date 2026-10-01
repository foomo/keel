package roundtripware

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

type (
	// RecoverOptions configures the [Recover] RoundTripware.
	RecoverOptions struct {
		// DisablePrintStack reports whether the stack trace is omitted from
		// the log entry.
		DisablePrintStack bool
	}
	// RecoverOption configures [RecoverOptions].
	RecoverOption func(options *RecoverOptions)
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

// Recover returns a RoundTripware that recovers and logs panics from the next
// handler. A recovered request returns a nil response and a nil error.
func Recover(opts ...RecoverOption) RoundTripware {
	options := GetDefaultRecoverOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return RecoverWithOptions(options)
}

// RecoverWithOptions is like [Recover] but takes fully populated options.
func RecoverWithOptions(opts RecoverOptions) RoundTripware {
	return func(l *zap.Logger, next Handler) Handler {
		return func(r *http.Request) (*http.Response, error) {
			span := trace.SpanFromContext(r.Context())
			if span.IsRecording() {
				span.AddEvent("Recover")
			}

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

			return next(r)
		}
	}
}
