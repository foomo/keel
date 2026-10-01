package roundtripware

import (
	"net/http"
)

// RoundTripperFunc adapts a [Handler] to [http.RoundTripper].
type RoundTripperFunc Handler

// RoundTrip calls fn(r).
func (fn RoundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}
