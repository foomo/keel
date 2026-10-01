package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/foomo/go/options"
	"github.com/nats-io/nats-server/v2/server"
)

// DefaultEmbeddedServerURL is the client URL of an [EmbeddedServer] with the
// default host and port.
const DefaultEmbeddedServerURL = "nats://0.0.0.0:4222"

// EmbeddedServer is an in-process NATS server usable as a keel service.
// Create it with [NewEmbeddedServer].
type EmbeddedServer struct {
	server      *server.Server
	port        int
	host        string
	maxPending  int64
	clientURL   string
	natsOptions []options.Option[*server.Options]
}

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

// EmbeddedServerWithPort sets the listen port. Defaults to 4222.
func EmbeddedServerWithPort(v int) options.Option[*EmbeddedServer] {
	return func(o *EmbeddedServer) {
		o.port = v
	}
}

// EmbeddedServerWithMaxPending sets the maximum number of bytes buffered per
// client connection. Defaults to 64 MiB.
func EmbeddedServerWithMaxPending(v int64) options.Option[*EmbeddedServer] {
	return func(o *EmbeddedServer) {
		o.maxPending = v
	}
}

// EmbeddedServerWithHost sets the listen host. Defaults to "0.0.0.0".
func EmbeddedServerWithHost(v string) options.Option[*EmbeddedServer] {
	return func(o *EmbeddedServer) {
		o.host = v
	}
}

// EmbeddedServerWithNatsOptions adds options applied to the NATS server
// options after host, port and max pending have been set.
func EmbeddedServerWithNatsOptions(v ...options.Option[*server.Options]) options.Option[*EmbeddedServer] {
	return func(o *EmbeddedServer) {
		o.natsOptions = append(o.natsOptions, v...)
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

// NewEmbeddedServer creates, but does not start, an embedded NATS server
// configured by opts. Server logging and signal handling are disabled.
func NewEmbeddedServer(opts ...options.Option[*EmbeddedServer]) (*EmbeddedServer, error) {
	inst := &EmbeddedServer{
		port:       4222,
		host:       "0.0.0.0",
		maxPending: 64 << 20, // 64 MiB
	}

	options.Apply(inst, opts...)

	natsOpts := &server.Options{
		Host:       inst.host,
		Port:       inst.port,
		NoLog:      true,
		NoSigs:     true,
		MaxPending: inst.maxPending,
	}

	options.Apply(natsOpts, inst.natsOptions...)

	ns, err := server.NewServer(natsOpts)
	if err != nil {
		return nil, fmt.Errorf("embednats: new server: %w", err)
	}

	var u url.URL

	u.Scheme = "nats"
	u.Host = net.JoinHostPort(inst.host, fmt.Sprintf("%d", inst.port))

	return &EmbeddedServer{server: ns, clientURL: u.String()}, nil
}

// MustNewEmbeddedServer is like [NewEmbeddedServer] but panics on error.
func MustNewEmbeddedServer(opts ...options.Option[*EmbeddedServer]) *EmbeddedServer {
	s, err := NewEmbeddedServer(opts...)
	if err != nil {
		panic(err)
	}

	return s
}

// ------------------------------------------------------------------------------------------------
// ~ Getter
// ------------------------------------------------------------------------------------------------

// ClientURL returns the URL clients should dial.
func (s *EmbeddedServer) ClientURL() string {
	return s.clientURL
}

// Server returns the underlying NATS server.
func (s *EmbeddedServer) Server() *server.Server {
	return s.server
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

// Start starts the server and blocks until it shuts down. It returns an error
// if the server is not ready for connections within 5 seconds.
func (s *EmbeddedServer) Start(ctx context.Context) error {
	s.server.Start()

	if !s.server.ReadyForConnections(5 * time.Second) {
		s.server.Shutdown()
		return errors.New("nats server not ready")
	}

	s.server.WaitForShutdown()

	return nil
}

// Close shuts the server down and waits for the shutdown to complete. It
// returns an error wrapping ctx.Err() if ctx is done first.
func (s *EmbeddedServer) Close(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	done := make(chan struct{})

	go func() {
		s.server.Shutdown()
		s.server.WaitForShutdown()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("embednats: shutdown: %w", ctx.Err())
	}
}
