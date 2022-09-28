package main

import (
	"fmt"
	"time"

	keellog "github.com/foomo/keel/log"
	keelhttp "github.com/foomo/keel/net/http"
	"github.com/foomo/keel/net/http/roundtripware"
)

func client() {
	l := keellog.Logger()

	client := keelhttp.NewHTTPClient(
		keelhttp.HTTPClientWithRoundTripware(l,
			roundtripware.Retry(
				roundtripware.RetryWithAttempts(12),
				roundtripware.RetryWithMaxDelay(time.Second),
			),
		))

	_, err := client.Get("http://localhost:8080/404") //nolint:all
	keellog.Must(l, err, "failed to retrieve response")

	fmt.Printf("Repetition process is finished with: %v\n", err) //nolint:forbidigo
}
