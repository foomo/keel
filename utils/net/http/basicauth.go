package httputils

import (
	"golang.org/x/crypto/bcrypt"
)

func HashBasicAuthPassword(v []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(v, bcrypt.DefaultCost)
}
