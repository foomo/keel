package service_test

import (
	"net/http"
	"time"

	"github.com/foomo/keel"
	"github.com/foomo/keel/service"
	"go.uber.org/zap"
)

func ExampleNewHTTP() {
	shutdown(3 * time.Second)

	svr := keel.NewServer(
		keel.WithLogger(zap.NewExample()),
	)

	l := svr.Logger()
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

	// Output:
	// {"level":"info","msg":"starting keel server"}
	// {"level":"info","msg":"starting keel service","keel_service_type":"http","keel_service_name":"demo","net_host_ip":"localhost","net_host_port":"8080"}
	// {"level":"debug","msg":"keel graceful shutdown"}
	// {"level":"info","msg":"stopping keel service","keel_service_type":"http","keel_service_name":"demo"}
	// {"level":"info","msg":"keel server stopped"}
}
