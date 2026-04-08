package messaging

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	semconvmsg "go.opentelemetry.io/otel/semconv/v1.40.0/messagingconv"
	"go.opentelemetry.io/otel/trace"
)

const instrName = "github.com/bestbytes/messaging"

// ---------------------------------------------------------------------------
// Package-level singleton
// ---------------------------------------------------------------------------

var (
	globalTel  *telemetry
	globalOnce sync.Once
	globalErr  error
)

// tel returns the package-level telemetry, initialising it once against the
// current OTel global providers. Transports call this directly — no
// constructor argument required.
func tel() *telemetry {
	globalOnce.Do(func() {
		globalTel, globalErr = newTelemetry(otel.GetTracerProvider(), otel.GetMeterProvider())
	})
	return globalTel // nil if init failed; callers guard with t != nil
}

// InitErr returns any error that occurred during package-level telemetry
// initialisation. Call it once after setting up your OTel providers if you
// want to surface init failures explicitly.
func InitErr() error { return globalErr }

// ResetForTest tears down the singleton so tests can inject a fresh provider.
// Must only be called from test code.
func ResetForTest() {
	globalOnce = sync.Once{}
	globalTel = nil
	globalErr = nil
}

// ---------------------------------------------------------------------------
// telemetry — instruments
// ---------------------------------------------------------------------------

type telemetry struct {
	tracer trace.Tracer

	// semconv-defined metrics
	sentMessages     semconvmsg.ClientSentMessages      // messaging.client.sent.messages
	consumedMessages semconvmsg.ClientConsumedMessages  // messaging.client.consumed.messages
	publishDuration  semconvmsg.ClientOperationDuration // messaging.client.operation.duration
	processDuration  semconvmsg.ProcessDuration         // messaging.process.duration

	// custom metric — no semconv equivalent; follows messaging.* naming
	consumerLagMu sync.Mutex
	consumerLag   metric.Int64ObservableGauge // messaging.consumer.lag
}

func newTelemetry(tp trace.TracerProvider, mp metric.MeterProvider) (*telemetry, error) {
	m := mp.Meter(instrName)
	t := &telemetry{tracer: tp.Tracer(instrName)}
	var err error

	// messaging.client.sent.messages — semconv ClientSentMessages
	t.sentMessages, err = semconvmsg.NewClientSentMessages(m)
	if err != nil {
		return nil, fmt.Errorf("messaging telemetry: sent messages: %w", err)
	}

	// messaging.client.consumed.messages — semconv ClientConsumedMessages
	t.consumedMessages, err = semconvmsg.NewClientConsumedMessages(m)
	if err != nil {
		return nil, fmt.Errorf("messaging telemetry: consumed messages: %w", err)
	}

	// messaging.client.operation.duration — semconv ClientOperationDuration
	t.publishDuration, err = semconvmsg.NewClientOperationDuration(m)
	if err != nil {
		return nil, fmt.Errorf("messaging telemetry: publish duration: %w", err)
	}

	// messaging.process.duration — semconv ProcessDuration
	t.processDuration, err = semconvmsg.NewProcessDuration(m)
	if err != nil {
		return nil, fmt.Errorf("messaging telemetry: process duration: %w", err)
	}

	return t, nil
}

// RegisterLag registers the messaging.consumer.lag observable gauge for a
// given subject against mp. Called once per Subscriber from its constructor.
// Follows the messaging.* naming pattern; no semconv equivalent exists.
func RegisterLag(mp metric.MeterProvider, subject string, lagFn func() int64) (metric.Int64ObservableGauge, error) {
	return mp.Meter(instrName).Int64ObservableGauge(
		"messaging.consumer.lag",
		metric.WithDescription("Number of messages waiting in the subscriber buffer"),
		metric.WithInt64Callback(func(_ context.Context, obs metric.Int64Observer) error {
			obs.Observe(lagFn(),
				metric.WithAttributes(attribute.String("messaging.destination.name", subject)),
			)
			return nil
		}),
	)
}

// ---------------------------------------------------------------------------
// Internal helpers called by transports
// ---------------------------------------------------------------------------

// RecordPublish opens a producer span, calls fn, records duration and counter.
func RecordPublish(ctx context.Context, subject string, system semconvmsg.SystemAttr, fn func(context.Context) error) error {
	t := tel()

	var span trace.Span
	if t != nil {
		ctx, span = t.tracer.Start(ctx, "messaging.publish",
			trace.WithSpanKind(trace.SpanKindProducer),
			trace.WithAttributes(
				semconvmsg.ClientSentMessages{}.AttrDestinationName(subject),
			),
		)
		defer span.End()
	}

	start := time.Now()
	err := fn(ctx)
	s := msFloat(start)

	if t == nil {
		return err
	}

	errType := errorType(err)
	t.sentMessages.Add(ctx, 1,
		"publish",
		system,
		t.sentMessages.AttrDestinationName(subject),
		t.sentMessages.AttrErrorType(errType),
	)
	t.publishDuration.Record(ctx, s,
		"publish",
		system,
		t.publishDuration.AttrDestinationName(subject),
		t.publishDuration.AttrErrorType(errType),
	)
	recordSpanResult(span, err)
	return err
}

// RecordProcess opens a consumer span, calls fn, records duration and counter.
func RecordProcess(ctx context.Context, subject string, system semconvmsg.SystemAttr, fn func(context.Context) error) error {
	t := tel()

	var span trace.Span
	if t != nil {
		ctx, span = t.tracer.Start(ctx, "messaging.process",
			trace.WithSpanKind(trace.SpanKindConsumer),
			trace.WithAttributes(
				semconvmsg.ClientConsumedMessages{}.AttrDestinationName(subject),
			),
		)
		defer span.End()
	}

	start := time.Now()
	err := fn(ctx)
	s := msFloat(start)

	if t == nil {
		return err
	}

	errType := errorType(err)
	t.consumedMessages.Add(ctx, 1,
		"receive",
		system,
		t.consumedMessages.AttrDestinationName(subject),
		t.consumedMessages.AttrErrorType(errType),
	)
	t.processDuration.Record(ctx, s,
		"process",
		system,
		t.processDuration.AttrDestinationName(subject),
		t.processDuration.AttrErrorType(errType),
	)
	recordSpanResult(span, err)
	return err
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func msFloat(start time.Time) float64 {
	return float64(time.Since(start).Microseconds()) / 1000.0
}

func errorType(err error) semconvmsg.ErrorTypeAttr {
	if err == nil {
		return semconvmsg.ErrorTypeAttr("")
	}
	return semconvmsg.ErrorTypeAttr(fmt.Sprintf("%T", err))
}

func recordSpanResult(span trace.Span, err error) {
	if span == nil {
		return
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
}
