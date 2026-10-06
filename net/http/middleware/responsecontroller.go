package middleware

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/foomo/keel/log"
	keelhttp "github.com/foomo/keel/net/http"
)

// ResponseController middleware runs h against the request's
// http.ResponseController before the handler, to override per connection what
// the server was constructed with: SetReadDeadline / SetWriteDeadline (the zero
// time clears a deadline), EnableFullDuplex, Flush.
//
//	// SSE: no write deadline
//	middleware.ResponseController(func(rc *http.ResponseController) error {
//		return rc.SetWriteDeadline(time.Time{})
//	})
//
//	// large upload: read for as long as it takes, bound the body instead
//	middleware.ResponseController(func(rc *http.ResponseController) error {
//		return rc.SetReadDeadline(time.Now().Add(10 * time.Minute))
//	})
//
// An error from h is logged and the handler still runs — a ResponseWriter that
// does not support the operation must not fail the request.
//
// IdleTimeout is out of reach here: it governs the connection between requests,
// when no handler is running, so there is nothing to hang a middleware on. It
// stays a server value (see keelhttp.NewServer) and must exceed the idle timeout
// of whatever sits in front of the pod — GCLB 600s, nginx-ingress 75s, ALB 60s —
// so that side always closes first. If the server expires first, the balancer
// hands the next request to a socket the server is closing and returns a 502.
func ResponseController(h func(rc *http.ResponseController) error) keelhttp.Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if h != nil {
				if err := h(http.NewResponseController(w)); err != nil {
					log.WithError(l, err).Warn("failed to apply response controller")
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
