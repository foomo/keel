package main

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/foomo/keel/service"
	"go.uber.org/zap"

	"github.com/foomo/keel"
	"github.com/foomo/keel/log"
)

func main() {
	svr := keel.NewServer(
		keel.WithHTTPZapService(true),
		keel.WithHTTPViperService(true),
		keel.WithHTTPPrometheusService(true),
		keel.WithHTTPHealthzService(true),
	)

	l := svr.Logger()

	go waitGroup(svr.CancelContext(), l.With(log.FServiceName("waitGroup")))

	// create demo service
	svs := http.NewServeMux()
	svs.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	svr.AddService(
		service.NewHTTP(l, "demo", "localhost:8080", svs),
	)

	svr.Run()
}

func waitGroup(ctx context.Context, l *zap.Logger) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				l.Info("Break the loop")
				return
			case <-time.After(3 * time.Second):
				l.Info("Hello in a loop")
			}
		}
	}()

	wg.Wait()
}
