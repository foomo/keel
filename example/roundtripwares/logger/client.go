package main

import (
	"github.com/foomo/keel/log"
	"github.com/foomo/keel/net/http"
	"github.com/foomo/keel/net/http/roundtripware"
)

func client() {
	l := log.Logger()

	client := http.NewHTTPClient(
		http.HTTPClientWithRoundTripware(l,
			roundtripware.Logger(),
		),
	)

	var err error

	_, err = client.Get("http://localhost:8080") //nolint:all
	log.Must(l, err, "failed to retrieve response")

	_, err = client.Get("http://localhost:8080/404") //nolint:all
	log.Must(l, err, "failed to retrieve response")
}
