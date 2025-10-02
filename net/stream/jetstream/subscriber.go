package jetstream

import (
	"context"

	"github.com/nats-io/nats.go"

	"github.com/foomo/keel/log"
	"github.com/foomo/keel/net/stream"
)

type (
	Subscriber struct {
		stream    *Stream
		subject   string
		namespace string
		unmarshal UnmarshalFn
		opts      []nats.SubOpt
	}
	UnmarshalFn func(data []byte, v interface{}) error
)

func (s *Subscriber) JS() nats.JetStreamContext {
	return s.stream.js
}

func (s *Subscriber) Subject() string {
	if s.namespace != "" {
		return s.namespace + "." + s.subject
	}

	return s.subject
}

func (s *Subscriber) SubOpts(opts ...nats.SubOpt) []nats.SubOpt {
	return append(s.opts, opts...)
}

func (s *Subscriber) Subscribe(handler stream.MsgHandler, opts ...nats.SubOpt) (*nats.Subscription, error) {
	return s.JS().Subscribe(s.Subject(), func(msg *nats.Msg) {
		ctx := context.Background()
		if err := handler(ctx, s.stream.l, msg); err != nil {
			s.errorHandler(err)
		}
	}, s.SubOpts(opts...)...)
}

func (s *Subscriber) ChanSubscribe(ch chan *nats.Msg, opts ...nats.SubOpt) (*nats.Subscription, error) {
	return s.JS().ChanSubscribe(s.Subject(), ch, opts...)
}

func (s *Subscriber) QueueSubscribe(queue string, handler stream.MsgHandler, opts ...nats.SubOpt) (*nats.Subscription, error) {
	return s.JS().QueueSubscribe(s.Subject(), queue, func(msg *nats.Msg) {
		ctx := context.Background()
		if err := handler(ctx, s.stream.l, msg); err != nil {
			s.errorHandler(err)
		}
	}, s.SubOpts(opts...)...)
}

func (s *Subscriber) Unmarshal(msg *nats.Msg, v interface{}) error {
	return s.unmarshal(msg.Data, v)
}

func (s *Subscriber) errorHandler(err error) {
	s.stream.l.Error("failed to handle message", log.FError(err))
}
