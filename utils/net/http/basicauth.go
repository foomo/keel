package httputils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashBasicAuthPassword returns the bcrypt hash of the password v using
// [bcrypt.DefaultCost], suitable for storing basic auth credentials.
func HashBasicAuthPassword(v []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(v, bcrypt.DefaultCost)
}
