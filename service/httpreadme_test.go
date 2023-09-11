package service_test

import (
	"context"
	"net/http"
	"os"

	"github.com/foomo/keel"
	"github.com/foomo/keel/config"
	"github.com/foomo/keel/env"
	"github.com/foomo/keel/service"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.uber.org/zap"
)

func ExampleNewHTTPReadme() {
	// define vars so it does not panic
	_ = os.Setenv("EXAMPLE_REQUIRED_BOOL", "true")
	_ = os.Setenv("EXAMPLE_REQUIRED_STRING", "foo")

	svr := keel.NewServer(
		keel.WithLogger(zap.NewExample()),
		keel.WithPrometheusMeter(true),
		keel.WithHTTPReadmeService(true),
	)

	// access some env vars
	_ = env.Get("EXAMPLE_STRING", "demo")
	_ = env.GetBool("EXAMPLE_BOOL", false)
	_ = env.MustGet("EXAMPLE_REQUIRED_STRING")
	_ = env.MustGetBool("EXAMPLE_REQUIRED_BOOL")

	l := svr.Logger()

	c := svr.Config()
	// config with fallback
	_ = config.GetBool(c, "example.bool", false)
	_ = config.GetString(c, "example.string", "fallback")
	// required configs
	_ = config.MustGetBool(c, "example.required.bool")
	_ = config.MustGetString(c, "example.required.string")

	m := svr.Meter()

	// add metrics
	fooBarCounter := promauto.NewCounter(prometheus.CounterOpts{
		Name: "foo_bar_total",
		Help: "Foo bar metrics",
	})
	fooBazCounter, _ := m.SyncInt64().Counter("foo_baz_total", instrument.WithDescription("Foo baz metrics"))

	fooBarCounter.Add(1)
	fooBazCounter.Add(svr.Context(), 1)

	// add http service
	svr.AddService(service.NewHTTP(l, "demp-http", "localhost:8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})))

	// add go routine service
	svr.AddService(service.NewGoRoutine(l, "demo-goroutine", func(ctx context.Context, l *zap.Logger) error {
		return nil
	}))

	go func() {
		waitFor("localhost:9001")
		l.Info(httpGet("http://localhost:9001/readme"))
		shutdown()
	}()

	svr.Run()

	// Output:
	// ### Env
	//
	// List of all accessed environment variables.
	//
	// | Key                       | Type     | Required | Default   |
	// | ------------------------- | -------- | -------- | --------- |
	// | `EXAMPLE_BOOL`            | `bool`   |          |           |
	// | `EXAMPLE_REQUIRED_BOOL`   | `bool`   |          |           |
	// | `EXAMPLE_REQUIRED_BOOL`   | `bool`   | `true`   |           |
	// | `EXAMPLE_REQUIRED_STRING` | `string` |          |           |
	// | `EXAMPLE_REQUIRED_STRING` | `string` | `true`   |           |
	// | `EXAMPLE_STRING`          | `string` |          | `demo`    |
	// | `LOG_DISABLE_CALLER`      | `bool`   |          |           |
	// | `LOG_DISABLE_STACKTRACE`  | `bool`   |          |           |
	// | `LOG_ENCODING`            | `string` |          | `json`    |
	// | `LOG_LEVEL`               | `string` |          | `info`    |
	// | `LOG_MODE`                | `string` |          | `prod`    |
	// | `OTEL_ENABLED`            | `bool`   |          |           |
	// | `OTEL_SERVICE_NAME`       | `string` |          | `service` |
	//
	// ### Config
	//
	// List of all registered config variables with their defaults.
	//
	// | Key                       | Type     | Required | Default    |
	// | ------------------------- | -------- | -------- | ---------- |
	// | `example.bool`            | `bool`   |          | `false`    |
	// | `example.required.bool`   | `bool`   | `true`   |            |
	// | `example.required.string` | `string` | `true`   |            |
	// | `example.string`          | `string` |          | `fallback` |
	// | `otel.enabled`            | `bool`   |          | `true`     |
	// | `service.readme.enabled`  | `bool`   |          | `true`     |
	//
	// ### Init Services
	//
	// List of all registered init services that are being immediately started.
	//
	// | Name     | Type            | Address                              |
	// | -------- | --------------- | ------------------------------------ |
	// | `readme` | `*service.HTTP` | `*http.ServeMux` on `localhost:9001` |
	//
	// ### Runtime Services
	//
	// List of all registered services that are being started.
	//
	// | Name             | Type                 | Description                            |
	// | ---------------- | -------------------- | -------------------------------------- |
	// | `demo-goroutine` | `*service.GoRoutine` | parallel: `1`                          |
	// | `demp-http`      | `*service.HTTP`      | `http.HandlerFunc` on `localhost:8080` |
	//
	// ### Health probes
	//
	// List of all registered healthz probes that are being called during startup and runtime.
	//
	// | Name             | Probe    | Type                 | Description                            |
	// | ---------------- | -------- | -------------------- | -------------------------------------- |
	// |                  | `always` | `*keel.Server`       |                                        |
	// | `demo-goroutine` | `always` | `*service.GoRoutine` | parallel: `1`                          |
	// | `demp-http`      | `always` | `*service.HTTP`      | `http.HandlerFunc` on `localhost:8080` |
	// | `readme`         | `always` | `*service.HTTP`      | `*http.ServeMux` on `localhost:9001`   |
	//
	// ### Closers
	//
	// List of all registered closers that are being called during graceful shutdown.
	//
	// | Name             | Type                 | Closer                   | Description                            |
	// | ---------------- | -------------------- | ------------------------ | -------------------------------------- |
	// | `demo-goroutine` | `*service.GoRoutine` | `ErrorCloserWithContext` | parallel: `1`                          |
	// | `demp-http`      | `*service.HTTP`      | `ErrorCloserWithContext` | `http.HandlerFunc` on `localhost:8080` |
	// | `readme`         | `*service.HTTP`      | `ErrorCloserWithContext` | `*http.ServeMux` on `localhost:9001`   |
	//
	// ### Metrics
	//
	// List of all registered metrics than are being exposed.
	//
	// | Name                               | Type    | Description                                                        |
	// | ---------------------------------- | ------- | ------------------------------------------------------------------ |
	// | `foo_bar_total`                    | COUNTER | Foo bar metrics                                                    |
	// | `foo_baz_total`                    | COUNTER | Foo baz metrics                                                    |
	// | `go_gc_duration_seconds`           | SUMMARY | A summary of the pause duration of garbage collection cycles.      |
	// | `go_goroutines`                    | GAUGE   | Number of goroutines that currently exist.                         |
	// | `go_info`                          | GAUGE   | Information about the Go environment.                              |
	// | `go_memstats_alloc_bytes_total`    | COUNTER | Total number of bytes allocated, even if freed.                    |
	// | `go_memstats_alloc_bytes`          | GAUGE   | Number of bytes allocated and still in use.                        |
	// | `go_memstats_buck_hash_sys_bytes`  | GAUGE   | Number of bytes used by the profiling bucket hash table.           |
	// | `go_memstats_frees_total`          | COUNTER | Total number of frees.                                             |
	// | `go_memstats_gc_sys_bytes`         | GAUGE   | Number of bytes used for garbage collection system metadata.       |
	// | `go_memstats_heap_alloc_bytes`     | GAUGE   | Number of heap bytes allocated and still in use.                   |
	// | `go_memstats_heap_idle_bytes`      | GAUGE   | Number of heap bytes waiting to be used.                           |
	// | `go_memstats_heap_inuse_bytes`     | GAUGE   | Number of heap bytes that are in use.                              |
	// | `go_memstats_heap_objects`         | GAUGE   | Number of allocated objects.                                       |
	// | `go_memstats_heap_released_bytes`  | GAUGE   | Number of heap bytes released to OS.                               |
	// | `go_memstats_heap_sys_bytes`       | GAUGE   | Number of heap bytes obtained from system.                         |
	// | `go_memstats_last_gc_time_seconds` | GAUGE   | Number of seconds since 1970 of last garbage collection.           |
	// | `go_memstats_lookups_total`        | COUNTER | Total number of pointer lookups.                                   |
	// | `go_memstats_mallocs_total`        | COUNTER | Total number of mallocs.                                           |
	// | `go_memstats_mcache_inuse_bytes`   | GAUGE   | Number of bytes in use by mcache structures.                       |
	// | `go_memstats_mcache_sys_bytes`     | GAUGE   | Number of bytes used for mcache structures obtained from system.   |
	// | `go_memstats_mspan_inuse_bytes`    | GAUGE   | Number of bytes in use by mspan structures.                        |
	// | `go_memstats_mspan_sys_bytes`      | GAUGE   | Number of bytes used for mspan structures obtained from system.    |
	// | `go_memstats_next_gc_bytes`        | GAUGE   | Number of heap bytes when next garbage collection will take place. |
	// | `go_memstats_other_sys_bytes`      | GAUGE   | Number of bytes used for other system allocations.                 |
	// | `go_memstats_stack_inuse_bytes`    | GAUGE   | Number of bytes in use by the stack allocator.                     |
	// | `go_memstats_stack_sys_bytes`      | GAUGE   | Number of bytes obtained from system for stack allocator.          |
	// | `go_memstats_sys_bytes`            | GAUGE   | Number of bytes obtained from system.                              |
	// | `go_threads`                       | GAUGE   | Number of OS threads created.                                      |
	// | `process_cpu_seconds_total`        | COUNTER | Total user and system CPU time spent in seconds.                   |
	// | `process_max_fds`                  | GAUGE   | Maximum number of open file descriptors.                           |
	// | `process_open_fds`                 | GAUGE   | Number of open file descriptors.                                   |
	// | `process_resident_memory_bytes`    | GAUGE   | Resident memory size in bytes.                                     |
	// | `process_start_time_seconds`       | GAUGE   | Start time of the process since unix epoch in seconds.             |
	// | `process_virtual_memory_bytes`     | GAUGE   | Virtual memory size in bytes.                                      |
	// | `process_virtual_memory_max_bytes` | GAUGE   | Maximum amount of virtual memory available in bytes.               |
}
