// Package httputils provides helpers for HTTP handlers: writing error
// responses with logging and telemetry, inspecting request host and client
// address, and hashing basic auth passwords.
//
// Each status-specific helper such as [NotFoundServerError] or
// [InternalServerError] delegates to [ServerError]:
//
//	if err != nil {
//		httputils.InternalServerError(l, w, r, err)
//		return
//	}
package httputils
