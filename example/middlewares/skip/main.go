package main

import (
	"net/http"

	"go.uber.org/zap"

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
	svs.HandleFunc("/skip", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	svr.AddServices(
		// with URI blacklist
		keel.NewServiceHTTP(l, "demo", "localhost:8080", svs,
			middleware.Skip(
				func(l *zap.Logger, name string, next http.Handler) http.Handler {
					return http.NotFoundHandler()
				},
				middleware.RequestURIBlacklistSkipper("/skip"),
			),
		),

		// with URI whitelist
		keel.NewServiceHTTP(l, "demo", ":8081", svs,
			middleware.Skip(
				func(l *zap.Logger, name string, next http.Handler) http.Handler {
					return http.NotFoundHandler()
				},
				middleware.RequestURIWhitelistSkipper("/"),
			),
		),
	)

	svr.Run()
}
