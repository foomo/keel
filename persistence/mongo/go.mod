module github.com/foomo/keel/persistence/mongo

go 1.27.0

replace github.com/foomo/keel => ../../

require (
	github.com/foomo/keel v0.31.0
	github.com/foomo/mongo-lock v0.2.0
	github.com/go-logr/zapr v1.3.0
	github.com/pkg/errors v0.9.1
	go.mongodb.org/mongo-driver/v2 v2.9.1
	go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/v2/mongo/otelmongo v0.0.0-20260930123116-8d041e719a38
	go.uber.org/zap v1.28.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/fbiville/markdown-table-formatter v0.3.0 // indirect
	github.com/foomo/opentelemetry-go v0.4.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/klauspost/compress v1.20.1 // indirect
	github.com/testcontainers/testcontainers-go v0.44.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.2.0 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.47.0-rc.1.0.20260924072922-c131dcfc3885 // indirect
	go.opentelemetry.io/otel/log v1.47.0-rc.1 // indirect
	go.opentelemetry.io/otel/metric v1.47.0-rc.1.0.20260924072922-c131dcfc3885 // indirect
	go.opentelemetry.io/otel/trace v1.47.0-rc.1.0.20260924072922-c131dcfc3885 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.57.0 // indirect
	golang.org/x/sync v0.23.0 // indirect
	golang.org/x/text v0.42.0 // indirect
)
