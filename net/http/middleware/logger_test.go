package middleware_test

import (
	"fmt"
	"net/http"

	"github.com/foomo/keel/log"
	httplog "github.com/foomo/keel/net/http/log"
	"github.com/foomo/keel/net/http/middleware"
	keeltest "github.com/foomo/keel/test"
	"go.uber.org/zap"
)

func ExampleLogger() {
	svr := keeltest.NewServer()

	// get logger
	l := svr.Logger()

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
		fmt.Println("ok")
	})

	svr.AddService(
		keeltest.NewServiceHTTP(l, "demo", svs,
			middleware.Logger(),
		),
	)

	svr.Start()

	resp, err := http.Get(svr.GetService("demo").URL() + "/") //nolint:noctx
	log.Must(l, err)
	defer resp.Body.Close()

	// Output: ok
}

func ExampleLoggerWithInjectLabeler() {
	svr := keeltest.NewServer()

	// get logger
	l := svr.Logger()

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
		fmt.Println("ok")
	})

	svr.AddService(
		keeltest.NewServiceHTTP(l, "demo", svs,
			func(l *zap.Logger, s string, next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if labeler, ok := httplog.LabelerFromRequest(r); ok {
						labeler.Add(zap.String("injected", "message"))
					}
					next.ServeHTTP(w, r)
				})
			},
			middleware.Logger(
				middleware.LoggerWithInjectLabeler(true),
			),
		),
	)

	svr.Start()

	resp, err := http.Get(svr.GetService("demo").URL() + "/") //nolint:noctx
	log.Must(l, err)
	defer resp.Body.Close()

	// Output: ok
}
