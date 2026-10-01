package roundtripware

import (
	"net/http"

	"go.uber.org/zap"
)

type (
	// Handler sends a request and returns its response, like
	// [http.RoundTripper.RoundTrip].
	Handler func(r *http.Request) (*http.Response, error)
	// RoundTripware wraps next with additional client-side request handling.
	RoundTripware func(l *zap.Logger, next Handler) Handler
)

// RoundTripper is an [http.RoundTripper] that runs requests through a chain
// of RoundTripwares before passing them to the parent RoundTripper.
type RoundTripper struct {
	http.RoundTripper
	handler Handler
}

// NewRoundTripper returns a RoundTripper that wraps parent with
// roundTripwares. They are applied in order, so the last one is the outermost
// and sees a request first.
func NewRoundTripper(l *zap.Logger, parent http.RoundTripper, roundTripwares ...RoundTripware) *RoundTripper {
	next := parent.RoundTrip
	for _, roundTripware := range roundTripwares {
		next = roundTripware(l, next)
	}

	return &RoundTripper{
		RoundTripper: parent,
		handler:      next,
	}
}

// RoundTrip sends req through the RoundTripware chain.
func (rt *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt.handler(req)
}
