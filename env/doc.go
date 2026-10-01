// Package env provides typed accessors for environment variables with
// fallback defaults.
//
// Get* functions return the parsed value or a fallback; MustGet* functions
// panic when the variable is not set. Every accessed key is recorded with its
// type and default, which can be inspected through [RequiredKeys],
// [Defaults], [Types] and [TypeOf]. All functions are safe for concurrent
// use.
//
//	port := env.GetInt("PORT", 8080)
//	token := env.MustGet("API_TOKEN")
package env
