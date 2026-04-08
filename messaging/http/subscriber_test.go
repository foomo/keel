package http_test

//
// func ExampleNewSubscriber() {
// 	c := codec.NewJSON[Event]()
//
// 	sub := http.NewSubscriber(c)
// 	go func() {
// 		if err := sub.Subscribe(context.Background(), "http", func(ctx context.Context, msg messaging.Message[Event]) error {
// 			fmt.Println(msg)
// 			return nil
// 		}); err != nil {
// 			panic(err)
// 		}
// 	}()
//
// 	s := httptest.NewServer(sub.Mux())
// 	e, err := json.Marshal(Event{ID: "1", Name: "foo"})
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	_, err = http.Post(s.URL+"/http", "application/json", bytes.NewReader(e))
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	// Output: {http {1 foo}}
// }
