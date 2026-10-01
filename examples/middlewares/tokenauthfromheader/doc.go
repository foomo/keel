// Tokenauthfromheader is an example service demonstrating the token auth middleware
// reading the token from a custom header.
//
// Run it with `go run ./examples/middlewares/tokenauthfromheader`, then try
// `curl -H 'X-Authorization: Custom some-random-token' localhost:8080`.
package main
