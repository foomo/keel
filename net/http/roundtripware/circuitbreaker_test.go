package roundtripware_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	keelhttp "github.com/foomo/keel/net/http"
	"github.com/foomo/keel/net/http/roundtripware"
	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

var cbSettings = &roundtripware.CircuitBreakerSettings{
	// Name is the name of the CircuitBreaker.
	Name: "circuit breaker test",
	// MaxRequests is the maximum number of requests allowed to pass through
	// when the CircuitBreaker is half-open.
	// If MaxRequests is 0, the CircuitBreaker allows only 1 request.
	MaxRequests: 1,
	// Interval is the cyclic period of the closed state
	// for the CircuitBreaker to clear the internal Counts.
	// If Interval is less than or equal to 0, the CircuitBreaker doesn't clear internal Counts during the closed state.
	// Interval : time.Second, - ignore interval for most of the tests to keep them more reliable
	// Timeout is the period of the open state,
	// after which the state of the CircuitBreaker becomes half-open.
	// If Timeout is less than or equal to 0, the timeout value of the CircuitBreaker is set to 60 seconds.
	Timeout: time.Millisecond * 100,
	// ReadyToTrip is called with a copy of Counts whenever a request fails in the closed state.
	// If ReadyToTrip returns true, the CircuitBreaker will be placed into the open state.
	// If ReadyToTrip is nil, default ReadyToTrip is used.
	// Default ReadyToTrip returns true when the number of consecutive failures is more than 5.
	ReadyToTrip: func(counts gobreaker.Counts) bool {
		return counts.ConsecutiveFailures > 3
	},
}

func TestCircuitBreaker(t *testing.T) {
	// create logger
	l := zaptest.NewLogger(t)

	// create http server with handler
	i := 0
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i++
		if i < 5 {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	// create http client
	client := keelhttp.NewHTTPClient(
		keelhttp.HTTPClientWithRoundTripware(l,
			roundtripware.CircuitBreaker(cbSettings,
				roundtripware.CircuitBreakerWithIsSuccessful(
					func(err error, req *http.Request, resp *http.Response) error {
						if resp.StatusCode >= http.StatusInternalServerError {
							return errors.New("invalid status code")
						}
						return nil
					}, true, true,
				),
			),
		),
	)

	// do requests to trigger the circuit breaker
	for i := 0; i <= 3; i++ {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, svr.URL, nil)
		require.NoError(t, err)
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
		}
		require.NotErrorIs(t, err, roundtripware.ErrCircuitBreaker)
	}

	// circuit breaker should now be triggered
	// this should result in a circuit breaker error
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, svr.URL, nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
	}
	require.ErrorIs(t, err, roundtripware.ErrCircuitBreaker)

	// wait for the timeout to hit
	time.Sleep(time.Millisecond * 100)

	req, err = http.NewRequestWithContext(context.Background(), http.MethodGet, svr.URL, nil)
	require.NoError(t, err)
	resp, err = client.Do(req)
	if err == nil {
		defer resp.Body.Close()
	}
	require.NoError(t, err)
}

func TestCircuitBreakerCopyBodies(t *testing.T) {
	requestData := "some request"
	responseData := "some response"

	// create logger
	l := zaptest.NewLogger(t)

	// create http server with handler
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.Equal(t, string(data), requestData)
		_, err = w.Write([]byte(responseData))
		if err != nil {
			panic(err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	// create http client
	client := keelhttp.NewHTTPClient(
		keelhttp.HTTPClientWithRoundTripware(l,
			roundtripware.CircuitBreaker(cbSettings,
				roundtripware.CircuitBreakerWithIsSuccessful(
					func(err error, req *http.Request, resp *http.Response) error {
						// read the bodies
						_, errRead := io.ReadAll(req.Body)
						require.NoError(t, errRead)

						_, errRead = io.ReadAll(resp.Body)
						require.NoError(t, errRead)

						// also try to close one of the bodies (should also be handled by the RoundTripware)
						req.Body.Close()

						return err
					}, true, true,
				),
			),
		),
	)

	// do requests to trigger the circuit breaker
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, svr.URL, strings.NewReader(requestData))
	require.NoError(t, err)
	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
	}
	require.NoError(t, err)
	// make sure the correct data is returned
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, string(data), responseData)
}

func TestCircuitBreakerReadFromNotCopiedBodies(t *testing.T) {
	requestData := "some request"
	responseData := "some response"

	// create logger
	l := zaptest.NewLogger(t)

	// create http server with handler
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.Equal(t, string(data), requestData)
		_, err = w.Write([]byte(responseData))
		if err != nil {
			panic(err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	// create http client
	client := keelhttp.NewHTTPClient(
		keelhttp.HTTPClientWithRoundTripware(l,
			roundtripware.CircuitBreaker(cbSettings,
				roundtripware.CircuitBreakerWithIsSuccessful(
					func(err error, req *http.Request, resp *http.Response) error {
						// read the bodies
						_, errRead := io.ReadAll(req.Body)
						if errRead != nil {
							return errRead
						}

						return err
					}, false, true,
				),
			),
		),
	)

	// do requests to trigger the circuit breaker
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, svr.URL, strings.NewReader(requestData))
	require.NoError(t, err)
	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
	}
	require.Error(t, err)

	// same thing for the response
	client = keelhttp.NewHTTPClient(
		keelhttp.HTTPClientWithRoundTripware(l,
			roundtripware.CircuitBreaker(cbSettings,
				roundtripware.CircuitBreakerWithIsSuccessful(
					func(err error, req *http.Request, resp *http.Response) error {
						// read the bodies
						_, errRead := io.ReadAll(resp.Body)
						if errRead != nil {
							return errRead
						}

						return err
					}, true, false,
				),
			),
		),
	)

	// do requests to trigger the circuit breaker
	req, err = http.NewRequestWithContext(context.Background(), http.MethodGet, svr.URL, strings.NewReader(requestData))
	require.NoError(t, err)
	resp, err = client.Do(req)
	if err == nil {
		defer resp.Body.Close()
	}
	require.Error(t, err)
}

func TestCircuitBreakerInterval(t *testing.T) {
	// create logger
	l := zaptest.NewLogger(t)

	// create http server with handler
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// always return an invalid status code
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	// create http client
	client := keelhttp.NewHTTPClient(
		keelhttp.HTTPClientWithRoundTripware(l,
			roundtripware.CircuitBreaker(&roundtripware.CircuitBreakerSettings{
				Name:        cbSettings.Name,
				MaxRequests: cbSettings.MaxRequests,
				Interval:    time.Second,
				Timeout:     time.Millisecond * 100,
				ReadyToTrip: func(counts gobreaker.Counts) bool {
					return counts.ConsecutiveFailures > 3
				},
			},
				roundtripware.CircuitBreakerWithIsSuccessful(
					func(err error, req *http.Request, resp *http.Response) error {
						if resp.StatusCode >= http.StatusInternalServerError {
							return errors.New("invalid status code")
						}
						return nil
					}, true, true,
				),
			),
		),
	)

	// send exactly 3 requests (lower than the maximum amount of allowed consecutive failures)
	for i := 0; i < 3; i++ {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, svr.URL, nil)
		require.NoError(t, err)
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
		}
		require.NotErrorIs(t, err, roundtripware.ErrCircuitBreaker)
	}

	// wait for the interval time
	time.Sleep(time.Second)

	// now we should be able to send 3 more requests without triggering the circuit breaker (last request should finally
	// trigger the circuit breaker, but the error will not yet be a circuitbreaker error)
	for i := 0; i <= 3; i++ {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, svr.URL, nil)
		require.NoError(t, err)
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
		}
		require.NotErrorIs(t, err, roundtripware.ErrCircuitBreaker)
	}

	// this request should now finally trigger the circuit breaker
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, svr.URL, nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
	}
	require.ErrorIs(t, err, roundtripware.ErrCircuitBreaker)
}
