package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// MaxTimeDifferenceBetweenNodes represents an offset that should be taken
// into account when creating e.g. jwt tokens with the `notBefore` flag.
// It is the default offset used by [NewRegisteredClaims].
var MaxTimeDifferenceBetweenNodes = 30 * time.Second

// NewStandardClaims returns registered claims using
// [MaxTimeDifferenceBetweenNodes] as offset.
//
// Deprecated: Use [NewRegisteredClaims] instead.
func NewStandardClaims() jwt.RegisteredClaims {
	return NewRegisteredClaims(
		WithOffset(MaxTimeDifferenceBetweenNodes),
	)
}

// NewStandardClaimsWithLifetime returns registered claims expiring after
// lifetime, using [MaxTimeDifferenceBetweenNodes] as offset.
//
// Deprecated: Use [NewRegisteredClaimsWithLifetime] instead.
func NewStandardClaimsWithLifetime(lifetime time.Duration) jwt.RegisteredClaims {
	return NewRegisteredClaimsWithLifetime(lifetime, WithOffset(MaxTimeDifferenceBetweenNodes))
}

// RegisteredClaimsOption configures how [NewRegisteredClaims] creates claims.
type RegisteredClaimsOption func(*registeredClaimsOptions)

// registeredClaimsOptions holds the settings applied by [RegisteredClaimsOption].
type registeredClaimsOptions struct {
	offset time.Duration
}

// WithOffset sets the offset subtracted from the current time to account for
// clock differences between nodes. Defaults to [MaxTimeDifferenceBetweenNodes].
func WithOffset(offset time.Duration) RegisteredClaimsOption {
	return func(o *registeredClaimsOptions) {
		o.offset = offset
	}
}

// NewRegisteredClaims returns [jwt.RegisteredClaims] with IssuedAt and
// NotBefore set to the current time minus the offset. The offset accounts
// for clock differences between nodes in a distributed system; offsets under
// one millisecond are ignored. If no [WithOffset] option is provided,
// [MaxTimeDifferenceBetweenNodes] is used.
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

// NewRegisteredClaimsWithLifetime returns claims as created by
// [NewRegisteredClaims] with ExpiresAt set to IssuedAt plus lifetime. Since
// IssuedAt is shifted back by the offset, the effective remaining lifetime
// is lifetime minus the offset.
func NewRegisteredClaimsWithLifetime(lifetime time.Duration, opts ...RegisteredClaimsOption) jwt.RegisteredClaims {
	claims := NewRegisteredClaims(opts...)
	claims.ExpiresAt = jwt.NewNumericDate(claims.IssuedAt.Add(lifetime))

	return claims
}
