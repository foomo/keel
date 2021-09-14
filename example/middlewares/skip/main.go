package main

import (
	"net/http"

	"github.com/foomo/keel"
	"github.com/foomo/keel/net/http/middleware"
)

func main() {
	svr := keel.NewServer()

	// get logger
	l := svr.Logger()

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		panic("handled")
	})
	svs.HandleFunc("/skip", func(w http.ResponseWriter, r *http.Request) {
		panic("unhandled")
	})

	svr.AddServices(
		// with URI blacklist
		keel.NewServiceHTTP(l, "demo", ":8080", svs,
			middleware.Skip(
				middleware.Recover(),
				middleware.RequestURIBlacklistSkipper("/skip"),
			),
		),

		// with URI whitelist
		keel.NewServiceHTTP(l, "demo", ":8081", svs,
			middleware.Skip(
				middleware.Recover(),
				middleware.RequestURIWhitelistSkipper("/"),
			),
		),
	)

	svr.Run()
}
