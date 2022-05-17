package main

import (
	"net/http"

	"github.com/davecgh/go-spew/spew"

	"github.com/foomo/keel"
)

func main() {
	svr := keel.NewServer(
		keel.WithHTTPZapService(true),
		keel.WithRemoteConfig("etcd", "http://localhost:2379", "example.yaml"),
	)

	// obtain the logger
	l := svr.Logger()
	c := svr.Config()

	spew.Dump(c.AllSettings())

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		spew.Dump(c.AllSettings())
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	svr.AddService(
		keel.NewServiceHTTP(l, "demo", "localhost:8080", svs),
	)

	svr.Run()
}