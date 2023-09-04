package roundtripware

import (
	"errors"
	"net/http"
	"time"

	keelerrors "github.com/foomo/keel/errors"
	"github.com/foomo/keel/log"
	"github.com/sony/gobreaker"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	"go.uber.org/zap"
)

var (
	// ErrCircuitBreaker is returned when the request failed because the circuit breaker did not let it go to the next
	// RoundTripware. It wraps the two gobreaker errors (ErrTooManyRequests & ErrOpenState) so only one comparison is
	// needed
	ErrCircuitBreaker = errors.New("circuit breaker triggered")

	// ErrIgnoreSuccessfulness can be returned by the IsSuccessful callback in order for the RoundTripware to ignore the
	// result of the function
	ErrIgnoreSuccessfulness = errors.New("ignored successfulness")

	// ErrReadFromActualBody when it is attempted to read from a body in the IsSuccessful callback that has not
	// previously been copied.
	ErrReadFromActualBody = errors.New("read from actual body")
)

// CircuitBreakerSettings is a copy of the gobreaker.Settings, except that the IsSuccessful function is omitted since we
// want to allow access to the request and response. See `CircuitBreakerWithIsSuccessful` for more.
type CircuitBreakerSettings struct {
	// Name is the name of the CircuitBreaker.
	Name string
	// MaxRequests is the maximum number of requests allowed to pass through
	// when the CircuitBreaker is half-open.
	// If MaxRequests is 0, the CircuitBreaker allows only 1 request.
	MaxRequests uint32
	// Interval is the cyclic period of the closed state
	// for the CircuitBreaker to clear the internal Counts.
	// If Interval is less than or equal to 0, the CircuitBreaker doesn't clear internal Counts during the closed state.
	Interval time.Duration
	// Timeout is the period of the open state,
	// after which the state of the CircuitBreaker becomes half-open.
	// If Timeout is less than or equal to 0, the timeout value of the CircuitBreaker is set to 60 seconds.
	Timeout time.Duration
	// ReadyToTrip is called with a copy of Counts whenever a request fails in the closed state.
	// If ReadyToTrip returns true, the CircuitBreaker will be placed into the open state.
	// If ReadyToTrip is nil, default ReadyToTrip is used.
	// Default ReadyToTrip returns true when the number of consecutive failures is more than 5.
	ReadyToTrip func(counts gobreaker.Counts) bool
	// OnStateChange is called whenever the state of the CircuitBreaker changes.
	OnStateChange func(name string, from gobreaker.State, to gobreaker.State)
}

type CircuitBreakerOptions struct {
	Counter syncint64.Counter

	IsSuccessful func(err error, req *http.Request, resp *http.Response) error
	CopyReqBody  bool
	CopyRespBody bool
}

func NewDefaultCircuitBreakerOptions() *CircuitBreakerOptions {
	return &CircuitBreakerOptions{
		Counter: nil,

		IsSuccessful: func(err error, req *http.Request, resp *http.Response) error {
			return err
		},
		CopyReqBody:  false,
		CopyRespBody: false,
	}
}

type CircuitBreakerOption func(opts *CircuitBreakerOptions)

// CircuitBreakerWithMetric adds a metric that counts the (un-)successful requests
func CircuitBreakerWithMetric(
	meter metric.Meter,
	meterName string,
	meterDescription string,
) CircuitBreakerOption {
	// intitialize the success counter
	counter, err := meter.SyncInt64().Counter(
		meterName,
		instrument.WithDescription(meterDescription),
	)
	if err != nil {
		panic(err)
	}

	return func(opts *CircuitBreakerOptions) {
		opts.Counter = counter
	}
}

func CircuitBreakerWithIsSuccessful(
	isSuccessful func(err error, req *http.Request, resp *http.Response) (e error),
	copyReqBody bool,
	copyRespBody bool,
) CircuitBreakerOption {
	return func(opts *CircuitBreakerOptions) {
		opts.IsSuccessful = isSuccessful
		opts.CopyReqBody = copyReqBody
		opts.CopyRespBody = copyRespBody
	}
}

// CircuitBreaker returns a RoundTripper which wraps all the following RoundTripwares and the Handler with a circuit
// breaker. This will prevent further request once a certain number of requests failed.
// NOTE: It's strongly advised to add this Roundripware before the metric middleware (if both are used). As the measure-
// ments of the execution time will otherwise be falsified
func CircuitBreaker(set *CircuitBreakerSettings, opts ...CircuitBreakerOption) RoundTripware {
	// intitialize the options
	o := NewDefaultCircuitBreakerOptions()
	for _, opt := range opts {
		opt(o)
	}

	// intitialize the gobreaker
	cbrSettings := gobreaker.Settings{
		Name:          set.Name,
		MaxRequests:   set.MaxRequests,
		Interval:      set.Interval,
		Timeout:       set.Timeout,
		ReadyToTrip:   set.ReadyToTrip,
		OnStateChange: set.OnStateChange,
	}
	circuitBreaker := gobreaker.NewTwoStepCircuitBreaker(cbrSettings)

	return func(l *zap.Logger, next Handler) Handler {
		return func(r *http.Request) (resp *http.Response, err error) {
			if r == nil {
				return nil, errors.New("request is nil")
			}

			// we need to detect the state change by ourselves, because the context does not allow us to hand in a context
			fromState := circuitBreaker.State()

			defer func() {
				// detect and log a state change
				toState := circuitBreaker.State()
				if fromState != toState {
					l.Warn("state change occurred",
						zap.String("state_from", fromState.String()),
						zap.String("state_to", toState.String()),
					)
				}

				attributes := []attribute.KeyValue{
					attribute.String("current_state", toState.String()),
					attribute.String("previous_state", fromState.String()),
					attribute.Bool("state_change", fromState != toState),
				}
				if err != nil {
					if o.Counter != nil {
						attributes := append(attributes, attribute.Bool("error", true))
						o.Counter.Add(r.Context(), 1, attributes...)
					}
				} else if o.Counter != nil {
					attributes := append(attributes, attribute.Bool("error", false))
					o.Counter.Add(r.Context(), 1, attributes...)
				}
			}()

			// clone the request and the body if wanted
			var errCopy error
			reqCopy, errCopy := copyRequest(r, o.CopyReqBody)
			if errCopy != nil {
				l.Error("unable to copy request", log.FError(errCopy))
				return nil, errCopy
			} else if o.CopyReqBody && reqCopy.Body != nil {
				// make sure the body is closed again - since it is a NopCloser it does not make a difference though
				defer reqCopy.Body.Close()
			}

			// check whether the circuit breaker is closed (an error is returned if not)
			done, err := circuitBreaker.Allow()

			// wrap the error in case it was produced because of the circuit breaker being (half-)open
			if errors.Is(err, gobreaker.ErrTooManyRequests) || errors.Is(err, gobreaker.ErrOpenState) {
				return nil, keelerrors.NewWrappedError(ErrCircuitBreaker, err)
			} else if err != nil {
				l.Error("unexpected error in circuit breaker",
					log.FError(err),
					zap.String("state", fromState.String()),
				)
				return nil, err
			}

			// continue with the middleware chain
			resp, err = next(r) //nolint:bodyclose

			var respCopy *http.Response
			if resp != nil {
				// clone the response and the body if wanted
				respCopy, errCopy = copyResponse(resp, o.CopyRespBody)
				if errCopy != nil {
					l.Error("unable to copy response", log.FError(errCopy))
					return nil, errCopy
				} else if o.CopyRespBody && respCopy.Body != nil {
					// make sure the body is closed again - since it is a NopCloser it does not make a difference though
					defer respCopy.Body.Close()
				}
			}

			if errSuccess := o.IsSuccessful(err, reqCopy, respCopy); errors.Is(errSuccess, errNoBody) {
				l.Error("encountered read from not previously copied request/response body",
					zap.Bool("copy_request", o.CopyReqBody),
					zap.Bool("copy_response", o.CopyRespBody),
				)
				// we actually want to return an error instead of the original request and error since the user
				// should be made aware that there is a misconfiguration
				return nil, ErrReadFromActualBody
			} else if !errors.Is(errSuccess, ErrIgnoreSuccessfulness) {
				done(errSuccess == nil)
			}

			return resp, nil
		}
	}
}
