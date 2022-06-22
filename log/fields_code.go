package log

import (
	"go.uber.org/zap"
)

const (
	CodeInstanceKey = "code_instance"
	CodePackageKey  = "code_package"
	CodeMethodKey   = "code_method"
	CodeLineKey     = "code_line"
)

func FCodeInstance(v string) zap.Field {
	return zap.String(CodeInstanceKey, v)
}

func FCodePackage(v string) zap.Field {
	return zap.String(CodePackageKey, v)
}

func FCodeMethod(v string) zap.Field {
	return zap.String(CodeMethodKey, v)
}

func FCodeLine(v int) zap.Field {
	return zap.Int(CodeLineKey, v)
}
