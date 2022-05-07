package jetstream

import (
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

type (
	Stream struct {
		l                      *zap.Logger
		addr                   string
		js                     nats.JetStreamContext
		conn                   *nats.Conn
		name                   string
		info                   *nats.StreamInfo
		config                 *nats.StreamConfig
		namespace              string
		natsOptions            []nats.Option
		reconnectMaxRetries    int
		reconnectTimeout       time.Duration
		reconnectFailedHandler func(error)
	}
	Option           func(*Stream)
	PublisherOption  func(*Publisher)
	SubscriberOption func(*Subscriber)
)

// WithNamespace option
func WithNamespace(v string) Option {
	return func(o *Stream) {
		o.namespace = v
	}
}

// WithReconnectFailedHandler option
func WithReconnectFailedHandler(v func(error)) Option {
	return func(o *Stream) {
		o.reconnectFailedHandler = v
	}
}

// WithReconnectTimeout option
func WithReconnectTimeout(v time.Duration) Option {
	return func(o *Stream) {
		o.reconnectTimeout = v
	}
}

// WithReconnectMaxRetries option
func WithReconnectMaxRetries(v int) Option {
	return func(o *Stream) {
		o.reconnectMaxRetries = v
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

func PublisherWithMarshal(marshal MarshalFn) PublisherOption {
	return func(o *Publisher) {
		o.marshal = marshal
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

func SubscriberWithUnmarshal(unmarshal UnmarshalFn) SubscriberOption {
	return func(o *Subscriber) {
		o.unmarshal = unmarshal
	}
}

func (s *Stream) connect() error {
	// connect nats
	conn, err := nats.Connect(s.addr, s.natsOptions...)
	if err != nil {
		return errors.Wrap(err, "failed to connect to nats addr "+s.addr)
	}
	s.l.Info("nats connected", zap.String("addr", s.addr))

	// create jet stream
	js, err := conn.JetStream(
		nats.PublishAsyncErrHandler(func(js nats.JetStream, msg *nats.Msg, err error) {
			s.l.Error("nats async publish error", log.FError(err))
		}),
	)
	if err != nil {
		return err
	}
	s.l.Info("jetstream created", zap.String("namespace", s.namespace))

	// create / update stream if config exists
	if s.config != nil {
		s.config.Name = s.Name()
		if _, err = js.StreamInfo(s.Name()); errors.Is(err, nats.ErrStreamNotFound) {
			if info, err := js.AddStream(s.config); err != nil {
				return errors.Wrap(err, "failed to add stream")
			} else if err != nil {
				return errors.Wrap(err, "failed to retrieve stream info")
			} else {
				s.info = info
			}
		} else if err != nil {
			return errors.Wrap(err, "failed get stream info")
		} else if info, err := js.UpdateStream(s.config); err != nil {
			return errors.Wrap(err, "failed to update stream")
		} else {
			s.info = info
		}
	}
	s.l.Info("jetstream configured")

	s.js = js
	s.conn = conn

	return nil
}

func (s *Stream) initNatsOptions() {
	natsOpts := append([]nats.Option{
		nats.ErrorHandler(func(conn *nats.Conn, subscription *nats.Subscription, err error) {
			s.l.Error("nats error", log.FError(err), log.FStreamQueue(subscription.Queue), log.FStreamSubject(subscription.Subject))
		}),
		nats.ClosedHandler(func(conn *nats.Conn) {
			if err := conn.LastError(); err != nil {
				s.l.Error("nats closed", log.FError(err))
			} else {
				s.l.Info("nats closed")
			}
		}),
		nats.ReconnectHandler(func(conn *nats.Conn) {
			s.l.Info("nats reconnected")
		}),
		nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
			if err != nil {
				s.l.Error("nats disconnected error", log.FError(err))

				var errRetry error
				for i := 0; i < s.reconnectMaxRetries; i++ {
					errRetry = s.connect()
					if errRetry != nil {
						s.l.Error("nats reconnect failed", log.FError(errRetry))
						time.Sleep(s.reconnectTimeout)
					} else {
						break
					}
				}

				// all retries failed
				if errRetry != nil {
					s.reconnectFailedHandler(errRetry)
				} else {
					s.l.Info("reconnected to nats after error")
				}
			} else {
				s.l.Info("nats disconnected")
			}
		}),
		nats.Timeout(time.Millisecond * 500),
	}, s.natsOptions...)

	s.natsOptions = natsOpts
}

func New(l *zap.Logger, name, addr string, opts ...Option) (*Stream, error) {
	stream := &Stream{
		l: l.With(
			log.FMessagingSystem("jetstream"),
			log.FName(name),
		),
		name: name,
		addr: addr,

		// default reconnect settings
		reconnectMaxRetries: 10,
		reconnectTimeout:    15 * time.Second,
		reconnectFailedHandler: func(e error) {
			panic(e)
		},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(stream)
		}
	}

	// default nats options
	stream.initNatsOptions()

	// initial connect
	if err := stream.connect(); err != nil {
		return nil, err
	}

	return stream, nil
}

func (s *Stream) Addr() string {
	return s.addr
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
		marshal:   json.Marshal,
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
		unmarshal: json.Unmarshal,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(sub)
		}
	}
	return sub
}

func (s *Stream) Close() {
	s.conn.Close()
}
