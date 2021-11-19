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
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	token := "some-random-token"

	// custom token provider
	tokenProvider := middleware.CookieTokenProvider("keel-token")

	svr.AddService(
		keel.NewServiceHTTP(l, "demo", ":8080", svs,
			middleware.TokenAuth(
				token,
				middleware.TokenAuthWithTokenProvider(tokenProvider),
			),
		),
	)

	svr.Run()
}
