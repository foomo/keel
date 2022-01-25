module github.com/foomo/keel/example

go 1.16

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/foomo/keel v0.0.0
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/google/uuid v1.3.0
	github.com/jackc/pgx/v4 v4.14.1
	github.com/nats-io/nats.go v1.13.0
	github.com/pkg/errors v0.9.1
	go.mongodb.org/mongo-driver v1.8.2
	go.opentelemetry.io/otel v0.20.0
	go.opentelemetry.io/otel/metric v0.20.0
	go.uber.org/zap v1.20.0
)

replace github.com/foomo/keel => ../
