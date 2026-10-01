// Package keeltemporal provides helpers for running Temporal clients and
// workers inside a keel server.
//
// It covers client construction with optional namespace registration and
// OpenTelemetry instrumentation ([NewClient]), a keel service wrapper for
// Temporal workers ([NewService]), adapters bridging Temporal's logger and
// metrics interfaces to zap and OpenTelemetry ([NewLogger],
// [NewMetricsHandler]), activity option helpers ([WithActivityOptions]) and
// error classification helpers ([NewActivityError], [IsActivityError],
// [AsApplicationError]).
//
// # Usage
//
//	c, err := keeltemporal.NewClient(ctx, "localhost:7233",
//		keeltemporal.ClientWithNamespace("my-namespace"),
//	)
//	if err != nil {
//		return err
//	}
//	w := worker.New(c, "my-queue", worker.Options{})
//	svr.AddService(keeltemporal.NewService(nil, "temporal", w))
package keeltemporal
