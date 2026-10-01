// Package cookie provides a declarative helper to read, write and delete an
// HTTP cookie with consistent attributes.
//
// A [Cookie] is created with [New] and functional [Option] values. Its domain
// is resolved per request by a [DomainProvider], which can restrict cookies to
// a set of allowed domains, and its expiry is computed from a [TimeProvider]:
//
//	c := cookie.New("session",
//		cookie.WithExpires(24*time.Hour),
//		cookie.WithDomainProvider(cookie.NewDomainProvider(
//			cookie.DomainProviderWithDomains("*.example.com"),
//		)),
//	)
//	if _, err := c.Set(w, r, value); err != nil {
//		// handle error
//	}
package cookie
