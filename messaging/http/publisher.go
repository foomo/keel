package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/foomo/keel/messaging"
	encodingx "github.com/foomo/keel/pkg/encoding"
)

// Publisher POSTs encoded messages to a base URL.
// Subject is appended to BaseURL as the path: POST {BaseURL}/{subject}
type Publisher[T any] struct {
	baseURL    string
	serializer encodingx.Codec[T]
	httpClient *http.Client
	// ContentType is sent as the Content-Type header. Defaults to
	// "application/json" if empty.
	ContentType string
}

// NewPublisher creates an HTTP publisher.
// baseURL is the target service root, e.g. "https://orders.internal".
// An optional *http.Client may be provided; if nil the default client is used.
func NewPublisher[T any](baseURL string, serializer encodingx.Codec[T], client *http.Client) *Publisher[T] {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &Publisher[T]{baseURL: baseURL, serializer: serializer, httpClient: client}
}

// Publish encodes v and POSTs it to {baseURL}/{subject}.
// A non-2xx response is treated as an error.
func (p *Publisher[T]) Publish(ctx context.Context, subject string, v T) error {
	return messaging.RecordPublish(ctx, subject, system, func(ctx context.Context) error {
		return p.post(ctx, subject, v)
	})
}

func (p *Publisher[T]) post(ctx context.Context, subject string, v T) error {
	b, err := p.serializer.Encode(v)
	if err != nil {
		return fmt.Errorf("http publisher encode: %w", err)
	}

	ct := p.ContentType
	if ct == "" {
		ct = "application/json"
	}

	url := p.baseURL + "/" + subject
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("http publisher build request: %w", err)
	}
	req.Header.Set("Content-Type", ct)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http publisher send: %w", err)
	}
	defer resp.Body.Close()
	// drain body so the connection can be reused (limit to 1 MiB)
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("http publisher: server returned %d for %s", resp.StatusCode, url)
	}
	return nil
}

func (p *Publisher[T]) Close() error { return nil }
