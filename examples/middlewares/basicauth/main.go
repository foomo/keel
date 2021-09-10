package main

import (
	"net/http"

	"github.com/foomo/keel"
	"github.com/foomo/keel/log"
	"github.com/foomo/keel/net/http/middleware"
	httputils "github.com/foomo/keel/utils/net/http"
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

	username := "demo"

	// create password hash
	passwordHash, err := httputils.HashBasicAuthPassword([]byte("demo"))
	log.Must(l, err, "failed to hash password")

	svr.AddService(
		keel.NewServiceHTTP(l, "demo", ":8080", svs,
			middleware.BasicAuth(
				username,
				passwordHash,
				middleware.BasicAuthWithRealm("demo"),
			),
		),
	)

	svr.Run()
}
