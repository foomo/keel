package roundtripware

import (
	"errors"
	"fmt"
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
	StateCounter syncint64.Counter

	SuccessCounter syncint64.Counter

	IsSuccessful func(err error, req *http.Request, resp *http.Response) error
	CopyReqBody  bool
	CopyRespBody bool
}

func NewDefaultCircuitBreakerOptions() *CircuitBreakerOptions {
	return &CircuitBreakerOptions{
		StateCounter: nil,

		SuccessCounter: nil,

		IsSuccessful: func(err error, req *http.Request, resp *http.Response) error {
			return err
		},
		CopyReqBody:  false,
		CopyRespBody: false,
	}
}

type CircuitBreakerOption func(opts *CircuitBreakerOptions)

// CircuitBreakerWithSuccessMetric adds a metric that counts the state changes of the circuit breaker
func CircuitBreakerWithStateChangeMetric(
	stateMeter metric.Meter,
	stateMeterName string,
	stateMeterDescription string,
) CircuitBreakerOption {
	// intitialize the state change counter
	stateCounter, err := stateMeter.SyncInt64().Counter(
		stateMeterName,
		instrument.WithDescription(stateMeterDescription),
	)
	if err != nil {
		panic(err)
	}

	return func(opts *CircuitBreakerOptions) {
		opts.StateCounter = stateCounter
	}
}

// CircuitBreakerWithSuccessMetric adds a metric that counts the (un-)successful requests
func CircuitBreakerWithSuccessMetric(
	successMeter metric.Meter,
	successMeterName string,
	successMeterDescription string,
) CircuitBreakerOption {
	// intitialize the success counter
	successCounter, err := successMeter.SyncInt64().Counter(
		successMeterName,
		instrument.WithDescription(successMeterDescription),
	)
	if err != nil {
		panic(err)
	}

	return func(opts *CircuitBreakerOptions) {
		opts.SuccessCounter = successCounter
	}
}

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
	circuitBreaker := gobreaker.NewCircuitBreaker(cbrSettings)

	return func(l *zap.Logger, next Handler) Handler {
		return func(r *http.Request) (*http.Response, error) {

			// we need to detect the state change by ourselves, because the context does not allow us to hand in a context
			fromState := circuitBreaker.State()

			// clone the request and the body if wanted
			reqCopy, errCopy := copyRequest(r, o.CopyReqBody)
			if errCopy != nil {
				l.Error("unable to copy request", log.FError(errCopy))
				return nil, errCopy
			} else if o.CopyReqBody && reqCopy.Body != nil {
				// make sure the body is closed again - since it is a NopCloser it does not make a difference though
				defer reqCopy.Body.Close()
			}

			// call the next handler enclosed in the circuit breaker.
			resp, err := circuitBreaker.Execute(func() (interface{}, error) {
				resp, err := next(r)

				// clone the response and the body if wanted
				respCopy, errCopy := copyResponse(resp, o.CopyRespBody)
				if errCopy != nil {
					l.Error("unable to copy response", log.FError(errCopy))
					return nil, errCopy
				} else if o.CopyRespBody && respCopy.Body != nil {
					// make sure the body is closed again - since it is a NopCloser it does not make a difference though
					defer respCopy.Body.Close()
				}

				return resp, o.IsSuccessful(err, reqCopy, respCopy)
			})

			// detect and log a state change
			toState := circuitBreaker.State()
			if fromState != toState {
				l.Warn("state change occurred",
					zap.String("from", fromState.String()),
					zap.String("to", toState.String()),
				)

				// recording the metric if desired
				if o.StateCounter != nil {
					attributes := []attribute.KeyValue{
						attribute.String("state_change", fmt.Sprintf("%s -> %s", fromState.String(), toState.String())),
					}
					o.StateCounter.Add(r.Context(), 1, attributes...)
				}
			}

			// wrap the error in case it was produced because of the circuit breaker being (half-)open
			if errors.Is(gobreaker.ErrTooManyRequests, err) || errors.Is(gobreaker.ErrOpenState, err) {
				err = keelerrors.NewWrappedError(ErrCircuitBreaker, err)
			}

			if err != nil {
				if o.SuccessCounter != nil {
					attributes := []attribute.KeyValue{
						attribute.Bool("success", false),
					}
					o.SuccessCounter.Add(r.Context(), 1, attributes...)
				}
				return nil, err
			}

			if o.SuccessCounter != nil {
				attributes := []attribute.KeyValue{
					attribute.Bool("success", true),
				}
				o.SuccessCounter.Add(r.Context(), 1, attributes...)
			}

			if res, ok := resp.(*http.Response); ok {
				return res, nil
			} else {
				return nil, errors.New("result is no *http.Response")
			}
		}
	}
}
