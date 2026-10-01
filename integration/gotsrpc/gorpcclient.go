package gotsrpc

// GoRPCClient is a gotsrpc GoRPC client that can be started and stopped.
type GoRPCClient interface {
	Start()
	Stop()
}
