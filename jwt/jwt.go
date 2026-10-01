package jwt

import (
	"github.com/golang-jwt/jwt/v5"
)

type (
	// JWT signs and parses RS256 tokens. Create it with [New].
	JWT struct {
		// Key is the current key used for signing and verification.
		Key Key
		// KeyFunc resolves the verification key for a parsed token.
		KeyFunc jwt.Keyfunc
		// DeprecatedKeys holds verification-only keys indexed by key ID,
		// e.g. kept after a key rotation.
		DeprecatedKeys map[string]Key
	}
	// Option configures a [JWT].
	Option func(*JWT)
)

// WithKeyFun sets the key function used to verify tokens. Defaults to
// [DefaultKeyFunc] over the current and deprecated keys.
func WithKeyFun(v jwt.Keyfunc) Option {
	return func(o *JWT) {
		o.KeyFunc = v
	}
}

// WithDeprecatedKeys adds keys that are still accepted for verification,
// indexed by their [Key.ID].
func WithDeprecatedKeys(v ...Key) Option {
	return func(o *JWT) {
		if len(v) > 0 {
			if o.DeprecatedKeys == nil {
				o.DeprecatedKeys = map[string]Key{}
			}

			for _, key := range v {
				o.DeprecatedKeys[key.ID] = key
			}
		}
	}
}

// New returns a [JWT] that signs with key. If no key function is set via
// [WithKeyFun], [DefaultKeyFunc] is used with key and any keys added via
// [WithDeprecatedKeys]. Nil options are ignored.
func New(key Key, opts ...Option) *JWT {
	inst := &JWT{
		Key: key,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	if inst.KeyFunc == nil {
		inst.KeyFunc = DefaultKeyFunc(key, inst.DeprecatedKeys)
	}

	return inst
}

// GetSignedToken returns claims as a token signed with RS256 using the
// private part of [JWT.Key], with the "kid" header set to its ID. It
// returns an error if signing fails, e.g. when no private key is set.
func (j *JWT) GetSignedToken(claims jwt.Claims) (string, error) {
	// create token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = j.Key.ID

	return token.SignedString(j.Key.Private)
}

// ParseWithClaims parses and validates token into claims, resolving the
// verification key with [JWT.KeyFunc].
func (j *JWT) ParseWithClaims(token string, claims jwt.Claims) (*jwt.Token, error) {
	return jwt.ParseWithClaims(token, claims, j.KeyFunc)
}
