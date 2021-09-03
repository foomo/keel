package middleware

import "github.com/google/uuid"

type SessionIDGenerator func() string

func DefaultSessionIDGenerator() string {
	return uuid.New().String()
}
