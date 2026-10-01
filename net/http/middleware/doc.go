// Package middleware provides HTTP server middlewares for keel services.
//
// Every middleware is a [github.com/foomo/keel/net/http.Middleware] and is
// typically passed to an HTTP service or combined with
// [github.com/foomo/keel/net/http.Compose]:
//
//	svs := service.NewHTTP(l, "api", ":8080", handler,
//		middleware.Telemetry(),
//		middleware.Logger(),
//		middleware.Recover(),
//	)
//
// # Options
//
// Configurable middlewares follow the same pattern: Xxx(opts ...XxxOption)
// applies functional options to the defaults returned by GetDefaultXxxOptions
// and delegates to XxxWithOptions, which takes a fully populated XxxOptions.
//
// # Middlewares
//
//   - Authentication: [BasicAuth], [TokenAuth], [JWT] and [RBAC], with the
//     token sources [HeaderTokenProvider] and [CookieTokenProvider].
//   - Request context: [RequestID], [SessionID], [TrackingID] and [Referer].
//   - Observability: [Telemetry], [Logger] and [ResponseTime].
//   - Transport: [CORS], [GZip], [MaxRequestBodySize] and
//     [ResponseController].
//   - Response headers: [ServerHeader] and [PoweredByHeader].
//   - Safety: [Recover].
//
// [Skip] bypasses a middleware for requests matched by a [Skipper], such as
// [RequestURIBlacklistSkipper] or [RequestURIWhitelistSkipper].
package middleware
