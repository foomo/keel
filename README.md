# keel

[![Go Report Card](https://goreportcard.com/badge/github.com/foomo/keel)](https://goreportcard.com/report/github.com/foomo/keel)
[![godoc](https://godoc.org/github.com/foomo/keel?status.svg)](https://godoc.org/github.com/foomo/keel)
[![GitHub Super-Linter](https://github.com/foomo/keel/workflows/CI/badge.svg)](https://github.com/marketplace/actions/super-linter)

> Opinionated way to run services.

## Stack

- Zap
- Viper
- Open Telemetry
- Nats
- GoTSRPC

## Examples

See the examples folder for usages

```go
package main

import (
  "net/http"

  "github.com/foomo/keel"
  "github.com/foomo/keel/service"
)

func main() {
  svr := keel.NewServer(
    keel.WithHTTPZapService(true),
    keel.WithHTTPViperService(true),
    keel.WithHTTPPrometheusService(true),
  )

  l := svr.Logger()

  svs := newService()

  svr.AddService(
    service.NewHTTP(l, "demo", "localhost:8080", svs),
  )

  svr.Run()
}

func newService() *http.ServeMux {
  s := http.NewServeMux()
  s.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte("OK"))
  })
  return s
}
```

## How to Contribute

Make a pull request...

## License

Distributed under MIT License, please see license file within the code for more details.
