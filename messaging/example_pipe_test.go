package messaging_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/foomo/keel/messaging"
	_chan "github.com/foomo/keel/messaging/chan"
	"github.com/foomo/keel/messaging/testing"
)

// failPublisher is a Publisher that always returns the given error.
type failPublisher[T any] struct{ err error }

func (p *failPublisher[T]) Publish(context.Context, string, T) error { return p.err }
func (p *failPublisher[T]) Close() error                             { return nil }

type Event struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Summary struct {
	Label string `json:"label"`
}

// ExamplePipe demonstrates wiring a source subscriber to a destination
// publisher using messaging.Pipe. Every message received on "events" is forwarded
// to the destination bus on the same subject.
func ExamplePipe() {
	ctx := context.Background()

	// Source transport.
	srcBus := _chan.NewBus[Event]()
	srcPub := _chan.NewPublisher(srcBus)
	srcSub, err := _chan.NewSubscriber(srcBus, 1)
	if err != nil {
		panic(err)
	}

	// Destination transport.
	dstBus := _chan.NewBus[Event]()
	dstPub := _chan.NewPublisher(dstBus)
	dstSub, err := _chan.NewSubscriber(dstBus, 1)
	if err != nil {
		panic(err)
	}

	// Listen on destination — print when a message arrives.
	// Pipe preserves the original subject, so we subscribe to "events".
	await := testing.GoAsync(ctx, func(ctx context.Context, done context.CancelFunc) {
		_ = dstSub.Subscribe(ctx, "events", func(_ context.Context, msg messaging.Message[Event]) error {
			fmt.Println(msg.Subject, msg.Payload)
			done()
			return nil
		})
	})

	// Pipe: source subscriber forwards to destination publisher.
	_ = testing.GoAsync(ctx, func(ctx context.Context, done context.CancelFunc) {
		if err := srcSub.Subscribe(ctx, "events", messaging.Pipe[Event](dstPub)); err != nil {
			panic(err)
		}
	})

	if err := srcPub.Publish(ctx, "events", Event{ID: "1", Name: "hello"}); err != nil {
		panic(err)
	}

	<-await
	// Output: events {1 hello}
}

// ExamplePipeMap demonstrates transforming messages from one type to another.
// Events are mapped to Summary values and published to the destination bus.
func ExamplePipeMap() {
	ctx := context.Background()

	// Source transport (Event).
	srcBus := _chan.NewBus[Event]()
	srcPub := _chan.NewPublisher(srcBus)
	srcSub, err := _chan.NewSubscriber(srcBus, 1)
	if err != nil {
		panic(err)
	}

	// Destination transport (Summary).
	dstBus := _chan.NewBus[Summary]()
	dstPub := _chan.NewPublisher(dstBus)
	dstSub, err := _chan.NewSubscriber(dstBus, 1)
	if err != nil {
		panic(err)
	}

	// Map function: Event → Summary, keeping the same subject.
	mapFn := func(_ context.Context, msg messaging.Message[Event]) (messaging.Message[Summary], error) {
		return messaging.NewMessage(msg.Subject, Summary{Label: msg.Payload.Name}), nil
	}

	// Listen on destination — PipeMap publishes with the subject from the MapFunc result.
	await := testing.GoAsync(ctx, func(ctx context.Context, done context.CancelFunc) {
		_ = dstSub.Subscribe(ctx, "events", func(_ context.Context, msg messaging.Message[Summary]) error {
			fmt.Println(msg.Subject, msg.Payload)
			done()
			return nil
		})
	})

	// PipeMap: source subscriber maps Event→Summary and forwards.
	_ = testing.GoAsync(ctx, func(ctx context.Context, done context.CancelFunc) {
		_ = srcSub.Subscribe(ctx, "events", messaging.PipeMap[Event, Summary](dstPub, mapFn))
	})

	if err := srcPub.Publish(ctx, "events", Event{ID: "1", Name: "hello"}); err != nil {
		panic(err)
	}

	<-await
	// Output: events {hello}
}

// ExamplePipe_withFilter demonstrates filtering messages before they are
// forwarded. Only events whose ID is not "skip" reach the destination.
func ExamplePipe_withFilter() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Source transport.
	srcBus := _chan.NewBus[Event]()
	srcPub := _chan.NewPublisher(srcBus)
	srcSub, err := _chan.NewSubscriber(srcBus, 1)
	if err != nil {
		panic(err)
	}

	// Destination transport.
	dstBus := _chan.NewBus[Event]()
	dstPub := _chan.NewPublisher(dstBus)
	dstSub, err := _chan.NewSubscriber(dstBus, 1)
	if err != nil {
		panic(err)
	}

	// Filter: drop events with ID "skip".
	filter := func(_ context.Context, msg messaging.Message[Event]) (bool, error) {
		return msg.Payload.ID != "skip", nil
	}

	// Listen on destination — expect exactly one message.
	await := testing.GoAsync(ctx, func(ctx context.Context, done context.CancelFunc) {
		_ = dstSub.Subscribe(ctx, "events", func(_ context.Context, msg messaging.Message[Event]) error {
			fmt.Println(msg.Subject, msg.Payload)
			done()
			return nil
		})
	})

	// Pipe with filter.
	_ = testing.GoAsync(ctx, func(ctx context.Context, done context.CancelFunc) {
		_ = srcSub.Subscribe(ctx, "events", messaging.Pipe[Event](dstPub, messaging.WithFilter[Event](filter)))
	})

	// Allow goroutines to register on their buses before publishing.
	time.Sleep(10 * time.Millisecond)

	// First message is filtered out.
	if err := srcPub.Publish(ctx, "events", Event{ID: "skip", Name: "ignored"}); err != nil {
		panic(err)
	}
	// Second message passes the filter.
	if err := srcPub.Publish(ctx, "events", Event{ID: "keep", Name: "accepted"}); err != nil {
		panic(err)
	}

	<-await
	// Output: events {keep accepted}
}

// ExamplePipeMap_withFilter demonstrates filtering messages before they are
// mapped and forwarded. The filter runs on the source type (Event); only
// passing events are transformed into Summary values.
func ExamplePipeMap_withFilter() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Source transport (Event).
	srcBus := _chan.NewBus[Event]()
	srcPub := _chan.NewPublisher(srcBus)
	srcSub, err := _chan.NewSubscriber(srcBus, 1)
	if err != nil {
		panic(err)
	}

	// Destination transport (Summary).
	dstBus := _chan.NewBus[Summary]()
	dstPub := _chan.NewPublisher(dstBus)
	dstSub, err := _chan.NewSubscriber(dstBus, 1)
	if err != nil {
		panic(err)
	}

	// Map function: Event → Summary.
	mapFn := func(_ context.Context, msg messaging.Message[Event]) (messaging.Message[Summary], error) {
		return messaging.NewMessage(msg.Subject, Summary{Label: msg.Payload.Name}), nil
	}

	// Filter: drop events with ID "skip".
	filter := func(_ context.Context, msg messaging.Message[Event]) (bool, error) {
		return msg.Payload.ID != "skip", nil
	}

	// Listen on destination — expect exactly one message.
	await := testing.GoAsync(ctx, func(ctx context.Context, done context.CancelFunc) {
		_ = dstSub.Subscribe(ctx, "events", func(_ context.Context, msg messaging.Message[Summary]) error {
			fmt.Println(msg.Subject, msg.Payload)
			done()
			return nil
		})
	})

	// PipeMap with filter.
	_ = testing.GoAsync(ctx, func(ctx context.Context, done context.CancelFunc) {
		_ = srcSub.Subscribe(ctx, "events", messaging.PipeMap[Event, Summary](dstPub, mapFn, messaging.WithFilter[Event](filter)))
	})

	// Allow goroutines to register on their buses before publishing.
	time.Sleep(10 * time.Millisecond)

	// First message is filtered out before the map runs.
	if err := srcPub.Publish(ctx, "events", Event{ID: "skip", Name: "ignored"}); err != nil {
		panic(err)
	}
	// Second message passes the filter, gets mapped, and arrives.
	if err := srcPub.Publish(ctx, "events", Event{ID: "keep", Name: "accepted"}); err != nil {
		panic(err)
	}

	<-await
	// Output: events {accepted}
}

// ExamplePipe_withDeadLetter demonstrates the dead-letter handler on a Pipe.
// When the downstream publisher returns an error, the dead-letter function
// receives the original message and the error.
func ExamplePipe_withDeadLetter() {
	ctx := context.Background()

	// Source transport.
	srcBus := _chan.NewBus[Event]()
	srcPub := _chan.NewPublisher(srcBus)
	srcSub, err := _chan.NewSubscriber(srcBus, 1)
	if err != nil {
		panic(err)
	}

	// A publisher that always fails.
	pubErr := errors.New("publish failed")
	badPub := &failPublisher[Event]{err: pubErr}

	received := make(chan struct{})

	// Dead-letter handler prints the original message and error.
	deadLetter := func(_ context.Context, msg messaging.Message[Event], err error) {
		fmt.Println("dead-letter:", msg.Subject, msg.Payload, err)
		close(received)
	}

	// Pipe with dead-letter: publish errors are routed to the handler.
	_ = testing.GoAsync(ctx, func(ctx context.Context, done context.CancelFunc) {
		_ = srcSub.Subscribe(ctx, "events", messaging.Pipe[Event](badPub, messaging.WithDeadLetter[Event](deadLetter)))
	})

	if err := srcPub.Publish(ctx, "events", Event{ID: "1", Name: "hello"}); err != nil {
		panic(err)
	}

	<-received
	// Output: dead-letter: events {1 hello} publish failed
}

// ExamplePipeMap_withDeadLetter demonstrates the dead-letter handler on a
// PipeMap. When the MapFunc returns an error, the dead-letter function receives
// the original source message and the map error.
func ExamplePipeMap_withDeadLetter() {
	ctx := context.Background()

	// Source transport (Event).
	srcBus := _chan.NewBus[Event]()
	srcPub := _chan.NewPublisher(srcBus)
	srcSub, err := _chan.NewSubscriber(srcBus, 1)
	if err != nil {
		panic(err)
	}

	// Destination transport (Summary).
	dstBus := _chan.NewBus[Summary]()
	dstPub := _chan.NewPublisher(dstBus)
	dstSub, err := _chan.NewSubscriber(dstBus, 1)
	if err != nil {
		panic(err)
	}

	// Map function that fails for events with ID "bad".
	mapErr := errors.New("bad event")
	mapFn := func(_ context.Context, msg messaging.Message[Event]) (messaging.Message[Summary], error) {
		if msg.Payload.ID == "bad" {
			return messaging.Message[Summary]{}, mapErr
		}
		return messaging.NewMessage(msg.Subject, Summary{Label: msg.Payload.Name}), nil
	}

	// Dead-letter handler prints the original message and error.
	deadLetter := func(_ context.Context, msg messaging.Message[Event], err error) {
		fmt.Println("dead-letter:", msg.Subject, msg.Payload, err)
	}

	// Listen on destination for the valid message.
	await := testing.GoAsync(ctx, func(ctx context.Context, done context.CancelFunc) {
		_ = dstSub.Subscribe(ctx, "events", func(_ context.Context, msg messaging.Message[Summary]) error {
			fmt.Println(msg.Subject, msg.Payload)
			done()
			return nil
		})
	})

	// PipeMap with dead-letter.
	_ = testing.GoAsync(ctx, func(ctx context.Context, done context.CancelFunc) {
		_ = srcSub.Subscribe(ctx, "events", messaging.PipeMap[Event, Summary](dstPub, mapFn, messaging.WithDeadLetter[Event](deadLetter)))
	})

	// First message fails the map — routed to dead-letter.
	if err := srcPub.Publish(ctx, "events", Event{ID: "bad", Name: "broken"}); err != nil {
		panic(err)
	}
	// Second message succeeds — arrives at destination.
	if err := srcPub.Publish(ctx, "events", Event{ID: "1", Name: "hello"}); err != nil {
		panic(err)
	}

	<-await
	// Output:
	// dead-letter: events {bad broken} bad event
	// events {hello}
}
