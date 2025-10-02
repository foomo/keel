package jwt

import (
	"github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
)

func DefaultKeyFunc(key Key, deprecatedKeys map[string]Key) jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodRS256.Name {
			return nil, errors.New("unexpected jwt signing method: " + token.Method.Alg())
		}

		if kid, ok := token.Header["kid"]; !ok {
			return nil, errors.New("missing key identifier")
		} else if kidString, ok := kid.(string); !ok {
			return nil, errors.New("invalid key identifier type")
		} else if oldKey, ok := deprecatedKeys[kidString]; ok {
			return oldKey.Public, nil
		} else if kidString == key.ID {
			return key.Public, nil
		} else {
			return nil, errors.New("unknown key identifier: " + kidString + " (" + key.ID + ")")
		}
	}
}
