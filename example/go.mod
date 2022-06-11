module github.com/foomo/keel/example

go 1.16

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/foomo/keel v0.0.0
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/google/uuid v1.3.0
	github.com/jackc/pgx/v4 v4.16.1
	github.com/nats-io/nats.go v1.16.0
	github.com/pkg/errors v0.9.1
	go.mongodb.org/mongo-driver v1.9.1
	go.opentelemetry.io/otel v1.7.0
	go.opentelemetry.io/otel/metric v0.30.0
	go.uber.org/zap v1.21.0
)

replace (
	github.com/foomo/keel => ../
	// TODO remove once https://github.com/spf13/viper/pull/1371 is merged
	github.com/spf13/viper v1.12.0 => github.com/franklinkim/viper v1.12.1-0.20220611111410-2d69ce7c2fe8
)
