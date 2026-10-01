// Package keeltest provides helpers for testing keel services.
//
// [Server] starts and closes [Service] instances such as [ServiceHTTP]
// within a test, [HTTPClient] sends requests against them, and [Inline]
// implements inline snapshot assertions that store expected values as
// trailing comments in the test source:
//
//	svr := keeltest.NewServer(t)
//	svr.AddService(keeltest.NewServiceHTTP(svr.Logger(), "demo", handler))
//	svr.Start()
//
//	body, _, err := keeltest.NewHTTPClient(
//		keeltest.HTTPClientWithBaseURL(svr.GetService("demo").URL()),
//	).Get(t.Context(), "/")
//	require.NoError(t, err)
//	keelassert.InlineEqual(t, string(body)) // INLINE: ok
//
// The packages [github.com/foomo/keel/keeltest/assert] and
// [github.com/foomo/keel/keeltest/require] provide the inline assertions.
package keeltest
