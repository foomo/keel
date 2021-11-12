package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/dustin/go-humanize"
	"github.com/foomo/keel/example/jetbench/msg"
	"github.com/foomo/keel/net/stream/jetstream"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

func main() {
	
	l, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}

	id := uuid.New().String()
	js, err := jetstream.New(
		l, 
		"jetbench-"+id, 
		"nats-main",
		jetstream.WithConfig(&nats.StreamConfig{
			Subjects:   []string{"jetbench."+id},
			Replicas:   1,
			Duplicates: 1 * time.Second,
			MaxAge:     24 * time.Hour,
			Storage:    nats.FileStorage,
			Retention:  nats.LimitsPolicy,
			Discard:    nats.DiscardOld,
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	spew.Dump("jetstream info", js.Info())

	p := js.Publisher(
		"jetbench."+ id,
		jetstream.PublisherWithMarshal(func(msg interface{}) ([]byte, error) {
			if m, ok := msg.(proto.Message); ok {
				return proto.Marshal(m)
			}
			return nil, errors.New("unexpected message type")
		}),
	)

	var (
		start = time.Now()
		m = &msg.Message{
			Sku: "012345678902323",
			Source: "01234567893256",
			InStock: true,
			Negotiable: true,
		}
		numMessages = 100000
	)

	for i:=0; i<numMessages; i++ {
		_, err := p.PublishMsg(m)
		if err != nil {
			log.Fatal(err)
		}
	}
	
	fmt.Println("published", humanize.Comma(int64(numMessages)), "messages in", time.Since(start))

	start = time.Now()
	s := js.Subscriber(
		"jetbench."+id,
		jetstream.SubscriberWithSubOpts(nats.AckAll()),
	)
	
	count := 0
	//done := make(chan bool)
	batchSize := 1000

	fmt.Println("using PullSubscribe and batchSize", batchSize)
	sub, err := s.JS().PullSubscribe("jetbench."+id, id)
	if err != nil {
		log.Fatal(err)
	}

	for {
		messages, err := sub.Fetch(batchSize)
		if err != nil {
			log.Println(err)
			break
		}
		fmt.Println("got batch", len(messages))

		for _, newMessage := range messages {
			count++
			ms := new(msg.Message)
			errUnmarshal := proto.Unmarshal(newMessage.Data, ms)
			if errUnmarshal != nil {
				log.Fatal(errUnmarshal)
			}
		}
		
		// ack last
		errAck := messages[len(messages)-1].Ack()
		if errAck != nil {
			log.Fatal(errAck)
		}
		
		if (count == numMessages) {
			//done <- true
			break
		}
	}

	// Async
	/*
	sub, err := s.Subscribe(func(ctx context.Context, logger *zap.Logger, message *nats.Msg) error {
		count++
		ms := new(msg.Message)
		errUnmarshal := proto.Unmarshal(message.Data, ms)
		if errUnmarshal != nil {
			return errUnmarshal
		}
		errAck := message.Ack()
		if errAck != nil {
			return errAck
		}
		if (count == numMessages) {
			done <- true
		}
		return nil
	}, nats.AckWait(10 * time.Second))
	if err != nil {
		log.Fatal(err)
	}
	*/
	defer sub.Unsubscribe()

	//<-done

	fmt.Println(sub.Subject, "received", humanize.Comma(int64(numMessages)), "messages in", time.Since(start))
}