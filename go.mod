module github.com/foomo/keel

go 1.16

require (
	github.com/foomo/gotsrpc/v2 v2.5.2
	github.com/go-logr/logr v1.2.3
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/google/uuid v1.3.0
	github.com/jackc/pgx/v4 v4.16.1
	github.com/mitchellh/mapstructure v1.5.0
	github.com/nats-io/nats-server/v2 v2.7.3 // indirect
	github.com/nats-io/nats.go v1.15.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.12.1
	github.com/spf13/viper v1.11.0
	github.com/stretchr/testify v1.7.1
	github.com/tinylib/msgp v1.1.6
	go.etcd.io/etcd/client/v3 v3.5.4
	go.mongodb.org/mongo-driver v1.9.1
	go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo v0.32.0
	go.opentelemetry.io/contrib/instrumentation/host v0.32.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.32.0
	go.opentelemetry.io/contrib/instrumentation/runtime v0.32.0
	go.opentelemetry.io/otel v1.7.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.7.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.7.0
	go.opentelemetry.io/otel/exporters/prometheus v0.30.0
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v0.30.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.7.0
	go.opentelemetry.io/otel/metric v0.30.0
	go.opentelemetry.io/otel/sdk v1.7.0
	go.opentelemetry.io/otel/sdk/metric v0.30.0
	go.opentelemetry.io/otel/trace v1.7.0
	go.temporal.io/api v1.7.1-0.20220223032354-6e6fe738916a
	go.temporal.io/sdk v1.14.0
	go.temporal.io/sdk/contrib/opentelemetry v0.1.0
	go.uber.org/zap v1.21.0
	golang.org/x/crypto v0.0.0-20220427172511-eb4f295cb31f
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
)
