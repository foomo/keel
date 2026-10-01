package telemetry

import (
	goruntime "github.com/foomo/go/runtime"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
)

// CodeCaller returns the code.function.name, code.file.path and
// code.line.number attributes of the caller skip frames above the caller of
// CodeCaller, or nil if the frame cannot be resolved.
func CodeCaller(skip int) []attribute.KeyValue {
	if fr := goruntime.CallFrame(skip + 1); !fr.Zero() {
		return []attribute.KeyValue{
			semconv.CodeFunctionName(fr.Name()),
			semconv.CodeFilePath(fr.File),
			semconv.CodeLineNumber(fr.Line),
		}
	}

	return nil
}

// CodeStacktrace returns a code.stacktrace attribute with up to num frames,
// skipping skip frames above the caller of CodeStacktrace.
func CodeStacktrace(num, skip int) attribute.KeyValue {
	return semconv.CodeStacktrace(goruntime.StackTrace(num, skip+1))
}
