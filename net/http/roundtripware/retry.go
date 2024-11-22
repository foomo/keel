package roundtripware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/avast/retry-go/v4"
	"go.uber.org/zap"
)

type (
	RetryOptions struct {
		Handler      RetryHandler
		retryOptions []retry.Option
	}
	RetryHandler func(*http.Response) error
	RetryOption  func(*RetryOptions)
)

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

func RetryWithAttempts(v uint) RetryOption {
	return func(o *RetryOptions) {
		o.retryOptions = append(o.retryOptions, retry.Attempts(v))
	}
}

func RetryWithContext(ctx context.Context) RetryOption {
	return func(o *RetryOptions) {
		o.retryOptions = append(o.retryOptions, retry.Context(ctx))
	}
}

func RetryWithDelay(delay time.Duration) RetryOption {
	return func(o *RetryOptions) {
		o.retryOptions = append(o.retryOptions, retry.Delay(delay))
	}
}

func RetryWithMaxDelay(maxDelay time.Duration) RetryOption {
	return func(o *RetryOptions) {
		o.retryOptions = append(o.retryOptions, retry.MaxDelay(maxDelay))
	}
}

func RetryWithDelayType(delayType retry.DelayTypeFunc) RetryOption {
	return func(o *RetryOptions) {
		o.retryOptions = append(o.retryOptions, retry.DelayType(delayType))
	}
}

func RetryWithOnRetry(onRetry retry.OnRetryFunc) RetryOption {
	return func(o *RetryOptions) {
		o.retryOptions = append(o.retryOptions, retry.OnRetry(onRetry))
	}
}

func RetryWithLastErrorOnly(lastErrorOnly bool) RetryOption {
	return func(o *RetryOptions) {
		o.retryOptions = append(o.retryOptions, retry.LastErrorOnly(lastErrorOnly))
	}
}

func RetryWithRetryIf(retryIf retry.RetryIfFunc) RetryOption {
	return func(o *RetryOptions) {
		o.retryOptions = append(o.retryOptions, retry.RetryIf(retryIf))
	}
}

// Retry returns a RoundTripper which retries failed requests
func Retry(opts ...RetryOption) RoundTripware {
	o := GetDefaultRetryOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	return func(l *zap.Logger, next Handler) Handler {
		return func(req *http.Request) (*http.Response, error) {
			var resp *http.Response
			err := retry.Do(func() error {
				var err error
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
