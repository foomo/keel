package http_test

import (
	"context"
	"fmt"
	"io"
	stdhttp "net/http"
	"net/http/httptest"

	"github.com/foomo/keel/messaging/http"
	"github.com/foomo/keel/pkg/encoding/json"
)

func ExampleNewPublisher() {
	c := json.NewCodec[Event]()

	s := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		req, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(req))
	}))

	pub := http.NewPublisher(s.URL, c, s.Client())
	if err := pub.Publish(context.Background(), "http", Event{ID: "1", Name: "foo"}); err != nil {
		panic(err)
	}

	// Output: {"id":"1","name":"foo"}
}
