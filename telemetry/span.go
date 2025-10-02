package telemetry

import (
	"context"
	"runtime"

	"github.com/grafana/pyroscope-go"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return Tracer().Start(ctx, spanName, opts...) //nolint:spancheck
}

func Span(ctx context.Context, handler func(ctx context.Context, span trace.Span) error) (err error) { //nolint:nonamedreturns
	name := "unknown"

	pc, _, _, ok := runtime.Caller(1)
	if ok {
		details := runtime.FuncForPC(pc)
		if details != nil {
			name = details.Name()
		}
	}

	pyroscope.TagWrapper(ctx, pyroscope.Labels("span_name", name), func(c context.Context) {
		ctx, span := Tracer().Start(ctx, name)

		defer func() {
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}

			span.End()
		}()

		err = handler(ctx, span)
	})

	return
}

func End(sp trace.Span, err error) {
	sp.SetStatus(codes.Ok, "")

	if err != nil {
		sp.RecordError(err)
		sp.SetStatus(codes.Error, err.Error())
	}

	sp.End()
}
