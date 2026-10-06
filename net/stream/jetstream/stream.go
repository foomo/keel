package jetstream

import (
	"encoding/json"
	"net/url"
	"slices"
	"strconv"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
	foomosemconv "github.com/foomo/opentelemetry-go/semconv"
)

type (
	// Stream wraps a NATS connection and its JetStream context for a single
	// named stream. Create one with [New].
	Stream struct {
		l                      *zap.Logger
		addr                   string
		js                     nats.JetStreamContext
		conn                   *nats.Conn
		name                   string
		info                   *nats.StreamInfo
		config                 *nats.StreamConfig
		configJSOptions        []nats.JSOpt
		namespace              string
		jsOptions              []nats.JSOpt
		natsOptions            []nats.Option
		reconnectMaxRetries    int
		reconnectTimeout       time.Duration
		reconnectFailedHandler func(error)
	}
	// Option configures a [Stream].
	Option func(*Stream)
	// PublisherOption configures a [Publisher].
	PublisherOption func(*Publisher)
	// SubscriberOption configures a [Subscriber].
	SubscriberOption func(*Subscriber)
)

// WithNamespace sets the namespace prefixed to the subjects of all
// publishers and subscribers created by the stream.
func WithNamespace(v string) Option {
	return func(o *Stream) {
		o.namespace = v
	}
}

// WithReconnectFailedHandler sets the function called with the last error
// when all reconnect attempts failed. Defaults to a function that panics.
func WithReconnectFailedHandler(v func(error)) Option {
	return func(o *Stream) {
		o.reconnectFailedHandler = v
	}
}

// WithReconnectTimeout sets the delay between reconnect attempts after a
// disconnect error. Defaults to 15 seconds.
func WithReconnectTimeout(v time.Duration) Option {
	return func(o *Stream) {
		o.reconnectTimeout = v
	}
}

// WithReconnectMaxRetries sets the number of reconnect attempts after a
// disconnect error. Defaults to 10.
func WithReconnectMaxRetries(v int) Option {
	return func(o *Stream) {
		o.reconnectMaxRetries = v
	}
}

// WithConfig sets the stream configuration used to create or update the
// stream on connect. The config name is overwritten with the stream name.
// opts are passed to the stream management calls.
func WithConfig(v *nats.StreamConfig, opts ...nats.JSOpt) Option {
	return func(o *Stream) {
		o.config = v
		o.configJSOptions = append(o.configJSOptions, opts...)
	}
}

// WithJSOptions appends options used when creating the JetStream context.
func WithJSOptions(v ...nats.JSOpt) Option {
	return func(o *Stream) {
		o.jsOptions = append(o.jsOptions, v...)
	}
}

// WithNatsOptions appends options used when connecting to NATS. They are
// applied after, and may override, the stream's default handlers.
func WithNatsOptions(v ...nats.Option) Option {
	return func(o *Stream) {
		o.natsOptions = append(o.natsOptions, v...)
	}
}

// PublisherWithPubOpts appends default publish options of the publisher.
func PublisherWithPubOpts(v ...nats.PubOpt) PublisherOption {
	return func(o *Publisher) {
		o.pubOpts = append(o.pubOpts, v...)
	}
}

// PublisherWithMarshal sets the payload encoder. Defaults to [json.Marshal].
func PublisherWithMarshal(marshal MarshalFn) PublisherOption {
	return func(o *Publisher) {
		o.marshal = marshal
	}
}

// PublisherWithHeader sets the header attached to every published message.
func PublisherWithHeader(v nats.Header) PublisherOption {
	return func(o *Publisher) {
		o.header = v
	}
}

// SubscriberWithNamespace overrides the stream's namespace for the
// subscriber's subject.
func SubscriberWithNamespace(v string) SubscriberOption {
	return func(o *Subscriber) {
		o.namespace = v
	}
}

// SubscriberWithSubOpts appends default subscribe options of the subscriber.
func SubscriberWithSubOpts(v ...nats.SubOpt) SubscriberOption {
	return func(o *Subscriber) {
		o.opts = append(o.opts, v...)
	}
}

// SubscriberWithUnmarshal sets the payload decoder. Defaults to
// [json.Unmarshal].
func SubscriberWithUnmarshal(unmarshal UnmarshalFn) SubscriberOption {
	return func(o *Subscriber) {
		o.unmarshal = unmarshal
	}
}

// New connects to the NATS server at addr and returns a [Stream] named
// name. If a config is set via [WithConfig], the stream is created or
// updated. On a disconnect error the stream reconnects according to the
// reconnect options. It returns an error if the initial connect fails.
func New(l *zap.Logger, name, addr string, opts ...Option) (*Stream, error) {
	stream := &Stream{
		l: log.WithAttributes(l,
			foomosemconv.MessagingSystemNats,
			foomosemconv.MessagingNATSStream(name),
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

// Addr returns the NATS server address.
func (s *Stream) Addr() string {
	return s.addr
}

// JS returns the JetStream context.
func (s *Stream) JS() nats.JetStreamContext {
	return s.js
}

// Conn returns the NATS connection.
func (s *Stream) Conn() *nats.Conn {
	return s.conn
}

// Name returns the stream name.
func (s *Stream) Name() string {
	return s.name
}

// Info returns the stream info obtained when creating or updating the
// stream, or nil if no config was set.
func (s *Stream) Info() *nats.StreamInfo {
	return s.info
}

// Publisher returns a [Publisher] for subject using the stream's namespace
// and JSON encoding unless overridden by opts. It is recorded for [Readme].
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

	{ // append to recoreded publishers
		value := publisher{
			Stream:    s.name,
			Namespace: s.namespace,
			Subject:   subject,
		}
		if !slices.ContainsFunc(publishers, func(p publisher) bool {
			return p.Stream == value.Stream && p.Namespace == value.Namespace && p.Subject == value.Subject
		}) {
			publishers = append(publishers, value)
		}
	}

	return pub
}

// Subscriber returns a [Subscriber] for subject using the stream's namespace
// and JSON decoding unless overridden by opts. It is recorded for [Readme].
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

	{ // append to recoreded publishers
		value := subscriber{
			Stream:    s.name,
			Namespace: s.namespace,
			Subject:   subject,
		}
		if !slices.ContainsFunc(subscribers, func(p subscriber) bool {
			return p.Stream == value.Stream && p.Namespace == value.Namespace && p.Subject == value.Subject
		}) {
			subscribers = append(subscribers, value)
		}
	}

	return sub
}

// Close closes the NATS connection.
func (s *Stream) Close() {
	s.conn.Close()
}

// serverAttrs returns OTel semconv attributes describing the server at
// rawURL. Returns nil if rawURL is empty or cannot be parsed.
func serverAttrs(rawURL string) []attribute.KeyValue {
	addr, port := serverAddrPort(rawURL)
	if addr == "" {
		return nil
	}

	return []attribute.KeyValue{
		semconv.ServerAddress(addr),
		semconv.ServerPort(port),
		semconv.NetworkProtocolName("jetstream"),
	}
}

// serverAddrPort extracts just the address and port from rawURL.
// Returns zero values if parsing fails.
func serverAddrPort(rawURL string) (string, int) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", 0
	}

	port, _ := strconv.Atoi(u.Port())

	return u.Hostname(), port
}

// connect establishes the NATS connection and JetStream context and creates
// or updates the stream if a config is set.
func (s *Stream) connect() error {
	// connect nats
	conn, err := nats.Connect(s.addr, s.natsOptions...)
	if err != nil {
		return errors.Wrap(err, "failed to connect to nats addr "+s.addr)
	}

	l := log.WithAttributes(s.l, serverAttrs(conn.ConnectedUrlRedacted())...)
	l.Debug("jetstream connected")

	// create jet stream
	js, err := conn.JetStream(
		append(
			[]nats.JSOpt{
				nats.PublishAsyncErrHandler(func(js nats.JetStream, msg *nats.Msg, err error) {
					log.WithError(l, err).Error("nats async publish error")
				}),
			},
			s.jsOptions...,
		)...,
	)
	if err != nil {
		return err
	}

	l.Debug("jetstream created")

	// create / update stream if config exists
	if s.config != nil {
		s.config.Name = s.Name()
		if _, err = js.StreamInfo(s.Name(), s.configJSOptions...); errors.Is(err, nats.ErrStreamNotFound) {
			if info, err := js.AddStream(s.config, s.configJSOptions...); err != nil {
				return errors.Wrap(err, "failed to add stream")
			} else {
				s.info = info
			}
		} else if err != nil {
			return errors.Wrap(err, "failed get stream info")
		} else if info, err := js.UpdateStream(s.config, s.configJSOptions...); err != nil {
			return errors.Wrap(err, "failed to update stream")
		} else {
			s.info = info
		}
	}

	l.Info("jetstream configured")

	s.js = js
	s.conn = conn

	return nil
}

// initNatsOptions prepends the default logging, reconnect and timeout
// options to the user supplied NATS options.
func (s *Stream) initNatsOptions() {
	natsOpts := append([]nats.Option{
		nats.ErrorHandler(func(conn *nats.Conn, subscription *nats.Subscription, err error) {
			log.WithError(s.l, err).Error("nats error",
				log.Attribute(semconv.MessagingDestinationName(subscription.Queue)),
				log.Attribute(semconv.MessagingDestinationSubscriptionName(subscription.Subject)),
			)
		}),
		nats.ClosedHandler(func(conn *nats.Conn) {
			if err := conn.LastError(); err != nil {
				log.WithError(s.l, err).Error("nats closed")
			} else {
				s.l.Info("nats closed")
			}
		}),
		nats.ReconnectHandler(func(conn *nats.Conn) {
			s.l.Info("nats reconnected")
		}),
		nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
			if err != nil {
				log.WithError(s.l, err).Error("nats disconnected error")

				var errRetry error
				for range s.reconnectMaxRetries {
					errRetry = s.connect()
					if errRetry != nil {
						log.WithError(s.l, errRetry).Error("nats reconnect failed")
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
