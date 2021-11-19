module github.com/foomo/keel

go 1.16

require (
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/google/uuid v1.1.2
	github.com/jackc/pgx/v4 v4.13.0
	github.com/mitchellh/mapstructure v1.4.2
	github.com/nats-io/nats.go v1.12.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.10.0
	github.com/spf13/viper v1.9.0
	github.com/stretchr/testify v1.7.0
	go.mongodb.org/mongo-driver v1.5.1
	go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo v0.20.0
	go.opentelemetry.io/contrib/instrumentation/host v0.20.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.20.0
	go.opentelemetry.io/contrib/instrumentation/runtime v0.20.0
	go.opentelemetry.io/otel v0.20.0
	go.opentelemetry.io/otel/exporters/metric/prometheus v0.20.0
	go.opentelemetry.io/otel/exporters/otlp v0.20.0
	go.opentelemetry.io/otel/exporters/stdout v0.20.0
	go.opentelemetry.io/otel/metric v0.20.0
	go.opentelemetry.io/otel/sdk v0.20.0
	go.opentelemetry.io/otel/sdk/metric v0.20.0
	go.opentelemetry.io/otel/trace v0.20.0
	go.uber.org/zap v1.17.0
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
)
