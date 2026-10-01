// Package service provides ready made [github.com/foomo/keel.Service] implementations
// for running HTTP servers, goroutines and message subscriptions under a keel
// server.
//
// [HTTP] wraps a [net/http.Server]; [GoRoutine] runs a function in one or more
// goroutines; [Subscription] runs a [goflux.Subscriber]. All of them implement
// Healthz and Close(ctx) error so they double as health probes and closers.
//
// The package also provides constructors for keel's built-in internal HTTP
// services, each with a NewDefaultHTTPXxx variant using the DefaultHTTPXxx
// name, address and path variables:
//
//   - [NewHealthz]: Kubernetes health probes (default :9400/healthz)
//   - [NewHTTPPrometheus]: Prometheus metrics (default :9200/metrics)
//   - [NewHTTPPProf]: pprof profiling (default localhost:6060/debug/pprof)
//   - [NewHTTPZap]: runtime log level control (default localhost:9100/log)
//   - [NewHTTPViper]: runtime configuration access (default localhost:9300/config)
//   - [NewHTTPReadme]: markdown readme (default localhost:9001/readme, deprecated)
package service
