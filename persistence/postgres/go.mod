module github.com/foomo/keel/persistence/postgres

go 1.25.0

replace github.com/foomo/keel => ../../

require (
	github.com/foomo/keel v0.0.0-00010101000000-000000000000
	github.com/lib/pq v1.10.9
	github.com/pkg/errors v0.9.1
	go.uber.org/zap v1.27.0
)

require (
	github.com/fbiville/markdown-table-formatter v0.3.0 // indirect
	go.opentelemetry.io/otel v1.38.0 // indirect
	go.opentelemetry.io/otel/trace v1.38.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
)
