package main

import (
	"net/http"
	"time"

	"github.com/foomo/keel"
	"github.com/foomo/keel/log"
)

func main() {
	svr := keel.NewServer(
		keel.WithHTTPZapService(true),
	)

	// obtain the logger
	l := svr.Logger()

	// alternatively you can always use
	// l := log.Logger()

	// measure tome time
	fDurationFn := log.FDurationFn(l)
	time.Sleep(200 * time.Millisecond)
	l.Info("measured some time", fDurationFn())

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.WithHTTPRequest(l, r).Info("handled request")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	svr.AddService(
		keel.NewServiceHTTP(l, "demo", "localhost:8080", svs),
	)

	svr.Run()
}
