# keel
[![GitHub Super-Linter](https://github.com/foomo/keel/workflows/CI/badge.svg)](https://github.com/marketplace/actions/super-linter)


> Opinionated way to run services.

## Stack

- Configuration: Viper
- Metrics: Prometheus
- Telemetry: Open Telemetry
- Logging: Zap

## Example

```go
package main

import (
	"github.com/foomo/keel"
	"github.com/foomo/keel/config"
	"github.com/foomo/keel/log"
	"github.com/foomo/keel/net/http/middleware"
	"github.com/foomo/keel/telemetry"
)

func main() {
	svr := keel.NewServer()

	// register Closers for graceful shutdowns
	svr.AddClosers(telemetry.Provider())

	// add zap service
	svr.AddServices(keel.NewDefaultServiceHTTPZap())

	// add viper service
	svr.AddServices(keel.NewDefaultServiceHTTPViper())

	// add prometheus service
	svr.AddServices(keel.NewDefaultServiceHTTPPrometheus())

	svr.Run()
}
```
