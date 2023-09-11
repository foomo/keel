package service_test

import (
	"io"
	"net"
	"net/http"
	"time"

	"github.com/foomo/keel"
	"github.com/foomo/keel/service"
	"go.uber.org/zap"
)

func ExampleNewHTTP() {
	svr := keel.NewServer(
		keel.WithLogger(zap.NewExample()),
	)

	l := svr.Logger()

	svr.AddService(
		service.NewHTTP(l, "demo", "localhost:8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		})),
	)

	go func() {
		if _, err := net.DialTimeout("tcp", "localhost:8080", 10*time.Second); err != nil {
			panic(err.Error())
		}
		resp, _ := http.Get("http://localhost:8080") //nolint:noctx
		defer resp.Body.Close()                      //nolint:govet
		b, _ := io.ReadAll(resp.Body)
		l.Info(string(b))
		shutdown()
	}()

	svr.Run()

	// Output:
	// {"level":"info","msg":"starting keel server"}
	// {"level":"info","msg":"starting keel service","keel_service_type":"http","keel_service_name":"demo","net_host_ip":"localhost","net_host_port":"8080"}
	// {"level":"info","msg":"OK"}
	// {"level":"debug","msg":"keel graceful shutdown"}
	// {"level":"info","msg":"stopping keel service","keel_service_type":"http","keel_service_name":"demo"}
	// {"level":"info","msg":"keel server stopped"}
}
