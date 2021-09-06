package jetstream

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

type (
	Stream struct {
		l           *zap.Logger
		js          nats.JetStreamContext
		conn        *nats.Conn
		name        string
		info        *nats.StreamInfo
		config      *nats.StreamConfig
		namespace   string
		natsOptions []nats.Option
	}
	Option           func(*Stream)
	Initializer      func(*zap.Logger, nats.JetStreamContext) error
	PublisherOption  func(*Publisher)
	SubscriberOption func(*Subscriber)
)

// WithNamespace option
func WithNamespace(v string) Option {
	return func(o *Stream) {
		o.namespace = v
	}
}

func WithConfig(v *nats.StreamConfig) Option {
	return func(o *Stream) {
		o.config = v
	}
}

// WithNatsOptions option
func WithNatsOptions(v ...nats.Option) Option {
	return func(o *Stream) {
		o.natsOptions = v
	}
}

func PublisherWithPubOpts(v ...nats.PubOpt) PublisherOption {
	return func(o *Publisher) {
		o.pubOpts = v
	}
}

func PublisherWithHeader(v nats.Header) PublisherOption {
	return func(o *Publisher) {
		o.header = v
	}
}

func SubscriberWithNamespace(v string) SubscriberOption {
	return func(o *Subscriber) {
		o.namespace = v
	}
}

func SubscriberWithSubOpts(v ...nats.SubOpt) SubscriberOption {
	return func(o *Subscriber) {
		o.opts = v
	}
}

func New(l *zap.Logger, name, addr string, opts ...Option) (*Stream, error) {
	stream := &Stream{
		l: l.With(
			log.FMessagingSystem("jetstream"),
		),
		name: name,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(stream)
		}
	}
	// default options
	natsOpts := append([]nats.Option{
		nats.ErrorHandler(
			func(conn *nats.Conn, subscription *nats.Subscription, err error) {
				l.Error("nats error", log.FError(err))
			}),
		nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
			l.Error("nats disconnected error", log.FError(err))
		}),
		nats.Timeout(time.Millisecond * 500),
	}, stream.natsOptions...)

	// connect nats
	conn, err := nats.Connect(addr, natsOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to nats")
	}
	stream.conn = conn

	// create jet stream
	js, err := conn.JetStream(
		nats.PublishAsyncErrHandler(func(js nats.JetStream, msg *nats.Msg, err error) {
			l.Error("nats disconnected error", log.FError(err))
		}),
	)
	if err != nil {
		return nil, err
	}
	stream.js = js

	// create / update stream if config exists
	if stream.config != nil {
		stream.config.Name = stream.Name()
		if _, err = js.StreamInfo(stream.Name()); err != nil {
			if info, err := js.AddStream(stream.config); err != nil {
				return nil, errors.Wrap(err, "failed to add stream")
			} else {
				stream.info = info
			}
		} else {
			if info, err := js.UpdateStream(stream.config); err != nil {
				return nil, errors.Wrap(err, "failed to update stream")
			} else {
				stream.info = info
			}
		}
	}

	return stream, nil
}

func (s *Stream) JS() nats.JetStreamContext {
	return s.js
}

func (s *Stream) Conn() *nats.Conn {
	return s.conn
}

func (s *Stream) Name() string {
	return s.name
}

func (s *Stream) Info() *nats.StreamInfo {
	return s.info
}

func (s *Stream) Publisher(subject string, opts ...PublisherOption) *Publisher {
	pub := &Publisher{
		stream:    s,
		subject:   subject,
		namespace: s.namespace,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(pub)
		}
	}
	return pub
}

func (s *Stream) Subscriber(subject string, opts ...SubscriberOption) *Subscriber {
	sub := &Subscriber{
		stream:    s,
		subject:   subject,
		namespace: s.namespace,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(sub)
		}
	}
	return sub
}
