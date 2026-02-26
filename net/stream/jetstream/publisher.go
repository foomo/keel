package jetstream

import (
	"github.com/nats-io/nats.go"
)

type (
	Publisher struct {
		stream    *Stream
		subject   string
		namespace string
		pubOpts   []nats.PubOpt
		marshal   MarshalFn
		header    nats.Header
	}
	MarshalFn func(v any) ([]byte, error)
)

func (s *Publisher) JS() nats.JetStreamContext {
	return s.stream.js
}

func (s *Publisher) Subject() string {
	if s.namespace != "" {
		return s.namespace + "." + s.subject
	}

	return s.subject
}

func (s *Publisher) NewMsg(v any) (*nats.Msg, error) {
	data, err := s.Marshal(v)
	if err != nil {
		return nil, err
	}

	msg := &nats.Msg{
		Subject: s.Subject(),
		Header:  s.header,
		Data:    data,
	}

	return msg, nil
}

func (s *Publisher) PubOpts(opts ...nats.PubOpt) []nats.PubOpt {
	return append(s.pubOpts, opts...)
}

func (s *Publisher) PublishMsg(data any, opts ...nats.PubOpt) (*nats.PubAck, error) {
	if msg, err := s.NewMsg(data); err != nil {
		return nil, err
	} else {
		return s.JS().PublishMsg(msg, s.PubOpts(opts...)...)
	}
}

func (s *Publisher) PublishMsgAsync(data any, opts ...nats.PubOpt) (nats.PubAckFuture, error) {
	if msg, err := s.NewMsg(data); err != nil {
		return nil, err
	} else {
		return s.JS().PublishMsgAsync(msg, s.PubOpts(opts...)...)
	}
}

func (s *Publisher) Marshal(v any) ([]byte, error) {
	return s.marshal(v)
}
