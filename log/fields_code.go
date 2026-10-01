package log

import (
	"go.uber.org/zap"
)

const (
	// CodeInstanceKey is the log field key "code_instance".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	CodeInstanceKey = "code_instance"
	// CodePackageKey is the log field key "code_package".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	CodePackageKey = "code_package"
	// CodeMethodKey is the log field key "code_method".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	CodeMethodKey = "code_method"
	// CodeLineKey is the log field key "code_line".
	//
	// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
	CodeLineKey = "code_line"
)

// FCodeInstance returns a field with the given value under [CodeInstanceKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FCodeInstance(v string) zap.Field {
	return zap.String(CodeInstanceKey, v)
}

// FCodePackage returns a field with the given value under [CodePackageKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FCodePackage(v string) zap.Field {
	return zap.String(CodePackageKey, v)
}

// FCodeMethod returns a field with the given value under [CodeMethodKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FCodeMethod(v string) zap.Field {
	return zap.String(CodeMethodKey, v)
}

// FCodeLine returns a field with the given value under [CodeLineKey].
//
// Deprecated: Use OpenTelemetry semconv attributes with [Attribute] instead.
func FCodeLine(v int) zap.Field {
	return zap.Int(CodeLineKey, v)
}
