package jwt

import (
	"time"

	"github.com/golang-jwt/jwt"
)

// MaxTimeDifferenceBetweenNodes represents an offset that should be taken
// into account when creating e.g. jwt tokens with the `notBefore` flag.
var MaxTimeDifferenceBetweenNodes = time.Second * 30

func NewStandardClaims() jwt.StandardClaims {
	now := time.Now().Add(-MaxTimeDifferenceBetweenNodes)

	return jwt.StandardClaims{
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
	}
}

func NewStandardClaimsWithLifetime(lifetime time.Duration) jwt.StandardClaims {
	claims := NewStandardClaims()
	claims.ExpiresAt = claims.IssuedAt + int64(lifetime.Seconds())

	return claims
}
