package main

import (
	"fmt"
	"net/http"

	"github.com/foomo/keel"
	"github.com/foomo/keel/config"
	"github.com/foomo/keel/service"
)

func main() {
	svr := keel.NewServer(
		// configure remote endpoint
		keel.WithRemoteConfig("etcd3", "http://localhost:2379", "cluster.yaml"),
	)

	// obtain the logger
	l := svr.Logger()
	c := svr.Config()

	// dump all settings
	// spew.Dump(c.AllSettings())

	// create config reader
	fooFn := config.GetString(c, "foo", "default_foo")
	fmt.Println("initial foo:", fooFn()) //nolint:forbidigo

	// watch changes
	config.WatchString(svr.Context(), fooFn, func(s string) {
		fmt.Println("change foo:", fooFn()) //nolint:forbidigo
	})

	ch := make(chan string)
	// watch changes
	config.WatchStringChan(svr.Context(), fooFn, ch)
	go func(ch chan string) {
		for {
			value := <-ch
			fmt.Println("channel foo:", value) //nolint:forbidigo
		}
	}(ch)

	// curl localhost:8080
	svr.AddService(
		service.NewHTTP(l, "demo", "localhost:8080", http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Println("current foo:", fooFn()) //nolint:forbidigo
			}),
		),
	)

	svr.Run()
}
