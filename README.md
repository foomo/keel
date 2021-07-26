# keel

[![Go Report Card](https://goreportcard.com/badge/github.com/foomo/keel)](https://goreportcard.com/report/github.com/foomo/keel)
[![godoc](https://godoc.org/github.com/foomo/keel?status.svg)](https://godoc.org/github.com/foomo/keel)
[![GitHub Super-Linter](https://github.com/foomo/keel/workflows/CI/badge.svg)](https://github.com/marketplace/actions/super-linter)


> Opinionated way to run services.

## Stack

- Metrics:        Prometheus
- Logging:        Zap
- Telemetry:      Open Telemetry
- Configuration:  Viper

## Example

```go
package main

import (
	"net/http"
	"os"

	"github.com/foomo/keel"
	"github.com/foomo/keel/log"
)

func main() {
	svr := keel.NewServer()

	l := svr.Logger()

	// add zap service listening on localhost:9100
	// allows you to view / change the log level: GET / PUT localhost:9100/log
	svr.AddServices(keel.NewDefaultServiceHTTPZap())

	// add viper service listening on localhost:9300
	// allows you to view / change the configuration: GET / PUT localhost:9300/config
	svr.AddServices(keel.NewDefaultServiceHTTPViper())

	// add prometheus service listening on 0.0.0.0:9200
	// allows you to collect prometheus metrics: GET 0.0.0.0:9200/metrics
	svr.AddServices(keel.NewDefaultServiceHTTPPrometheus())

	svr.AddService(
		keel.NewServiceHTTP(log.WithServiceName(l, "demo"), ":8080", newService()),
	)

	svr.Run()
}

func newService() *http.ServeMux {
	s := http.NewServeMux()
	s.HandleFunc("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("exiting")
	})
	s.HandleFunc("/exit", func(w http.ResponseWriter, r *http.Request) {
		os.Exit(1)
	})
	return s
}

```

## How to Contribute

Make a pull request...

## License

Distributed under MIT License, please see license file within the code for more details.
