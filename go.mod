module github.com/foomo/keel

go 1.16

require (
	github.com/go-logr/logr v1.2.3
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/google/uuid v1.3.0
	github.com/jackc/pgx/v4 v4.15.0
	github.com/mitchellh/mapstructure v1.4.3
	github.com/nats-io/nats-server/v2 v2.7.3 // indirect
	github.com/nats-io/nats.go v1.13.1-0.20220121202836-972a071d373d
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.12.1
	github.com/spf13/viper v1.10.1
	github.com/stretchr/testify v1.7.1
	github.com/tinylib/msgp v1.1.6
	go.mongodb.org/mongo-driver v1.8.4
	go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo v0.30.0
	go.opentelemetry.io/contrib/instrumentation/host v0.29.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.30.0
	go.opentelemetry.io/contrib/instrumentation/runtime v0.30.0
	go.opentelemetry.io/otel v1.6.3
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.6.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.6.3
	go.opentelemetry.io/otel/exporters/prometheus v0.27.0
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v0.27.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.6.0
	go.opentelemetry.io/otel/metric v0.27.0
	go.opentelemetry.io/otel/sdk v1.6.3
	go.opentelemetry.io/otel/sdk/metric v0.27.0
	go.opentelemetry.io/otel/trace v1.6.3
	go.temporal.io/api v1.7.0
	go.temporal.io/sdk v1.13.1
	go.temporal.io/sdk/contrib/opentelemetry v0.1.0
	go.uber.org/zap v1.21.0
	golang.org/x/crypto v0.0.0-20220307211146-efcb8507fb70
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
)
