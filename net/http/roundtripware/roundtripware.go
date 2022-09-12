package roundtripware

import (
	"net/http"

	"go.uber.org/zap"
)

type (
	Handler       func(r *http.Request) (*http.Response, error)
	RoundTripware func(l *zap.Logger, next Handler) Handler
)

type RoundTripper struct {
	http.RoundTripper
	handler Handler
}

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

func (rt *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt.handler(req)
}
