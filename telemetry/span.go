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

func Wrap(ctx context.Context, handler func(ctx context.Context) error) (err error) { //nolint:nonamedreturns
	name := "unkownn"
	pc, _, _, ok := runtime.Caller(1)
	if ok {
		details := runtime.FuncForPC(pc)
		if details != nil {
			name = details.Name()
		}
	}
	pyroscope.TagWrapper(ctx, pyroscope.Labels("span_name", name), func(c context.Context) {
		ctx, sp := Tracer().Start(ctx, name)
		defer func() {
			if err != nil {
				sp.RecordError(err)
				sp.SetStatus(codes.Error, err.Error())
			}
			sp.End()
		}()
		err = handler(ctx)
	})
	return
}

func End(sp trace.Span, err error) {
	if err != nil {
		sp.RecordError(err)
		sp.SetStatus(codes.Error, err.Error())
	}
	sp.End()
}
