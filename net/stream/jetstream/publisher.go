package jetstream

import (
	"github.com/nats-io/nats.go"
)

type (
	// Publisher publishes marshaled payloads to a single subject of a
	// [Stream]. Create one with [Stream.Publisher].
	Publisher struct {
		stream    *Stream
		subject   string
		namespace string
		pubOpts   []nats.PubOpt
		marshal   MarshalFn
		header    nats.Header
	}
	// MarshalFn encodes a payload into message data.
	MarshalFn func(v any) ([]byte, error)
)

// JS returns the JetStream context of the underlying [Stream].
func (s *Publisher) JS() nats.JetStreamContext {
	return s.stream.js
}

// Subject returns the subject messages are published to, prefixed with
// the namespace and a dot if a namespace is set.
func (s *Publisher) Subject() string {
	if s.namespace != "" {
		return s.namespace + "." + s.subject
	}

	return s.subject
}

// NewMsg marshals v and returns a message addressed to [Publisher.Subject]
// carrying the publisher's header. It returns the marshal error, if any.
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

// PubOpts returns the publisher's default publish options followed by opts.
func (s *Publisher) PubOpts(opts ...nats.PubOpt) []nats.PubOpt {
	return append(s.pubOpts, opts...)
}

// PublishMsg marshals data and publishes it synchronously, waiting for the
// server acknowledgement. opts are appended to the publisher's defaults.
func (s *Publisher) PublishMsg(data any, opts ...nats.PubOpt) (*nats.PubAck, error) {
	if msg, err := s.NewMsg(data); err != nil {
		return nil, err
	} else {
		return s.JS().PublishMsg(msg, s.PubOpts(opts...)...)
	}
}

// PublishMsgAsync marshals data and publishes it without waiting for the
// acknowledgement, which is delivered through the returned future. opts are
// appended to the publisher's defaults.
func (s *Publisher) PublishMsgAsync(data any, opts ...nats.PubOpt) (nats.PubAckFuture, error) {
	if msg, err := s.NewMsg(data); err != nil {
		return nil, err
	} else {
		return s.JS().PublishMsgAsync(msg, s.PubOpts(opts...)...)
	}
}

// Marshal encodes v with the publisher's [MarshalFn].
func (s *Publisher) Marshal(v any) ([]byte, error) {
	return s.marshal(v)
}
