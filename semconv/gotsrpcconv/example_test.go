package gotsrpcconv_test

import (
	"context"
	"fmt"

	"github.com/foomo/keel/semconv/gotsrpcconv"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/noop"
)

func ExampleNewExecutionDuration() {
	// Create an ExecutionDuration instrument from a meter.
	meter := noop.NewMeterProvider().Meter("example")

	duration, err := gotsrpcconv.NewExecutionDuration(meter)
	if err != nil {
		panic(err)
	}

	// Record a 150ms execution.
	duration.Record(context.Background(), 0.150, "mypackage", "MyService", "MyFunc")

	fmt.Println(duration.Name())
	fmt.Println(duration.Unit())
	fmt.Println(duration.Description())

	// Output:
	// gotsrpc.execution.duration
	// s
	// Duration of GOTSRPC execution.
}

func ExampleExecutionDuration_AttrError() {
	duration, _ := gotsrpcconv.NewExecutionDuration(nil)

	attr := duration.AttrError(true)
	fmt.Println(attr.Key)
	fmt.Println(attr.Value.AsBool())

	// Record with error attribute.
	duration.Record(
		context.Background(),
		0.250,
		"mypackage", "MyService", "MyFunc",
		attr,
	)

	// Output:
	// gotsrpc.error
	// true
}

func ExampleNewExecutionDuration_nilMeter() {
	// Passing a nil meter returns a safe no-op instrument.
	duration, err := gotsrpcconv.NewExecutionDuration(nil)
	if err != nil {
		panic(err)
	}

	// Recording is a no-op but does not panic.
	duration.Record(context.Background(), 0.100, "pkg", "Svc", "Fn")
	duration.RecordSet(context.Background(), 0.200, attribute.NewSet())

	fmt.Println("ok")

	// Output:
	// ok
}
