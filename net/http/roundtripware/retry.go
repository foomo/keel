package roundtripware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/avast/retry-go/v4"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type (
	// RetryOptions configures the [Retry] RoundTripware.
	RetryOptions struct {
		// Handler classifies a response; a non-nil error triggers a retry.
		Handler      RetryHandler
		retryOptions []retry.Option
	}
	// RetryHandler returns a non-nil error if the response should be retried.
	RetryHandler func(*http.Response) error
	// RetryOption configures [RetryOptions].
	RetryOption func(*RetryOptions)
)

// GetDefaultRetryOptions returns the default options, which retry every
// response whose status is not 200 OK using the retry-go defaults.
func GetDefaultRetryOptions() RetryOptions {
	return RetryOptions{
		Handler: func(resp *http.Response) error {
			if resp.StatusCode != http.StatusOK {
				return errors.New("status code not ok")
			}

			return nil
		},
	}
}

// RetryWithAttempts applies [retry.Attempts].
func RetryWithAttempts(v uint) RetryOption {
	return func(o *RetryOptions) {
		o.retryOptions = append(o.retryOptions, retry.Attempts(v))
	}
}

// RetryWithContext applies [retry.Context].
func RetryWithContext(ctx context.Context) RetryOption {
	return func(o *RetryOptions) {
		o.retryOptions = append(o.retryOptions, retry.Context(ctx))
	}
}

// RetryWithDelay applies [retry.Delay].
func RetryWithDelay(delay time.Duration) RetryOption {
	return func(o *RetryOptions) {
		o.retryOptions = append(o.retryOptions, retry.Delay(delay))
	}
}

// RetryWithMaxDelay applies [retry.MaxDelay].
func RetryWithMaxDelay(maxDelay time.Duration) RetryOption {
	return func(o *RetryOptions) {
		o.retryOptions = append(o.retryOptions, retry.MaxDelay(maxDelay))
	}
}

// RetryWithDelayType applies [retry.DelayType].
func RetryWithDelayType(delayType retry.DelayTypeFunc) RetryOption {
	return func(o *RetryOptions) {
		o.retryOptions = append(o.retryOptions, retry.DelayType(delayType))
	}
}

// RetryWithOnRetry applies [retry.OnRetry].
func RetryWithOnRetry(onRetry retry.OnRetryFunc) RetryOption {
	return func(o *RetryOptions) {
		o.retryOptions = append(o.retryOptions, retry.OnRetry(onRetry))
	}
}

// RetryWithLastErrorOnly applies [retry.LastErrorOnly].
func RetryWithLastErrorOnly(lastErrorOnly bool) RetryOption {
	return func(o *RetryOptions) {
		o.retryOptions = append(o.retryOptions, retry.LastErrorOnly(lastErrorOnly))
	}
}

// RetryWithRetryIf applies [retry.RetryIf].
func RetryWithRetryIf(retryIf retry.RetryIfFunc) RetryOption {
	return func(o *RetryOptions) {
		o.retryOptions = append(o.retryOptions, retry.RetryIf(retryIf))
	}
}

// Retry returns a RoundTripware that retries requests failing with an error
// or rejected by the [RetryHandler]. It returns the last response and the
// retry-go error. Request bodies are not rewound between attempts.
func Retry(opts ...RetryOption) RoundTripware {
	o := GetDefaultRetryOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}

	return func(l *zap.Logger, next Handler) Handler {
		return func(req *http.Request) (*http.Response, error) {
			span := trace.SpanFromContext(req.Context())
			if span.IsRecording() {
				span.AddEvent("Retry")
			}

			var (
				attempt int
				resp    *http.Response
			)

			err := retry.Do(func() error {
				attempt++

				var err error

				if attempt > 1 && span.IsRecording() {
					span.SetAttributes(semconv.HTTPRequestResendCount(attempt - 1))
				}

				resp, err = next(req) //nolint:bodyclose
				if err != nil {
					return err
				}

				return o.Handler(resp)
			}, o.retryOptions...)

			return resp, err
		}
	}
}
