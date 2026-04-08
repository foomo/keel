package http

import (
	"context"
	"io"
	"net/http"

	"github.com/foomo/keel/log"
	"github.com/foomo/keel/messaging"
	encodingx "github.com/foomo/keel/pkg/encoding"
	"go.uber.org/zap"
)

const (
	// DefaultMaxBodySize is the maximum request body size the subscriber will
	// read. Override per-subscriber via the WithMaxBodySize option.
	DefaultMaxBodySize int64 = 1 << 20 // 1 MiB
)

// Subscriber decodes incoming POST requests and dispatches them to a Handler.
// It owns an *http.ServeMux that routes are registered on; the mux is exposed
// via Handler() so the caller can hand it to any http.Server — including keel's
// service.NewHTTP.
//
// Subscriber intentionally does not start its own net.Listener. Lifecycle
// (start, graceful shutdown) belongs to the server layer above it.
type Subscriber[T any] struct {
	serializer  encodingx.Codec[T]
	mux         *http.ServeMux
	maxBodySize int64
}

// SubscriberOption configures a Subscriber.
type SubscriberOption func(*subscriberConfig)

type subscriberConfig struct {
	maxBodySize int64
}

// WithMaxBodySize sets the maximum request body size the subscriber will read.
// Requests exceeding this limit receive 413 Request Entity Too Large.
func WithMaxBodySize(n int64) SubscriberOption {
	return func(c *subscriberConfig) { c.maxBodySize = n }
}

// NewSubscriber creates an HTTP subscriber. Call Subscribe to register subjects,
// then pass Mux() to service.NewHTTP (keel) or http.ListenAndServe.
func NewSubscriber[T any](serializer encodingx.Codec[T], opts ...SubscriberOption) *Subscriber[T] {
	cfg := &subscriberConfig{maxBodySize: DefaultMaxBodySize}
	for _, o := range opts {
		o(cfg)
	}
	return &Subscriber[T]{serializer: serializer, mux: http.NewServeMux(), maxBodySize: cfg.maxBodySize}
}

func (s *Subscriber[T]) Mux() *http.ServeMux {
	return s.mux
}

// Subscribe registers handler for POST /{subject} on the mux.
// Multiple subjects can be registered before the server starts.
// ctx is used as the base context for all handler invocations.
//
// Responses: 204 on success, 400 on decode failure, 405 on wrong method,
// 413 if body exceeds max size, 500 if the handler returns an error.
func (s *Subscriber[T]) Subscribe(ctx context.Context, subject string, handler messaging.Handler[T]) error {
	s.mux.Handle("/"+subject, s.Handler(subject, handler))
	// Subscribe is non-blocking here — the server is started externally.
	// Block until ctx is cancelled so the caller's goroutine stays alive.
	<-ctx.Done()
	return nil
}

// Handler returns an http.HandlerFunc that dispatches incoming POST requests
// to the given handler.
func (s *Subscriber[T]) Handler(subject string, handler messaging.Handler[T]) http.HandlerFunc {
	l := log.Logger()

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		defer r.Body.Close()

		body, err := io.ReadAll(io.LimitReader(r.Body, s.maxBodySize+1))
		if err != nil {
			l.Warn("http: read body failed", zap.String("subject", subject), zap.Error(err))
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if int64(len(body)) > s.maxBodySize {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return
		}

		var v T
		if err := s.serializer.Decode(body, &v); err != nil {
			l.Warn("http: decode failed", zap.String("subject", subject), zap.Error(err))
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		msg := messaging.Message[T]{Subject: subject, Payload: v}
		err = messaging.RecordProcess(r.Context(), subject, system, func(hCtx context.Context) error {
			return handler(hCtx, msg)
		})
		if err != nil {
			l.Error("http: handler failed", zap.String("subject", subject), zap.Error(err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// Close is a no-op; shutdown is handled by keel / the outer http.Server.
func (s *Subscriber[T]) Close() error { return nil }
