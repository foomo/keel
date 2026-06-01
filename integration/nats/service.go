package nats

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

const DefaultServiceURL = "nats://0.0.0.0:4222"

type Service struct {
	server     *server.Server
	port       int
	host       string
	maxPending int64
	clientURL  string
}

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func ServiceWithPort(v int) options.Option[*Service] {
	return func(o *Service) {
		o.port = v
	}
}

func ServiceWithMaxPending(v int64) options.Option[*Service] {
	return func(o *Service) {
		o.maxPending = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func NewService(opts ...options.Option[*Service]) (*Service, error) {
	inst := &Service{
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

	ns, err := server.NewServer(natsOpts)
	if err != nil {
		return nil, fmt.Errorf("embednats: new server: %w", err)
	}

	var u url.URL

	u.Scheme = "nats"
	u.Host = net.JoinHostPort(inst.host, fmt.Sprintf("%d", inst.port))

	return &Service{server: ns, clientURL: u.String()}, nil
}

func MustNewService(opts ...options.Option[*Service]) *Service {
	s, err := NewService(opts...)
	if err != nil {
		panic(err)
	}

	return s
}

// ------------------------------------------------------------------------------------------------
// ~ Getter
// ------------------------------------------------------------------------------------------------

// ClientURL returns the URL clients should dial.
func (s *Service) ClientURL() string {
	return s.clientURL
}

func (s *Service) Server() *server.Server {
	return s.server
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (s *Service) Start(ctx context.Context) error {
	s.server.Start()

	if !s.server.ReadyForConnections(5 * time.Second) {
		s.server.Shutdown()
		return errors.New("nats server not ready")
	}

	return nil
}

func (s *Service) Close(ctx context.Context) error {
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
