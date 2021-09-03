package main

import (
	"net/http"
	"os"

	"github.com/foomo/keel"
	"github.com/foomo/keel/log"
	"github.com/foomo/keel/net/http/middleware"
)

func main() {
	svr := keel.NewServer()
	l := svr.Logger()

	domains := []string{"*.example.com"}

	domainMapping := map[string]string{
		"foo.example.com": "bar.example.com",
	}

	domainProvider := middleware.MappingDomainProvider(domains, domainMapping)

	svr.AddService(
		keel.NewServiceHTTP(
			log.WithServiceName(l, "demo"),
			":8080",
			newService(),
			middleware.SessionID(
				middleware.SessionIDWithCookieDomainProvider(
					domainProvider,
				)
			)
		),
	)

	svr.Run()
}

func newService() *http.ServeMux {
	s := http.NewServeMux()
	s.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
    w.Write([]byte("Hello World!"))
	})
	return s
}
