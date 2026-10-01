// Package context provides helpers to store and retrieve request-scoped
// identifiers in a [context.Context]: the referer, request ID, session ID and
// tracking ID.
//
// Middleware in [github.com/foomo/keel/net/http/middleware] stores the request,
// session and tracking IDs of incoming requests, and round-trippers in
// [github.com/foomo/keel/net/http/roundtripware] forward these values on
// outgoing requests.
package context
