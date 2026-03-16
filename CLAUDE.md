# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Keel is an opinionated Go framework for building and running services on Kubernetes. It orchestrates multiple services with graceful shutdown, structured logging (Zap), configuration (Viper), OpenTelemetry observability, and Kubernetes health probes.

## Build & Development Commands

**Prerequisites:** [mise](https://mise.jdx.dev) must be installed (manages lefthook and golangci-lint).

```bash
make check          # Run tidy, generate, lint, and test (full CI check)
make tidy           # Run go mod tidy across all modules
make lint           # Run golangci-lint across all modules
make lint.fix       # Run golangci-lint with --fix
make test           # Run tests with coverage (uses -tags=safe)
make test.race      # Run tests with race detector
make test.update    # Run tests with -update flag (update golden files)
make generate       # Run go generate
```

**Running a single test:**
```bash
go test -tags=safe -run TestName ./path/to/package/...
```

**Running tests in a submodule:**
```bash
cd persistence/mongo && go test -tags=safe ./...
```

## Multi-Module Repository

This is a Go workspace (`go.work`) with multiple modules:

- `.` — Root module (`github.com/foomo/keel`)
- `./examples` — Example applications
- `./integration/gotsrpc` — GoTSRPC integration
- `./integration/temporal` — Temporal workflow integration
- `./net/stream` — NATS streaming
- `./persistence/mongo` — MongoDB persistence
- `./persistence/postgres` — PostgreSQL persistence

When modifying a submodule, run `go mod tidy` in that module's directory. The `make tidy` command does this for all modules.

## Architecture

### Core Pattern: Server → Services

The `Server` (`server.go`) orchestrates `Service` instances. Services implement:

```go
type Service interface {
    Name() string
    Start(ctx context.Context) error
    Close(ctx context.Context) error
}
```

**Service types** (in `service/` package):
- `service.HTTP` — HTTP listeners with configurable middleware
- `service.GoRoutine` — Background goroutine tasks
- Init services — Started immediately during `NewServer()`, before `Run()`

**Server configuration** uses the functional options pattern (`option.go`):
```go
svr := keel.NewServer(
    keel.WithHTTPZapService(...),
    keel.WithHTTPPrometheusService(...),
)
svr.AddServices(myService)
svr.Run()
```

### Graceful Shutdown

The server manages shutdown via multiple closer interface types defined in `interfaces/`. Any type implementing `Close()`, `Shutdown()`, `Stop()`, or `Unsubscribe()` (with optional `error` return and `context.Context` parameter) can be registered as a closer.

### Health Probes

Kubernetes-aligned probes in `healthz/`: `TypeStartup`, `TypeReadiness`, `TypeLiveness`, `TypeAlways`. Register probes via `server.AddStartupHealthzers()`, `server.AddReadinessHealthzers()`, etc.

### Key Packages

- `config/` — Viper-based configuration management
- `env/` — Environment variable helpers
- `log/` — Zap logger setup and structured field helpers
- `telemetry/` — OpenTelemetry trace/metric/log providers
- `metrics/` — Prometheus metrics and OTEL instrumentation
- `net/http/` — HTTP middleware (CORS, gzip, JWT auth)
- `net/http/roundtripware/` — HTTP client round-trip middleware
- `keeltest/` — Test server helpers and assertion utilities

## Git Conventions

- **Branch naming:** Must start with `feature/` or `fix/` (enforced by lefthook pre-commit hook)
- **Commit messages:** [Conventional Commits](https://www.conventionalcommits.org/) format enforced by lefthook commit-msg hook
  - Format: `type(scope?): subject` (max 50 chars)
  - Types: `build`, `chore`, `ci`, `docs`, `feat`, `fix`, `perf`, `refactor`, `style`, `test`, `sec`, `wip`, `revert`

## Linting

- Uses golangci-lint v2 with `default: all` linters (most enabled, specific ones disabled)
- Build tag: `safe` (used in tests and lint runs)
- Formatters: `gofmt` and `goimports`
- `examples/` and `tmp/` directories are excluded from linting

## Releases

Multi-module release tagging: `make release TAG=1.0.0` creates `v1.0.0` for root and `<submodule>/v1.0.0` for each submodule.
