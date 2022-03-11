package main

import (
	"net/http"

	"github.com/foomo/keel"
	keelhttp "github.com/foomo/keel/net/http"
	"github.com/foomo/keel/net/http/middleware"
)

func main() {
	svr := keel.NewServer()

	// get logger
	l := svr.Logger()

	// custom request id generator
	requestIDGenerator := func() string {
		return "my-custom-request-id"
	}

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(r.Header.Get(keelhttp.HeaderXRequestID)))
	})

	svr.AddService(
		keel.NewServiceHTTP(l, "demo", "localhost:8080", svs,
			middleware.RequestID(
				middleware.RequestIDWithSetResponseHeader(true),
				middleware.RequestIDWithGenerator(requestIDGenerator),
			),
		),
	)

	svr.Run()
}
