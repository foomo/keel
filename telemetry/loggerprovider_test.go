package telemetry_test

import (
	"context"
	"testing"

	"github.com/foomo/keel/telemetry"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/logtest"
	"go.uber.org/zap"
)

func TestNewZapBridgeCore(t *testing.T) {
	rec := logtest.NewRecorder()
	l := zap.New(telemetry.NewZapBridgeCore(rec))

	l.Info("hello", zap.String("k", "v"))

	var records []logtest.Record
	for _, rs := range rec.Result() {
		records = append(records, rs...)
	}

	if len(records) != 1 {
		t.Fatalf("got %d records, want 1", len(records))
	}

	if got := records[0].Body.AsString(); got != "hello" {
		t.Errorf("body = %q, want %q", got, "hello")
	}

	var found bool

	for _, kv := range records[0].Attributes {
		if kv.Key == "k" && kv.Value.AsString() == "v" {
			found = true
		}
	}

	if !found {
		t.Errorf("missing attribute k=v, got %v", records[0].Attributes)
	}
}

func ExampleNewZapLoggerProvider() {
	l := zap.NewExample()
	ctx := context.Background()

	_, _ = telemetry.NewZapLoggerProvider(ctx, l)

	// Raw log record
	record := log.Record{}
	record.SetSeverity(log.SeverityInfo)
	record.SetBody(log.StringValue("something really cool"))

	tl := telemetry.LoggerProvider().Logger(telemetry.Name)
	tl.Emit(ctx, record)

	// Output:
	// {"level":"info","msg":"something really cool"}
}
