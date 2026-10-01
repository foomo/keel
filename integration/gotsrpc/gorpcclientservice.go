package gotsrpc

import (
	"context"
)

// GoRPCClientService runs a [GoRPCClient] for the lifetime of a service.
type GoRPCClientService struct {
	client GoRPCClient
	cancel context.CancelFunc
}

// NewGoRPCClientService returns a service wrapping client.
func NewGoRPCClientService(client GoRPCClient) *GoRPCClientService {
	return &GoRPCClientService{
		client: client,
	}
}

// Start starts the client and blocks until ctx is canceled or [GoRPCClientService.Close]
// is called, then stops the client and returns the context error.
func (s *GoRPCClientService) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.client.Start()
	<-ctx.Done()
	s.client.Stop()

	return ctx.Err()
}

// Close unblocks a running [GoRPCClientService.Start]. It always returns nil.
func (s *GoRPCClientService) Close(ctx context.Context) error {
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}

	return nil
}
