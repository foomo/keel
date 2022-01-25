package roundtripware

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/foomo/keel/log"
	keeltime "github.com/foomo/keel/time"
)

// Logger returns a RoundTripware which logs all requests
func Logger() RoundTripware {
	msg := "sent request"
	return func(l *zap.Logger, next Handler) Handler {
		return func(req *http.Request) (*http.Response, error) {
			start := keeltime.Now()
			resp, err := next(req)
			log.WithHTTPRequestOut(l, req).Info(msg,
				log.FDuration(keeltime.Now().Sub(start)),
				log.FHTTPStatusCode(resp.StatusCode),
				log.FHTTPRequestContentLength(resp.ContentLength),
			)
			return resp, err
		}
	}
}
