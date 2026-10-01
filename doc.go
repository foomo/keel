// Package keel provides an opinionated runtime for services and jobs running
// on Kubernetes, with graceful shutdown, structured logging, configuration,
// OpenTelemetry and health probes.
//
// # Server
//
// A [Server] orchestrates a set of [Service] values. It is created with
// [NewServer] and configured through [Option] values such as
// [WithHTTPZapService], [WithHTTPHealthzService] or [WithTelemetry]. Logging
// defaults to [log.Logger] and configuration to [config.Config].
//
//	svr := keel.NewServer(
//		keel.WithHTTPZapService(true),
//		keel.WithHTTPViperService(true),
//		keel.WithHTTPPrometheusService(true),
//		keel.WithHTTPHealthzService(true),
//	)
//
//	svr.AddService(service.NewHTTP(svr.Logger(), "demo", "localhost:8080", handler))
//
//	svr.Run()
//
// # Services
//
// Init services (registered by the WithHTTPXxxService options or
// [WithInitService]) are started immediately inside [NewServer]. Services
// added with [Server.AddService], [Server.AddHTTPService],
// [Server.AddGoRoutine] and friends are started by [Server.Run]. Ready made
// implementations live in [github.com/foomo/keel/service]. [ServiceFunc]
// adapts a plain function and [ServiceEnabler] starts and stops a service
// dynamically.
//
// # Graceful shutdown
//
// [Server.Run] blocks until a shutdown signal (SIGINT and SIGTERM by default,
// see [WithShutdownSignals]) is received, [Server.ShutdownCancel] is called,
// or a service fails. The server then reports not ready, and calls every
// registered closer followed by the trace, meter and logger providers, all
// bounded by the graceful period (KEEL_GRACEFUL_PERIOD seconds, default 30,
// see [WithGracefulPeriod]). A closer is any value implementing one of the
// Close, Shutdown, Stop or Unsubscribe shapes declared in
// [github.com/foomo/keel/interfaces]; see [IsCloser]. Services are
// registered as closers automatically.
//
// # Health probes
//
// Probes are registered per [healthz.Type] with [Server.AddStartupHealthzers],
// [Server.AddReadinessHealthzers], [Server.AddLivenessHealthzers] and
// [Server.AddAlwaysHealthzers]; see [IsHealthz] for the accepted shapes. They
// are served by the service enabled through [WithHTTPHealthzService].
//
// # Jobs
//
// A [Job], created with [NewJob] and configured through [JobOption] values,
// runs a list of steps to completion instead of blocking until a signal, as
// suited for a Kubernetes Job. Both [Server] and [Job] implement [Runtime].
//
//	job := keel.NewJob(keel.JobWithTelemetry())
//	job.AddStep("migrate", func(ctx context.Context, l *zap.Logger) error {
//		return migrate(ctx)
//	})
//	job.Run()
package keel
