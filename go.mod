module github.com/foomo/keel

go 1.16

require (
	github.com/foomo/gotsrpc/v2 v2.5.3
	github.com/go-logr/logr v1.2.3
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/google/uuid v1.3.0
	github.com/lib/pq v1.10.6
	github.com/mitchellh/mapstructure v1.5.0
	github.com/nats-io/nats-server/v2 v2.7.3 // indirect
	github.com/nats-io/nats.go v1.16.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.12.2
	github.com/spf13/viper v1.12.0
	github.com/stretchr/testify v1.8.0
	github.com/tinylib/msgp v1.1.6
	go.mongodb.org/mongo-driver v1.9.1
	go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo v0.32.0
	go.opentelemetry.io/contrib/instrumentation/host v0.32.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.32.0
	go.opentelemetry.io/contrib/instrumentation/runtime v0.32.0
	go.opentelemetry.io/otel v1.8.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.7.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.7.0
	go.opentelemetry.io/otel/exporters/prometheus v0.30.0
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v0.30.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.7.0
	go.opentelemetry.io/otel/metric v0.30.0
	go.opentelemetry.io/otel/sdk v1.7.0
	go.opentelemetry.io/otel/sdk/metric v0.30.0
	go.opentelemetry.io/otel/trace v1.8.0
	go.temporal.io/api v1.8.0
	go.temporal.io/sdk v1.15.0
	go.temporal.io/sdk/contrib/opentelemetry v0.1.0
	go.uber.org/zap v1.21.0
	golang.org/x/crypto v0.0.0-20220525230936-793ad666bf5e
	golang.org/x/sync v0.0.0-20220513210516-0976fa681c29
)

replace (
	// TODO remove once https://github.com/spf13/viper/pull/1371 is merged
	github.com/spf13/viper v1.12.0 => github.com/franklinkim/viper v1.12.1-0.20220611111410-2d69ce7c2fe8
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.32.0 => github.com/foomo/opentelemetry-go-contrib/instrumentation/net/http/otelhttp v0.32.1-0.20220517120905-10e2553b9bac
)
