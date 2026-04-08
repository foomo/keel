package nats_test

// func ExampleNewPublisher() {
// 	srv, conn := startNATSServer()
// 	defer srv.Shutdown()
// 	defer conn.Close()
//
// 	// Subscribe via raw NATS to observe the published message.
// 	ch := make(stream []byte, 1)
// 	sub, err := conn.Subscribe("events", func(msg *nats.Msg) {
// 		ch <- msg.Data
// 	})
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer sub.Unsubscribe()
//
// 	pub := nats.NewPublisher(conn, codec.NewJSON[Event]())
// 	if err := pub.Publish(context.Background(), "events", Event{ID: "1", Name: "foo"}); err != nil {
// 		panic(err)
// 	}
//
// 	fmt.Println(string(<-ch))
//
// 	// Output: {"id":"1","name":"foo"}
// }
