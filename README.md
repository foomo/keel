[![Build Status](https://github.com/foomo/keel/actions/workflows/test.yml/badge.svg?branch=main&event=push)](https://github.com/foomo/keel/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/foomo/keel)](https://goreportcard.com/report/github.com/foomo/keel)
[![Coverage Status](https://coveralls.io/repos/github/foomo/keel/badge.svg?branch=main&)](https://coveralls.io/github/foomo/keel?branch=main)
[![GoDoc](https://godoc.org/github.com/foomo/keel?status.svg)](https://godoc.org/github.com/foomo/keel)

<p align="center">
  <img alt="sesamy" src=".github/assets/keel.png"/>
</p>

# keel

> Opinionated way to run services on Kubernetes

## Stack

- Zap
- Nats
- Viper
- GoTSRPC
- Temporal
- OpenTelemetry

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

_Made with ♥ [foomo](https://www.foomo.org) by [bestbytes](https://www.bestbytes.com)_
