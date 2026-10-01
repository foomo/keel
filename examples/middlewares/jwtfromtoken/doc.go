// Jwtfromtoken is an example service demonstrating the JWT middleware reading a
// required token from a custom header.
//
// Run it with `go run ./examples/middlewares/jwtfromtoken`, fetch a token with
// `curl localhost:8080/token`, then try
// `curl -H 'X-Authorization: Custom <token>' localhost:8080`.
package main
