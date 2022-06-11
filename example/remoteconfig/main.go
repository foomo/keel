package main

import (
	"fmt"
	"net/http"

	"github.com/foomo/keel"
	"github.com/foomo/keel/config"
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
	fmt.Println("initial foo:", fooFn())

	// watch changes
	config.WatchString(svr.CancelContext(), fooFn, func(s string) {
		fmt.Println("change foo:", fooFn())
	})

	ch := make(chan string)
	// watch changes
	config.WatchStringChan(svr.CancelContext(), fooFn, ch)
	go func(ch chan string) {
		for {
			value := <-ch
			fmt.Println("channel foo:", value)
		}
	}(ch)

	// curl localhost:8080
	svr.AddService(
		keel.NewServiceHTTP(l, "demo", "localhost:8080", http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Println("current foo:", fooFn())
			}),
		),
	)

	svr.Run()
}
