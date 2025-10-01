package telemetry

import (
	"context"
	"os"
	"runtime"

	"github.com/foomo/keel/env"
	otelpyroscope "github.com/grafana/otel-profiling-go"
	"github.com/grafana/pyroscope-go"
	"go.opentelemetry.io/otel"
)

func NewProfiler(ctx context.Context) (*pyroscope.Profiler, error) {
	tags := map[string]string{}
	if v := os.Getenv("HOSTNAME"); v != "" {
		tags["pod"] = v
	}
	if v := os.Getenv("OTEL_SERVICE_GIT_REF"); v != "" {
		tags["service_git_ref"] = v
	}
	if v := os.Getenv("OTEL_SERVICE_REPOSITORY"); v != "" {
		tags["service_repository"] = v
	}
	if v := os.Getenv("OTEL_SERVICE_ROOT_PATH"); v != "" {
		tags["service_root_path"] = v
	}

	profileTypes := []pyroscope.ProfileType{
		// Default
		pyroscope.ProfileCPU,
		pyroscope.ProfileAllocObjects,
		pyroscope.ProfileAllocSpace,
		pyroscope.ProfileInuseObjects,
		pyroscope.ProfileInuseSpace,
	}
	if env.GetBool("OTEL_PROFILE_BLOCK_ENABLED", false) {
		runtime.SetBlockProfileRate(env.GetInt("OTEL_PROFILE_BLOCK_RATE", 5))
		profileTypes = append(profileTypes,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		)
	}
	if env.GetBool("OTEL_PROFILE_MUTEX_ENABLED", false) {
		runtime.SetMutexProfileFraction(env.GetInt("OTEL_PROFILE_MUTEX_FRACTION", 5))
		profileTypes = append(profileTypes,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
		)
	}
	if env.GetBool("OTEL_PROFILE_GOROUTINES_ENABLED", false) {
		profileTypes = append(profileTypes,
			pyroscope.ProfileGoroutines,
		)
	}

	resource, err := NewResource(ctx)
	if err != nil {
		return nil, err
	}

	for _, value := range resource.Attributes() {
		if value.Key == "service.name" {
			continue
		}
		tags[string(value.Key)] = value.Value.Emit()
	}

	p, err := pyroscope.Start(pyroscope.Config{
		ApplicationName: env.Get("OTEL_SERVICE_NAME", DefaultServiceName),
		// Logger:          internalpyroscope.NewLogger(),
		ProfileTypes: profileTypes,
		Tags:         tags,
	})
	if err != nil {
		return nil, err
	}

	otel.SetTracerProvider(otelpyroscope.NewTracerProvider(otel.GetTracerProvider()))
	return p, nil
}
