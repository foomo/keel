package jetstream

import (
	"context"

	"github.com/nats-io/nats.go"

	"github.com/foomo/keel/log"
	"github.com/foomo/keel/net/stream"
)

type (
	// Subscriber subscribes to a single subject of a [Stream]. Create one with
	// [Stream.Subscriber].
	Subscriber struct {
		stream    *Stream
		subject   string
		namespace string
		unmarshal UnmarshalFn
		opts      []nats.SubOpt
	}
	// UnmarshalFn decodes message data into v.
	UnmarshalFn func(data []byte, v any) error
)

// JS returns the JetStream context of the underlying [Stream].
func (s *Subscriber) JS() nats.JetStreamContext {
	return s.stream.js
}

// Subject returns the subscribed subject, prefixed with the namespace and a
// dot if a namespace is set.
func (s *Subscriber) Subject() string {
	if s.namespace != "" {
		return s.namespace + "." + s.subject
	}

	return s.subject
}

// SubOpts returns the subscriber's default subscribe options followed by
// opts.
func (s *Subscriber) SubOpts(opts ...nats.SubOpt) []nats.SubOpt {
	return append(s.opts, opts...)
}

// Subscribe creates an asynchronous subscription that calls handler for each
// message with a background context and the stream's logger. Handler errors
// are logged. opts are appended to the subscriber's defaults.
func (s *Subscriber) Subscribe(handler stream.MsgHandler, opts ...nats.SubOpt) (*nats.Subscription, error) {
	return s.JS().Subscribe(s.Subject(), func(msg *nats.Msg) {
		ctx := context.Background()
		if err := handler(ctx, s.stream.l, msg); err != nil {
			s.errorHandler(err)
		}
	}, s.SubOpts(opts...)...)
}

// ChanSubscribe creates a subscription that delivers messages to ch. Only
// opts are used; the subscriber's default options are not applied.
func (s *Subscriber) ChanSubscribe(ch chan *nats.Msg, opts ...nats.SubOpt) (*nats.Subscription, error) {
	return s.JS().ChanSubscribe(s.Subject(), ch, opts...)
}

// QueueSubscribe is like [Subscriber.Subscribe] but joins the queue group
// queue, so each message is delivered to only one member of the group.
func (s *Subscriber) QueueSubscribe(queue string, handler stream.MsgHandler, opts ...nats.SubOpt) (*nats.Subscription, error) {
	return s.JS().QueueSubscribe(s.Subject(), queue, func(msg *nats.Msg) {
		ctx := context.Background()
		if err := handler(ctx, s.stream.l, msg); err != nil {
			s.errorHandler(err)
		}
	}, s.SubOpts(opts...)...)
}

// Unmarshal decodes the data of msg into v with the subscriber's
// [UnmarshalFn].
func (s *Subscriber) Unmarshal(msg *nats.Msg, v any) error {
	return s.unmarshal(msg.Data, v)
}

// errorHandler logs a handler error.
func (s *Subscriber) errorHandler(err error) {
	s.stream.l.Error("failed to handle message", log.FError(err))
}
