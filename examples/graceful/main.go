package main

import (
	"context"
	"net/http"
	"sync"
	"syscall"
	"time"

	"github.com/foomo/keel/interfaces"
	"github.com/foomo/keel/service"
	"go.uber.org/zap"

	"github.com/foomo/keel"
	"github.com/foomo/keel/log"
)

func main() {
	svr := keel.NewServer(
		//keel.WithLogger(zap.NewExample()),
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
		l.Info("handling request...")
		time.Sleep(3 * time.Second)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
		l.Info("... handled request")
	})

	svr.AddService(
		service.NewHTTP(l, "demo", "localhost:8080", svs),
	)

	svr.AddCloser(interfaces.CloseFunc(func(ctx context.Context) error {
		l.Info("custom closer")
		return nil
	}))

	go svr.Run()
	time.Sleep(1 * time.Second)
	l.Info("1. starting test")

	{
		l.Info("2. checking healthz")
		readiness(l, "http://localhost:9400/healthz/readiness")
	}

	go func() {
		l.Info("2. sending request")
		if r, err := http.Get("http://localhost:8080"); err != nil {
			l.Fatal(err.Error())
		} else {
			l.Info("  /", zap.Int("status", r.StatusCode))
		}
	}()
	time.Sleep(100 * time.Millisecond)

	l.Info("3. sending shutdown signal")
	if err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM); err != nil {
		l.Fatal(err.Error())
	}

	{
		l.Info("2. checking healthz")
		readiness(l, "http://localhost:9400/healthz/readiness")
	}

	l.Info("4. waiting for shutdown")
	time.Sleep(10 * time.Second)
	l.Info("  done")
}

func readiness(l *zap.Logger, url string) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		l.Error(err.Error())
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		l.Error(err.Error())
		return
	}
	l.Info(url, zap.Int("status", resp.StatusCode))
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
