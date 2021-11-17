package jwt

import (
	"github.com/golang-jwt/jwt"
)

type (
	JWT struct {
		// key for signing
		Key Key
		// KeyFunc provider
		KeyFunc jwt.Keyfunc
		// DeprecatedKeys  e.g. due to rotation
		DeprecatedKeys map[string]Key
	}
	Option func(*JWT)
)

// WithKeyFun middleware option
func WithKeyFun(v jwt.Keyfunc) Option {
	return func(o *JWT) {
		o.KeyFunc = v
	}
}

// WithDeprecatedKeys middleware option
func WithDeprecatedKeys(v ...Key) Option {
	return func(o *JWT) {
		deprecatedKeys := make(map[string]Key, len(v))
		for _, key := range deprecatedKeys {
			deprecatedKeys[key.ID] = key
		}
		o.DeprecatedKeys = deprecatedKeys
	}
}

// New returns a new JWT for the given key and optional old keys e.g. due to rotation
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

func (j *JWT) GetSignedToken(claims jwt.Claims) (string, error) {
	// create token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = j.Key.ID
	return token.SignedString(j.Key.Private)
}

func (j *JWT) ParseWithClaims(token string, claims jwt.Claims) (*jwt.Token, error) {
	return jwt.ParseWithClaims(token, claims, j.KeyFunc)
}
