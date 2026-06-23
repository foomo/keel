package main

import (
	"context"
	"time"

	"github.com/foomo/keel"
	"github.com/foomo/keel/interfaces"
	"go.uber.org/zap"
)

func main() {
	l := zap.NewExample().Named("root")

	job := keel.NewJob(
		keel.JobWithName("example"),
		keel.JobWithLogger(l.Named("job")),
		// Wire telemetry from OTEL env vars (e.g. OTEL_METRICS_EXPORTER="otlp" to
		// push metrics on exit instead of exposing them — jobs exit before a scrape):
		// keel.JobWithTelemetry(),
		// keel.JobWithPushgatewayMeter("http://localhost:9091"),
		// Run steps concurrently (fail-fast, 0 = unbounded):
		// keel.JobWithParallel(0),
		keel.JobWithTimeout(30*time.Second),
	)

	// Register resources to be flushed/closed during finalization.
	job.AddCloser(interfaces.CloserFunc(func(ctx context.Context) error {
		l.Named("closer").Info("closing resources")
		return nil
	}))

	// Steps run in registration order, each in its own span.
	job.AddStep("extract", func(ctx context.Context, l *zap.Logger) error {
		l.Info("extracting data")
		return nil
	})

	job.AddStep("transform", func(ctx context.Context, l *zap.Logger) error {
		l.Info("transforming data")
		return nil
	})

	job.AddStep("load", func(ctx context.Context, l *zap.Logger) error {
		l.Info("loading data")
		return nil
	})

	// Run exits the process: 0 on success, 1 on failure, so Kubernetes can apply
	// its backoffLimit. Use RunE instead if you need to handle the error yourself.
	job.Run()
}
