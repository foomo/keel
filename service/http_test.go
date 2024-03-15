package service_test

import (
	"net/http"
	"time"

	"github.com/foomo/keel"
	"github.com/foomo/keel/service"
	"go.uber.org/zap"
)

func ExampleNewHTTP() {
	svr := keel.NewServer(
		keel.WithLogger(zap.NewExample()),
		keel.WithGracefulTimeout(time.Second),
		keel.WithShutdownTimeout(10*time.Second),
	)

	l := svr.Logger()

	svr.AddService(
		service.NewHTTP(l, "demo", "localhost:8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		})),
	)

	go func() {
		waitFor("localhost:8080")
		l.Info(httpGet("http://localhost:8080"))
		shutdown()
	}()

	svr.Run()

	// Output:
	// {"level":"info","msg":"starting keel server"}
	// {"level":"info","msg":"starting keel service","keel_service_type":"http","keel_service_name":"demo","net_host_ip":"localhost","net_host_port":"8080"}
	// {"level":"info","msg":"OK"}
	// {"level":"info","msg":"keel graceful shutdown"}
	// {"level":"info","msg":"keel graceful shutdown timeout","graceful_timeout":"1s","shutdown_timeout":"10s"}
	// {"level":"info","msg":"keel graceful shutdown timeout complete"}
	// {"level":"info","msg":"keel graceful shutdown closers"}
	// {"level":"info","msg":"stopping keel service","keel_service_type":"http","keel_service_name":"demo"}
	// {"level":"debug","msg":"keel graceful shutdown closer closed","name":"*service.HTTP"}
	// {"level":"debug","msg":"keel graceful shutdown closer closed","name":"trace.noopTracerProvider"}
	// {"level":"debug","msg":"keel graceful shutdown closer closed","name":"nonrecording.noopMeterProvider"}
	// {"level":"info","msg":"keel graceful shutdown complete"}
	// {"level":"info","msg":"keel server stopped"}
}
