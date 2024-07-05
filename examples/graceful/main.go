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

	go func() {
		c := make(chan bool, 1)
		time.Sleep(2 * time.Second)

		l.Info("1. starting checks")
		go func() {
			c <- true
			for {
				call(l.Named("http"), "http://localhost:8080")
				call(l.Named("readiness"), "http://localhost:9400/healthz/readiness")
				time.Sleep(time.Second)
			}
		}()
		<-c
		close(c)

		l.Info("2. sleeping for 5 seconds")
		time.Sleep(5 * time.Second)

		l.Info("3. sending shutdown signal")
		if err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM); err != nil {
			l.Fatal(err.Error())
		}
	}()

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

	svr.Run()
	l.Info("done")

	// Output:
	// {"level":"info","logger":"root.server","msg":"starting keel server"}
	// {"level":"info","logger":"root.server","msg":"starting keel service","keel_service_type":"http","keel_service_name":"healthz","net_host_ip":"localhost","net_host_port":"9400"}
	// {"level":"info","logger":"root","msg":"starting keel service","keel_service_type":"http","keel_service_name":"http","net_host_ip":"localhost","net_host_port":"8080"}
	// {"level":"info","logger":"root","msg":"1. starting checks"}
	// {"level":"info","logger":"root","msg":"2. sleeping for 5 seconds"}
	// {"level":"info","logger":"root.http","msg":"ok","url":"http://localhost:8080","status":200}
	// {"level":"info","logger":"root.readiness","msg":"ok","url":"http://localhost:9400/healthz/readiness","status":200}
	// {"level":"info","logger":"root.http","msg":"ok","url":"http://localhost:8080","status":200}
	// {"level":"info","logger":"root.readiness","msg":"ok","url":"http://localhost:9400/healthz/readiness","status":200}
	// {"level":"info","logger":"root.http","msg":"ok","url":"http://localhost:8080","status":200}
	// {"level":"info","logger":"root.readiness","msg":"ok","url":"http://localhost:9400/healthz/readiness","status":200}
	// {"level":"info","logger":"root.http","msg":"ok","url":"http://localhost:8080","status":200}
	// {"level":"info","logger":"root.readiness","msg":"ok","url":"http://localhost:9400/healthz/readiness","status":200}
	// {"level":"info","logger":"root.http","msg":"ok","url":"http://localhost:8080","status":200}
	// {"level":"info","logger":"root.readiness","msg":"ok","url":"http://localhost:9400/healthz/readiness","status":200}
	// {"level":"info","logger":"root","msg":"3. sending shutdown signal"}
	// {"level":"info","logger":"root.server","msg":"keel graceful shutdown","graceful_period":"30s"}
	// {"level":"info","logger":"root.server","msg":"keel graceful shutdown: closers"}
	// {"level":"info","logger":"root","msg":"stopping keel service","keel_service_type":"http","keel_service_name":"http"}
	// {"level":"debug","logger":"root.server","msg":"keel graceful shutdown: closer closed","name":"*service.HTTP"}
	// {"level":"info","logger":"root.closer","msg":"closing stuff"}
	// {"level":"error","logger":"root.http","msg":"failed to send request","url":"http://localhost:8080","error":"Get \"http://localhost:8080\": dial tcp [::1]:8080: connect: connection refused"}
	// {"level":"debug","logger":"root.server","msg":"healthz probe failed","error_type":"*errors.errorString","error_message":"service not running","http_target":"/healthz/readiness"}
	// {"level":"info","logger":"root.readiness","msg":"ok","url":"http://localhost:9400/healthz/readiness","status":503}
	// {"level":"error","logger":"root.http","msg":"failed to send request","url":"http://localhost:8080","error":"Get \"http://localhost:8080\": dial tcp [::1]:8080: connect: connection refused"}
	// {"level":"debug","logger":"root.server","msg":"healthz probe failed","error_type":"*errors.errorString","error_message":"service not running","http_target":"/healthz/readiness"}
	// {"level":"info","logger":"root.readiness","msg":"ok","url":"http://localhost:9400/healthz/readiness","status":503}
	// {"level":"error","logger":"root.http","msg":"failed to send request","url":"http://localhost:8080","error":"Get \"http://localhost:8080\": dial tcp [::1]:8080: connect: connection refused"}
	// {"level":"debug","logger":"root.server","msg":"healthz probe failed","error_type":"*errors.errorString","error_message":"service not running","http_target":"/healthz/readiness"}
	// {"level":"info","logger":"root.readiness","msg":"ok","url":"http://localhost:9400/healthz/readiness","status":503}
	// {"level":"info","logger":"root.closer","msg":"done closing stuff"}
	// {"level":"debug","logger":"root.server","msg":"keel graceful shutdown: closer closed","name":"interfaces.CloserFunc"}
	// {"level":"info","logger":"root.server","msg":"stopping keel service","keel_service_type":"http","keel_service_name":"healthz"}
	// {"level":"debug","logger":"root.server","msg":"keel graceful shutdown: closer closed","name":"*service.HTTP"}
	// {"level":"debug","logger":"root.server","msg":"keel graceful shutdown: closer closed","name":"noop.TracerProvider"}
	// {"level":"debug","logger":"root.server","msg":"keel graceful shutdown: closer closed","name":"noop.MeterProvider"}
	// {"level":"info","logger":"root.server","msg":"keel graceful shutdown: complete"}
	// {"level":"info","logger":"root.server","msg":"keel server stopped"}
	// {"level":"info","logger":"root","msg":"done"}
}

func call(l *zap.Logger, url string) {
	l = l.With(zap.String("url", url))

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
}
