package nats_test

// func ExampleNewSubscriber() {
// 	srv, conn := startNATSServer()
// 	defer srv.Shutdown()
// 	defer conn.Close()
//
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()
//
// 	done := make(stream struct{})
// 	sub := nats.NewSubscriber(conn, codec.NewJSON[Event]())
// 	go func() {
// 		defer close(done)
// 		_ = sub.Subscribe(ctx, "events", func(_ context.Context, msg messaging.Message[Event]) error {
// 			fmt.Println(msg)
// 			cancel()
// 			return nil
// 		})
// 	}()
//
// 	// Wait for subscription to be ready, then publish raw JSON.
// 	conn.Flush()
// 	if err := conn.Publish("events", []byte(`{"id":"1","name":"foo"}`)); err != nil {
// 		panic(err)
// 	}
//
// 	<-done
//
// 	// Output: {events {1 foo}}
// }
