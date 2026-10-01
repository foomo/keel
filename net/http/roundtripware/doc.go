// Package roundtripware provides HTTP client middlewares, called
// RoundTripwares, that wrap an [net/http.RoundTripper].
//
// A [RoundTripware] wraps a [Handler] the same way a server middleware wraps
// an [net/http.Handler]. [NewRoundTripper] chains RoundTripwares around a
// parent transport:
//
//	client := &http.Client{
//		Transport: roundtripware.NewRoundTripper(l, http.DefaultTransport,
//			roundtripware.RequestID(),
//			roundtripware.Logger(),
//		),
//	}
//
// # RoundTripwares
//
//   - Context propagation: [RequestID], [SessionID], [TrackingID] and
//     [Referer] copy values from the request context into request headers.
//   - Resilience: [CircuitBreaker], [Retry] and [Recover].
//   - Observability: [Logger] and the deprecated [Metric].
//   - Transport: [GZip].
//   - Debugging: [Dump], [DumpRequest] and [DumpResponse].
package roundtripware
