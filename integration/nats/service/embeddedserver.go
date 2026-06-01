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

const DefaultEmbeddedServerURL = "nats://0.0.0.0:4222"

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

func EmbeddedServerWithPort(v int) options.Option[*EmbeddedServer] {
	return func(o *EmbeddedServer) {
		o.port = v
	}
}

func EmbeddedServerWithMaxPending(v int64) options.Option[*EmbeddedServer] {
	return func(o *EmbeddedServer) {
		o.maxPending = v
	}
}

func EmbeddedServerWithHost(v string) options.Option[*EmbeddedServer] {
	return func(o *EmbeddedServer) {
		o.host = v
	}
}

func EmbeddedServerWithNatsOptions(v ...options.Option[*server.Options]) options.Option[*EmbeddedServer] {
	return func(o *EmbeddedServer) {
		o.natsOptions = append(o.natsOptions, v...)
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

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

func (s *EmbeddedServer) Server() *server.Server {
	return s.server
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (s *EmbeddedServer) Start(ctx context.Context) error {
	s.server.Start()

	if !s.server.ReadyForConnections(5 * time.Second) {
		s.server.Shutdown()
		return errors.New("nats server not ready")
	}

	s.server.WaitForShutdown()

	return nil
}

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
