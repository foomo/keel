// Package http provides HTTP client and server constructors tuned for
// services running on Kubernetes, together with middleware composition and
// common header name constants.
//
// # Clients
//
// [NewInternalHTTPClient] targets other workloads in the same cluster and
// [NewExternalHTTPClient] targets third-party APIs on the public internet.
// Both are customized with [HTTPClientOption] values; transport options must
// be applied before wrapping options such as [HTTPClientWithTelemetry]:
//
//	client := http.NewInternalHTTPClient(
//		http.HTTPClientWithMaxConnsPerHost(64),
//		http.HTTPClientWithTelemetry(),
//	)
//
// Create one client per upstream and reuse it for the lifetime of the process.
//
// # Servers
//
// [NewServer] returns a [net/http.Server] with production timeouts whose
// handler is wrapped with [Middleware] values via [Compose].
package http
