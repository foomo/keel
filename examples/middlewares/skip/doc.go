// Skip is an example service demonstrating the skip middleware with request URI
// blacklist and whitelist skippers.
//
// Run it with `go run ./examples/middlewares/skip`, then compare
// `curl localhost:8080/skip` with `curl localhost:8081/skip`.
package main
