package main

import (
	"context"
	"net/http"
	"syscall"
	"time"

	"github.com/foomo/keel/interfaces"
	"github.com/foomo/keel/service"
	"go.uber.org/zap"

	"github.com/foomo/keel"
)

func main() {
	service.DefaultHTTPHealthzAddr = "localhost:9400"

	l := zap.NewExample().Named("root")

	l.Info("1. starting readiness checks")
	go call(l.Named("readiness"), "http://localhost:9400/healthz/readiness")

	svr := keel.NewServer(
		keel.WithLogger(l.Named("server")),
		keel.WithHTTPHealthzService(true),
	)

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	})

	svr.AddService(
		service.NewHTTP(l, "http", "localhost:8080", svs),
	)

	svr.AddCloser(interfaces.CloserFunc(func(ctx context.Context) error {
		l := l.Named("closer")
		l.Info("closing stuff")
		time.Sleep(3 * time.Second)
		l.Info("done closing stuff")
		return nil
	}))

	go func() {

		l.Info("3. starting http checks")
		go call(l.Named("http"), "http://localhost:8080")

		l.Info("4. sleeping for 5 seconds")
		time.Sleep(5 * time.Second)

		l.Info("5. sending shutdown signal")
		if err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM); err != nil {
			l.Fatal(err.Error())
		}

	}()

	svr.Run()
	l.Info("done")
}

func call(l *zap.Logger, url string) {
	l = l.With(zap.String("url", url))
	for {
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				l.With(zap.Error(err)).Error("failed to create request")
				return
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				l.With(zap.Error(err)).Error("failed to send request")
				return
			}
			l.Info("ok", zap.Int("status", resp.StatusCode))
		}()
		time.Sleep(time.Second)
	}
}
