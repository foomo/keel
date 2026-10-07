package roundtripware

import (
	"errors"
	"net/http"
	"time"

	"github.com/foomo/keel/log"
	keelsemconv "github.com/foomo/keel/semconv"
	foomosemconv "github.com/foomo/opentelemetry-go/semconv"
	"github.com/foomo/opentelemetry-go/semconv/circuitbreakerconv"
	"github.com/sony/gobreaker"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

var (
	// ErrCircuitBreaker is returned when the circuit breaker did not let the request pass to the next
	// RoundTripware. It is joined with the underlying gobreaker error (ErrTooManyRequests or ErrOpenState) so a
	// single [errors.Is] comparison is needed.
	ErrCircuitBreaker = errors.New("circuit breaker triggered")

	// ErrIgnoreSuccessfulness can be returned by the IsSuccessful callback to have the request counted neither as
	// success nor as failure.
	ErrIgnoreSuccessfulness = errors.New("ignored successfulness")

	// ErrReadFromActualBody is returned when the IsSuccessful callback reads a request or response body that was not
	// copied (see [CircuitBreakerWithIsSuccessful]).
	ErrReadFromActualBody = errors.New("read from actual body")
)

// CircuitBreakerSettings is a copy of [gobreaker.Settings] without the IsSuccessful function, which is replaced by
// one with access to the request and response. See [CircuitBreakerWithIsSuccessful].
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

// CircuitBreakerOptions configures the [CircuitBreaker] RoundTripware.
type CircuitBreakerOptions struct {
	// Counter, when non-nil, counts requests with state and error attributes.
	Counter metric.Int64Counter

	// IsSuccessful decides whether a request counts as success (nil error) or
	// failure. It receives copies of the request and response.
	IsSuccessful func(err error, req *http.Request, resp *http.Response) error
	// CopyReqBody reports whether the request body is copied for IsSuccessful.
	CopyReqBody bool
	// CopyRespBody reports whether the response body is copied for IsSuccessful.
	CopyRespBody bool
}

// NewDefaultCircuitBreakerOptions returns the default options: no counter, no
// body copies and an IsSuccessful that treats every transport error as failure.
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

// CircuitBreakerOption configures [CircuitBreakerOptions].
type CircuitBreakerOption func(opts *CircuitBreakerOptions)

// CircuitBreakerWithMetric adds an Int64Counter named meterName that counts
// successful and failed requests. It panics if the counter cannot be created.
func CircuitBreakerWithMetric(
	meter metric.Meter,
	meterName string,
	meterDescription string,
) CircuitBreakerOption {
	// intitialize the success counter
	counter, err := meter.Int64Counter(
		meterName,
		metric.WithDescription(meterDescription),
	)
	if err != nil {
		panic(err)
	}

	return func(opts *CircuitBreakerOptions) {
		opts.Counter = counter
	}
}

// CircuitBreakerWithIsSuccessful sets the callback that classifies a request
// as success (nil) or failure (non-nil); returning [ErrIgnoreSuccessfulness]
// ignores the request. copyReqBody and copyRespBody enable copying the bodies
// so the callback can read them; reading a body that was not copied makes the
// RoundTripware return [ErrReadFromActualBody].
func CircuitBreakerWithIsSuccessful(
	isSuccessful func(err error, req *http.Request, resp *http.Response) error,
	copyReqBody bool,
	copyRespBody bool,
) CircuitBreakerOption {
	return func(opts *CircuitBreakerOptions) {
		opts.IsSuccessful = isSuccessful
		opts.CopyReqBody = copyReqBody
		opts.CopyRespBody = copyRespBody
	}
}

// CircuitBreaker returns a RoundTripware which wraps all the following RoundTripwares and the Handler with a circuit
// breaker. It rejects requests with [ErrCircuitBreaker] once the breaker trips and logs state changes.
//
// It is strongly advised to add this RoundTripware before the metric RoundTripware (if both are used), as the
// execution time measurements will otherwise be falsified.
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
		return func(r *http.Request) (resp *http.Response, err error) { //nolint:nonamedreturns
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
						log.Attributes(
							keelsemconv.KeelCircuitBreakerPreviousState(string(circuitBreakerState(fromState))),
							foomosemconv.CircuitBreakerState(string(circuitBreakerState(toState))),
						)...,
					)
				}

				if o.Counter != nil {
					attributes := []attribute.KeyValue{
						foomosemconv.CircuitBreakerState(string(circuitBreakerState(toState))),
						keelsemconv.KeelCircuitBreakerPreviousState(string(circuitBreakerState(fromState))),
						keelsemconv.KeelCircuitBreakerStateChange(fromState != toState),
					}
					if err != nil {
						attributes = append(attributes, foomosemconv.ErrorType(err))
					}

					o.Counter.Add(r.Context(), 1, metric.WithAttributes(attributes...))
				}
			}()

			// clone the request and the body if wanted
			var errCopy error

			reqCopy, errCopy := copyRequest(r, o.CopyReqBody)
			if errCopy != nil {
				log.WithError(l, errCopy).Error("unable to copy request")
				return nil, errCopy
			} else if o.CopyReqBody && reqCopy.Body != nil {
				// make sure the body is closed again - since it is a NopCloser it does not make a difference though
				defer reqCopy.Body.Close()
			}

			// check whether the circuit breaker is closed (an error is returned if not)
			done, err := circuitBreaker.Allow()

			// wrap the error in case it was produced because of the circuit breaker being (half-)open
			if errors.Is(err, gobreaker.ErrTooManyRequests) || errors.Is(err, gobreaker.ErrOpenState) {
				return nil, errors.Join(ErrCircuitBreaker, err)
			} else if err != nil {
				log.WithError(l, err).Error("unexpected error in circuit breaker",
					log.Attribute(foomosemconv.CircuitBreakerState(string(circuitBreakerState(fromState)))),
				)

				return nil, err
			}

			// continue with the middleware chain
			resp, err = next(r)

			var respCopy *http.Response
			if resp != nil {
				// clone the response and the body if wanted
				respCopy, errCopy = copyResponse(resp, o.CopyRespBody)
				if errCopy != nil {
					log.WithError(l, errCopy).Error("unable to copy response")
					return nil, errCopy
				} else if o.CopyRespBody && respCopy.Body != nil {
					// make sure the body is closed again - since it is a NopCloser it does not make a difference though
					defer respCopy.Body.Close()
				}
			}

			if errSuccess := o.IsSuccessful(err, reqCopy, respCopy); errors.Is(errSuccess, errNoBody) {
				l.Error("encountered read from not previously copied request/response body")
				// we actually want to return an error instead of the original request and error since the user
				// should be made aware that there is a misconfiguration
				return nil, ErrReadFromActualBody
			} else if !errors.Is(errSuccess, ErrIgnoreSuccessfulness) {
				done(errSuccess == nil)
			}

			// return the response and error received from the next call
			return resp, err
		}
	}
}

// circuitBreakerState maps a gobreaker state to the circuit_breaker.state
// enum value.
func circuitBreakerState(s gobreaker.State) circuitbreakerconv.StateAttr {
	switch s {
	case gobreaker.StateOpen:
		return circuitbreakerconv.StateOpen
	case gobreaker.StateHalfOpen:
		return circuitbreakerconv.StateHalfOpen
	default:
		return circuitbreakerconv.StateClosed
	}
}
