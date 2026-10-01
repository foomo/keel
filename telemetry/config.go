package telemetry

import (
	"runtime/debug"
	"sync"
)

// Name is the instrumentation scope name used by [Tracer] and [Meter].
// DefaultServiceName is the service name used by [NewProfiler] when
// OTEL_SERVICE_NAME is unset.
var (
	Name               = "github.com/foomo/keel/telemetry"
	DefaultServiceName = "undefined"
)

// Version returns the keel module version for use as the OTel
// instrumentation scope version. It is resolved once from the build info and
// is empty if unavailable.
var Version = sync.OnceValue(func() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}

	if bi.Main.Path == Name {
		return bi.Main.Version
	}

	for _, d := range bi.Deps {
		if d.Path == Name {
			return d.Version
		}
	}

	return ""
})
