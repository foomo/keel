package log

import (
	"go.uber.org/zap"
)

const (
	// Deprecated: use semconv messaging attributes instead.
	CodeInstanceKey = "code_instance"
	// Deprecated: use semconv messaging attributes instead.
	CodePackageKey = "code_package"
	// Deprecated: use semconv messaging attributes instead.
	CodeMethodKey = "code_method"
	// Deprecated: use semconv messaging attributes instead.
	CodeLineKey = "code_line"
)

// Deprecated: use semconv messaging attributes instead.
func FCodeInstance(v string) zap.Field {
	return zap.String(CodeInstanceKey, v)
}

// Deprecated: use semconv messaging attributes instead.
func FCodePackage(v string) zap.Field {
	return zap.String(CodePackageKey, v)
}

// Deprecated: use semconv messaging attributes instead.
func FCodeMethod(v string) zap.Field {
	return zap.String(CodeMethodKey, v)
}

// Deprecated: use semconv messaging attributes instead.
func FCodeLine(v int) zap.Field {
	return zap.Int(CodeLineKey, v)
}
