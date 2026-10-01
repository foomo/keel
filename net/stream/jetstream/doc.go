// Package jetstream provides a thin wrapper around NATS JetStream with
// structured logging, reconnect handling and typed publishers and subscribers.
//
// A [Stream] holds the NATS connection and JetStream context. Use
// [Stream.Publisher] and [Stream.Subscriber] to create subject bound
// publishers and subscribers that marshal and unmarshal message payloads
// (JSON by default):
//
//	s, err := jetstream.New(l, "orders", nats.DefaultURL,
//		jetstream.WithNamespace("shop"),
//		jetstream.WithConfig(&nats.StreamConfig{Subjects: []string{"shop.>"}}),
//	)
//	if err != nil {
//		return err
//	}
//	defer s.Close()
//
//	pub := s.Publisher("created")
//	_, err = pub.PublishMsg(order)
//
// All publishers and subscribers created through a [Stream] are recorded
// and listed by [Readme].
package jetstream
