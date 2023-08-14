package main

import (
	"net/http"

	"github.com/foomo/keel"
	"github.com/foomo/keel/log"
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
			roundtripware.SessionID(),
			roundtripware.TrackingID(),
		),
	)

	// create demo service
	svs := http.NewServeMux()

	// send internal http request
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// send request
		if req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, "http://localhost:8080/internal", nil); err != nil {
			httputils.InternalServerError(l, w, r, err)
			return
		} else if resp, err := httpClient.Do(req); err != nil {
			httputils.InternalServerError(l, w, r, err)
			return
		} else {
			defer resp.Body.Close()
			w.WriteHeader(http.StatusOK)
			log.WithHTTPRequest(l, r).Info("handled request")
			_, _ = w.Write([]byte(r.Header.Get(keelhttp.HeaderXRequestID) + " - " + resp.Header.Get(keelhttp.HeaderXRequestID)))
			log.WithHTTPRequestOut(l, req).Info("sent internal request")
		}
	})
	// handle internal http request
	svs.HandleFunc("/internal", func(w http.ResponseWriter, r *http.Request) {
		log.WithHTTPRequest(l, r).Info("handled internal request")
		w.WriteHeader(http.StatusOK)
	})

	svr.AddService(
		keel.NewServiceHTTP(l, "demo", "localhost:8080", svs,
			// add middleware
			middleware.RequestID(),
			// add middleware
			middleware.SessionID(),
			// add middleware
			middleware.TrackingID(),
		),
	)

	svr.Run()
}
