package roundtripware

import (
	"net/http"
)

type RoundTripperFunc Handler

func (fn RoundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}
