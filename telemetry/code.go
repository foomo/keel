package telemetry

import (
	runtimex "github.com/foomo/go/runtime"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
)

func CodeCaller(skip int) []attribute.KeyValue {
	if shortName, _, file, line, ok := runtimex.Caller(skip + 1); ok {
		return []attribute.KeyValue{
			semconv.CodeFunctionName(shortName),
			semconv.CodeFilePath(file),
			semconv.CodeLineNumber(line),
		}
	}

	return nil
}

func CodeStacktrace(num, skip int) attribute.KeyValue {
	return semconv.CodeStacktrace(runtimex.StackTrace(num, skip+1))
}
