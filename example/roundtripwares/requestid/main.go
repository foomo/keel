package main

import (
	"net/http"

	"github.com/foomo/keel"
	keelhttp "github.com/foomo/keel/net/http"
	"github.com/foomo/keel/net/http/middleware"
	"github.com/foomo/keel/net/http/roundtripware"
	httputils "github.com/foomo/keel/utils/net/http"
)

func main() {
	svr := keel.NewServer()

	// get logger
	l := svr.Logger()

	httpClient := keelhttp.NewHTTPClient(
		keelhttp.HTTPClientWithRoundTripware(l,
			roundtripware.RequestID(),
		),
	)

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// send internal http request
		if req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, "http://localhost:8080/internal", nil); err != nil {
			httputils.InternalServerError(l, w, r, err)
			return
		} else if resp, err := httpClient.Do(req); err != nil {
			httputils.InternalServerError(l, w, r, err)
			return
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(r.Header.Get(keelhttp.HeaderXRequestID) + " - " + resp.Header.Get(keelhttp.HeaderXRequestID)))
		}
	})
	svs.HandleFunc("/internal", func(w http.ResponseWriter, r *http.Request) {
		l.Info("internal: " + r.Header.Get(keelhttp.HeaderXRequestID))
		w.WriteHeader(http.StatusOK)
	})

	svr.AddService(
		keel.NewServiceHTTP(l, "demo", "localhost:8080", svs,
			middleware.RequestID(),
		),
	)

	svr.Run()
}
