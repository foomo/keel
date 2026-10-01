// Package jwt provides helpers for signing and verifying RS256 JSON Web
// Tokens on top of github.com/golang-jwt/jwt/v5, including support for key
// rotation via deprecated verification keys.
//
// A [Key] pairs an RSA public key with an optional private key and an
// identifier that is written to and read from the token's "kid" header.
// [JWT] signs tokens with its current key and verifies them using
// [DefaultKeyFunc] unless a custom key function is supplied:
//
//	key, deprecated, err := jwt.NewKeysFromFilenames("public.pem", "private.pem", nil)
//	j := jwt.New(key, jwt.WithDeprecatedKeys(deprecated...))
//	token, err := j.GetSignedToken(jwt.NewRegisteredClaimsWithLifetime(time.Hour))
package jwt
