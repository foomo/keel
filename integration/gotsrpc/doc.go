// Package gotsrpc provides keel service wrappers for gotsrpc GoRPC clients
// and proxies.
//
// [GoRPCClientService] and [GoRPCProxyService] adapt a [GoRPCClient] or
// [GoRPCProxy] to the keel service lifecycle: the wrapped value is started
// when the service starts and stopped once the service is closed or its
// context is canceled.
package gotsrpc
