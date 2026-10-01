package keel

import (
	"context"
)

// Service is a unit of work run by a [Server]. Start is called with the server
// context and is expected to block until the service has stopped; a returned
// error other than [net/http.ErrServerClosed] fails the server. Services
// typically also implement a closer interface (see [IsCloser]) so they are
// stopped on graceful shutdown.
type Service interface {
	Start(ctx context.Context) error
}
