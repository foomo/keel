package main

import (
	"net/http"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/foomo/keel"
	"github.com/foomo/keel/log"
	"github.com/foomo/keel/net/stream/jetstream"
	httputils "github.com/foomo/keel/utils/net/http"
)

func main() {
	// docker run -it -p 4222:4222 --rm nats:2.7-alpine --jetstream
	svr := keel.NewServer()

	l := svr.Logger()

	stream, err := jetstream.New(
		l,
		"demo",
		"http://localhost:4222",
		// set a custom namespace
		jetstream.WithNamespace("com.example"),
		// provide config which will create / update the stream
		jetstream.WithConfig(&nats.StreamConfig{
			Subjects:   []string{"*.*.test"},
			Replicas:   1,
			Retention:  nats.LimitsPolicy,
			Duplicates: 1 * time.Second,
			MaxMsgs:    1000000,
			MaxAge:     4 * time.Hour,
			Storage:    nats.MemoryStorage,
			Discard:    nats.DiscardOld,
		}),
	)
	log.Must(l, err, "failed to create stream")

	subject := "test"

	// create a sender for a subject
	pub := stream.Publisher(subject)
	// create a receiver for a subject
	sub := stream.Subscriber(
		subject,
		jetstream.SubscriberWithSubOpts(
			nats.DeliverNew(),
		),
	)

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// publish message
		if _, err := pub.JS().Publish(pub.Subject(), []byte("foo"), pub.PubOpts()...); err != nil {
			httputils.InternalServerError(l, w, r, err)
			return
		}
		l.Info("sent message")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// subscribe to the subject
	subscription, err := sub.JS().Subscribe(sub.Subject(), func(msg *nats.Msg) {
		l.Info("received message", log.FValue(string(msg.Data)), log.FMessagingDestination(msg.Subject))
	}, sub.SubOpts()...)
	log.Must(l, err, "failed to subscribe to subject")

	// add closes (NOTE: order matters)
	svr.AddClosers(subscription, stream.Conn())

	svr.AddService(
		keel.NewServiceHTTP(l, "demo", ":8080", svs),
	)

	svr.Run()
}
