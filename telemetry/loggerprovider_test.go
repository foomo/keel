package telemetry_test

import (
	"context"

	"github.com/foomo/keel/telemetry"
	"go.opentelemetry.io/otel/log"
	"go.uber.org/zap"
)

func ExampleNewZapLoggerProvider() {
	l := zap.NewExample()
	ctx := context.Background()

	_, _ = telemetry.NewZapLoggerProvider(ctx, l)

	// Raw log record
	record := log.Record{}
	record.SetSeverity(log.SeverityInfo)
	record.SetBody(log.StringValue("something really cool"))

	tl := telemetry.Logger()
	tl.Emit(ctx, record)

	// Output:
	// {"level":"info","msg":"something really cool"}
}
