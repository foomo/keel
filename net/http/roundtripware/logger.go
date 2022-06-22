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
			statusCode := -1
			contentLength := int64(-1)
			if resp != nil {
				statusCode = resp.StatusCode
				contentLength = resp.ContentLength
			}
			log.WithHTTPRequestOut(l, req).Info(msg,
				log.FDuration(keeltime.Now().Sub(start)),
				log.FHTTPStatusCode(statusCode),
				log.FHTTPRequestContentLength(contentLength),
			)
			return resp, err
		}
	}
}
