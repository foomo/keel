package gotsrpc

// GoRPCProxy is a gotsrpc GoRPC proxy that can be started and stopped.
type GoRPCProxy interface {
	Start() error
	Stop()
}
