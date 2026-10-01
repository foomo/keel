package gotsrpc

import (
	"context"
)

// GoRPCProxyService runs a [GoRPCProxy] for the lifetime of a service.
type GoRPCProxyService struct {
	proxy  GoRPCProxy
	cancel context.CancelFunc
}

// NewGoRPCProxyService returns a service wrapping proxy.
func NewGoRPCProxyService(proxy GoRPCProxy) *GoRPCProxyService {
	return &GoRPCProxyService{
		proxy: proxy,
	}
}

// Start starts the proxy and blocks until ctx is canceled or [GoRPCProxyService.Close]
// is called, then stops the proxy and returns the context error. If the proxy
// fails to start, its error is returned immediately.
func (s *GoRPCProxyService) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)

	s.cancel = cancel
	if err := s.proxy.Start(); err != nil {
		return err
	}

	<-ctx.Done()
	s.proxy.Stop()

	return ctx.Err()
}

// Close unblocks a running [GoRPCProxyService.Start]. It always returns nil.
func (s *GoRPCProxyService) Close(ctx context.Context) error {
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}

	return nil
}
