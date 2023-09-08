package service_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/foomo/keel"
	"github.com/foomo/keel/config"
	"github.com/foomo/keel/service"
	"go.uber.org/zap"
)

func ExampleNewHTTPDocs() {
	// define vars so it does not panic
	_ = os.Setenv("EXAMPLE_REQUIRED_BOOL", "true")
	_ = os.Setenv("EXAMPLE_REQUIRED_STRING", "foo")

	svr := keel.NewServer(
		keel.WithLogger(zap.NewNop()),
		keel.WithHTTPDocsService(true),
	)

	l := svr.Logger()

	c := svr.Config()
	// config with fallback
	_ = config.GetBool(c, "example.bool", false)
	_ = config.GetString(c, "example.string", "fallback")
	// required configs
	_ = config.MustGetBool(c, "example.required.bool")
	_ = config.MustGetString(c, "example.required.string")

	svr.AddService(service.NewHTTP(l, "demp-http", "localhost:8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})))

	svr.AddService(service.NewGoRoutine(l, "demo-goroutine", func(ctx context.Context, l *zap.Logger) error {
		return nil
	}))

	go func() {
		resp, _ := http.Get("http://localhost:9001/docs") //nolint:noctx
		defer resp.Body.Close()                           //nolint:govet
		b, _ := io.ReadAll(resp.Body)
		fmt.Print(string(b))
		shutdown()
	}()

	svr.Run()

	// Output:
	// ## Keel Server
	//
	// ### Config
	//
	// List of all registered config variabled with their defaults.
	//
	// | Key                       | Type     | Required | Default    |
	// | ------------------------- | -------- | -------- | ---------- |
	// | `example.bool`            | `bool`   |          | `false`    |
	// | `example.required.bool`   | `bool`   | `true`   |            |
	// | `example.required.string` | `string` | `true`   |            |
	// | `example.string`          | `string` |          | `fallback` |
	// | `service.docs.enabled`    | `bool`   |          | `true`     |
	//
	// ### Init Services
	//
	// List of all registered init services that are being immediately started.
	//
	// | Name   | Type            | Address                              |
	// | ------ | --------------- | ------------------------------------ |
	// | `docs` | `*service.HTTP` | `*http.ServeMux` on `localhost:9001` |
	//
	// ### Services
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
	// List of all registered healthz probes that are being called during startup and runntime.
	//
	// | Name     | Type                 | Description                            |
	// | -------- | -------------------- | -------------------------------------- |
	// | `always` | `*keel.Server`       |                                        |
	// | `always` | `*service.GoRoutine` | parallel: `1`                          |
	// | `always` | `*service.HTTP`      | `*http.ServeMux` on `localhost:9001`   |
	// | `always` | `*service.HTTP`      | `http.HandlerFunc` on `localhost:8080` |
	//
	// ### Closers
	//
	// List of all registered closers that are being called during graceful shutdown.
	//
	// | Name                 | Type                     |
	// | -------------------- | ------------------------ |
	// | `*service.GoRoutine` | `ErrorCloserWithContext` |
	// | `*service.HTTP`      | `ErrorCloserWithContext` |
	// | `*service.HTTP`      | `ErrorCloserWithContext` |
	//
	// ### Metrics
	//
	// List of all registered metrics than are being exposed.
	//
	// | Name                               | Type    | Description                                                        |
	// | ---------------------------------- | ------- | ------------------------------------------------------------------ |
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
