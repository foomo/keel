package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// MaxTimeDifferenceBetweenNodes represents an offset that should be taken
// into account when creating e.g. jwt tokens with the `notBefore` flag.
var MaxTimeDifferenceBetweenNodes = 30 * time.Second

// Deprecated: NewStandardClaims use NewRegisteredClaims instead.
func NewStandardClaims() jwt.RegisteredClaims {
	return NewRegisteredClaims(
		WithOffset(MaxTimeDifferenceBetweenNodes),
	)
}

// Deprecated: NewStandardClaimsWithLifetime use NewRegisteredClaimsWithLifetime instead.
func NewStandardClaimsWithLifetime(lifetime time.Duration) jwt.RegisteredClaims {
	return NewRegisteredClaimsWithLifetime(lifetime, WithOffset(MaxTimeDifferenceBetweenNodes))
}

// RegisteredClaimsOption configures how RegisteredClaims are created.
type RegisteredClaimsOption func(*registeredClaimsOptions)

type registeredClaimsOptions struct {
	offset time.Duration
}

// WithOffset sets the offset to account for time differences between nodes.
func WithOffset(offset time.Duration) RegisteredClaimsOption {
	return func(o *registeredClaimsOptions) {
		o.offset = offset
	}
}

// NewRegisteredClaims returns a new jwt.RegisteredClaims with the IssuedAt and NotBefore fields set to the current time plus the given offset.
// The offset can be used to account for time differences between nodes in a distributed system.
// If no offset option is provided, MaxTimeDifferenceBetweenNodes is used as the default.
func NewRegisteredClaims(opts ...RegisteredClaimsOption) jwt.RegisteredClaims {
	o := &registeredClaimsOptions{offset: MaxTimeDifferenceBetweenNodes}
	for _, opt := range opts {
		opt(o)
	}
	// set IssuedAt and NotBefore to the current time minus the offset to account for time differences between nodes
	now := time.Now()
	if o.offset.Milliseconds() > 0 {
		now = now.Add(o.offset * -1)
	}

	return jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
	}
}

// NewRegisteredClaimsWithLifetime returns a new jwt.RegisteredClaims with the IssuedAt and NotBefore fields set to the current time plus the given optional offset and the ExpiresAt field set to the current time plus the given lifetime.
func NewRegisteredClaimsWithLifetime(lifetime time.Duration, opts ...RegisteredClaimsOption) jwt.RegisteredClaims {
	claims := NewRegisteredClaims(opts...)
	claims.ExpiresAt = jwt.NewNumericDate(claims.IssuedAt.Add(lifetime))

	return claims
}
