package roundtripware

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	keelerrors "github.com/foomo/keel/errors"
	"github.com/foomo/keel/log"
	"github.com/sony/gobreaker"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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

// CircuitBreakerSettings is a copy of the gobreaker.Settings, except that the IsSuccessful function is ommited since we
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

type circuitBreakerOptions struct {
	stateMeter            metric.Meter
	stateMeterName        string
	stateMeterDescription string

	successMeter            metric.Meter
	successMeterName        string
	successMeterDescription string

	isSuccessful func(err error, req *http.Request, resp *http.Response) error
	copyReqBody  bool
	copyRespBody bool
}

func newDefaultCircuitBreakerOptions() *circuitBreakerOptions {
	return &circuitBreakerOptions{
		stateMeter:            nil,
		stateMeterName:        "",
		stateMeterDescription: "",

		successMeter:            nil,
		successMeterName:        "",
		successMeterDescription: "",

		isSuccessful: func(err error, req *http.Request, resp *http.Response) error {
			return err
		},
		copyReqBody:  false,
		copyRespBody: false,
	}
}

type CircuitBreakerOption func(opts *circuitBreakerOptions)

// CircuitBreakerWithSuccessMetric adds a metric that counts the state changes of the circuit breaker
func CircuitBreakerWithStateChangeMetric(
	stateMeter metric.Meter,
	stateMeterName string,
	stateMeterDescription string,
) CircuitBreakerOption {
	return func(opts *circuitBreakerOptions) {
		opts.stateMeter = stateMeter
		opts.stateMeterName = stateMeterName
		opts.stateMeterDescription = stateMeterDescription
	}
}

// CircuitBreakerWithSuccessMetric adds a metric that counts the (un-)successful requests
func CircuitBreakerWithSuccessMetric(
	successMeter metric.Meter,
	successMeterName string,
	successMeterDescription string,
) CircuitBreakerOption {
	return func(opts *circuitBreakerOptions) {
		opts.successMeter = successMeter
		opts.successMeterName = successMeterName
		opts.successMeterDescription = successMeterDescription
	}
}

func CircuitBreakerWithIsSuccessful(
	isSuccessful func(err error, req *http.Request, resp *http.Response) error,
	copyReqBody bool,
	copyRespBody bool,
) CircuitBreakerOption {
	return func(opts *circuitBreakerOptions) {
		opts.isSuccessful = isSuccessful
		opts.copyReqBody = copyReqBody
		opts.copyRespBody = copyRespBody
	}
}

// CircuitBreaker returns a RoundTripper which wraps all the following RoundTripwares and the Handler with a circuit
// breaker. This will prevent further request once a certain number of requests failed.
// NOTE: It's strongly adviced to add this Roundripware before the metric middleware (if both are used). As the measure-
// ments of the execution time will otherwise be falsified
func CircuitBreaker(set *CircuitBreakerSettings, opts ...CircuitBreakerOption) RoundTripware {
	// intialize the options
	o := newDefaultCircuitBreakerOptions()
	for _, opt := range opts {
		opt(o)
	}

	// intialize the state change counter
	var stateCounter syncint64.Counter
	if o.stateMeter != nil {
		var err error
		stateCounter, err = o.stateMeter.SyncInt64().Counter(
			o.stateMeterName,
			instrument.WithDescription(o.stateMeterDescription),
		)
		if err != nil {
			panic(err)
		}
	}

	// intialize the state (un-)success counter
	var successCounter syncint64.Counter
	if o.successMeter != nil {
		var err error
		successCounter, err = o.successMeter.SyncInt64().Counter(
			o.successMeterName,
			instrument.WithDescription(o.successMeterDescription),
		)
		if err != nil {
			panic(err)
		}
	}

	// Initialize the gobreaker
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
			var (
				ctx     context.Context
				labeler *otelhttp.Labeler
			)
			if stateCounter != nil || successCounter != nil {
				ctx, labeler = LabelerFromContext(r.Context())
				r = r.WithContext(ctx)
			}

			// we need to detect the state change by ourselves, because the context does not allow us to hand in a context
			fromState := circuitBreaker.State()

			// clone the request and the body if wanted
			reqCopy, errCopy := copyRequest(r, o.copyReqBody)
			if errCopy != nil {
				l.Error("unable to copy request", log.FError(errCopy))
				return nil, errCopy
			} else if o.copyReqBody {
				// make sure the body is closed again - since it is a NopCloser it does not make a difference though
				defer reqCopy.Body.Close()
			}

			// call the next handler enclosed in the circuit breaker.
			resp, err := circuitBreaker.Execute(func() (interface{}, error) {

				resp, err := next(r)

				// clone the response and the body if wanted
				respCopy, errCopy := copyResponse(resp, o.copyRespBody)
				if errCopy != nil {
					l.Error("unable to copy response", log.FError(errCopy))
					return nil, errCopy
				} else if o.copyRespBody {
					// make sure the body is closed again - since it is a NopCloser it does not make a difference though
					defer respCopy.Body.Close()
				}

				return resp, o.isSuccessful(err, reqCopy, respCopy)
			})

			// detect and log a state change
			toState := circuitBreaker.State()
			if fromState != toState {
				l.Warn("state change occured",
					zap.String("from", fromState.String()),
					zap.String("to", toState.String()),
				)

				// recording the metric if desired
				if stateCounter != nil {
					attributes := append(
						labeler.Get(),
						attribute.String("state_change", fmt.Sprintf("%s -> %s", fromState.String(), toState.String())),
					)
					stateCounter.Add(ctx, 1, attributes...)
				}
			}

			// wrap the error in case it was produced because of the circuit breaker being (half-)open
			if errors.Is(gobreaker.ErrTooManyRequests, err) || errors.Is(gobreaker.ErrOpenState, err) {
				err = keelerrors.NewWrappedError(ErrCircuitBreaker, err)
			}

			if err != nil {
				if successCounter != nil {
					attributes := append(
						labeler.Get(),
						attribute.Bool("success", false),
					)
					successCounter.Add(ctx, 1, attributes...)
				}
				return nil, err
			}

			if successCounter != nil {
				attributes := append(
					labeler.Get(),
					attribute.Bool("success", true),
				)
				successCounter.Add(ctx, 1, attributes...)
			}

			return resp.(*http.Response), nil
		}
	}
}

// errNoBody is a sentinel error value used by failureToReadBody so we
// can detect that the lack of body was intentional.
var errNoBody = errors.New("sentinel error value")

// failureToReadBody is an io.ReadCloser that just returns errNoBody on
// Read. It's swapped in when we don't actually want to consume
// the body, but need a non-nil one, and want to distinguish the
// error from reading the dummy body.
type failureToReadBody struct{}

func (failureToReadBody) Read([]byte) (int, error) { return 0, errNoBody }
func (failureToReadBody) Close() error             { return nil }

// emptyBody is an instance of empty reader.
var emptyBody = io.NopCloser(strings.NewReader(""))

func copyRequest(req *http.Request, body bool) (*http.Request, error) {
	// we don't care about the context, since it is only used for the isSuccessful check
	out := req.Clone(context.Background())

	// duplicate the body
	if !body {
		// For content length of zero. Make sure the body is an empty
		// reader, instead of returning error through failureToReadBody{}.
		if req.ContentLength == 0 {
			out.Body = emptyBody
		} else {
			// if it is attempted to read from the body in isSuccessful we actually want the read to fail
			out.Body = failureToReadBody{}
		}

	} else if req.Body == nil {
		req.Body = nil
		out.Body = nil
	} else {
		var err error
		out.Body, req.Body, err = drainBody(req.Body)
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

func copyResponse(resp *http.Response, body bool) (*http.Response, error) {
	// we don't care about the context, since it is only used for the isSuccessful check
	out := new(http.Response)
	*out = *resp

	// duplicate the body
	if !body {
		// For content length of zero. Make sure the body is an empty
		// reader, instead of returning error through failureToReadBody{}.
		if resp.ContentLength == 0 {
			out.Body = emptyBody
		} else {
			// if it is attempted to read from the body in isSuccessful we actually want the read to fail
			out.Body = failureToReadBody{}
		}
	} else if resp.Body == nil {
		out.Body = nil
	} else {
		var err error
		out.Body, resp.Body, err = drainBody(resp.Body)
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

// copied from httputil
func drainBody(b io.ReadCloser) (r1, r2 io.ReadCloser, err error) {
	if b == nil || b == http.NoBody {
		// No copying needed. Preserve the magic sentinel meaning of NoBody.
		return http.NoBody, http.NoBody, nil
	}
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, b, err
	}
	if err = b.Close(); err != nil {
		return nil, b, err
	}
	return io.NopCloser(&buf), io.NopCloser(bytes.NewReader(buf.Bytes())), nil
}
