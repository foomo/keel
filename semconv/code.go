package semconv

import (
	"github.com/foomo/keel/internal/runtimeutil"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

func CodeCaller(skip int) []attribute.KeyValue {
	if shortName, _, file, line, ok := runtimeutil.Caller(skip + 1); ok {
		return []attribute.KeyValue{
			semconv.CodeFunctionName(shortName),
			semconv.CodeFilePath(file),
			semconv.CodeLineNumber(line),
		}
	}

	return nil
}

func CodeStacktrace(num, skip int) attribute.KeyValue {
	return semconv.CodeStacktrace(runtimeutil.StackTrace(num, skip+1))
}
