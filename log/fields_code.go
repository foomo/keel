package log

import (
	"go.uber.org/zap"
)

const (
	CodeInstanceKey = "code_instance"
	CodeMethodKey   = "code_method"
	CodeLineKey     = "code_line"
)

func FCodeInstance(v string) zap.Field {
	return zap.String(CodeInstanceKey, v)
}

func FCodeMethod(v string) zap.Field {
	return zap.String(CodeMethodKey, v)
}

func FCodeLine(v int) zap.Field {
	return zap.Int(CodeLineKey, v)
}
