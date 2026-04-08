package service

import (
	"github.com/foomo/keel/messaging"
)

type Subscription[T any] struct {
	topic   string
	handler messaging.Handler[T]
}

func (s *Subscription[T]) Name() string {
	return s.topic
}
