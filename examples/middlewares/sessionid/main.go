package main

import (
	"net/http"
	"time"

	"github.com/foomo/keel"
	keelhttp "github.com/foomo/keel/net/http"
	"github.com/foomo/keel/net/http/cookie"
	"github.com/foomo/keel/net/http/middleware"
)

func main() {
	svr := keel.NewServer()

	// get logger
	l := svr.Logger()

	// custom cookie time provider
	timeProvider := cookie.NewTimeProvider(
		cookie.TimeProviderWithOffset(-time.Millisecond * 500),
	)

	// custom cookie domain provider
	domainProvider := cookie.NewDomainProvider(
		cookie.DomainProviderWithDomains("foo.com", "*.bar.com"),
		cookie.DomainProviderWithMappings(map[string]string{"source.com": "target.com"}),
	)

	// custom cookie provider
	sessionCookie := cookie.New(
		"session",
		cookie.WithExpires(time.Hour*60),
		cookie.WithTimeProvider(timeProvider),
		cookie.WithDomainProvider(domainProvider),
	)

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(r.Header.Get(keelhttp.HeaderXSessionID)))
		_, _ = w.Write([]byte(middleware.SessionIDFromContext(r.Context())))
	})

	svr.AddService(
		keel.NewServiceHTTP(l, "demo", "localhost:8080", svs,
			middleware.SessionID(
				// automatically set cookie if not exists
				middleware.SessionIDWithSetCookie(true),
				// define a custom cookie provider
				middleware.SessionIDWithCookie(sessionCookie),
			),
		),
	)

	svr.Run()
}
