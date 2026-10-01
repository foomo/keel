// Package healthz provides the probe types and interfaces used by keel to
// implement Kubernetes startup, readiness and liveness checks.
//
// A probe is any value implementing one of [BoolHealthzer],
// [BoolHealthzerWithContext], [ErrorHealthzer] or [ErrorHealthzWithContext]
// (keel also accepts the pinger interfaces from
// [github.com/foomo/keel/interfaces]). Probes are registered on a
// [github.com/foomo/keel.Server] per [Type] and are served by the health
// service from [github.com/foomo/keel/service.NewHealthz].
package healthz
