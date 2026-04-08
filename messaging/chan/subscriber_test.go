package _chan_test

import (
	"context"
	"fmt"

	"github.com/foomo/keel/messaging"
	_chan "github.com/foomo/keel/messaging/chan"
	"github.com/foomo/keel/messaging/testing"
)

func ExampleNewSubscriber() {
	bus := _chan.NewBus[Event]()
	sub, err := _chan.NewSubscriber(bus, 1)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	await := testing.GoAsync(ctx, func(ctx context.Context, done context.CancelFunc) {
		_ = sub.Subscribe(ctx, "events", func(_ context.Context, msg messaging.Message[Event]) error {
			fmt.Println(msg)
			done()
			return nil
		})
	})

	pub := _chan.NewPublisher(bus)
	if err := pub.Publish(ctx, "events", Event{ID: "2", Name: "bar"}); err != nil {
		panic(err)
	}

	<-await
	// Output: {events {2 bar}}
}
