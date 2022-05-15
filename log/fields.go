package log

import (
	"fmt"

	"go.uber.org/zap"
)

const (
	NumKey = "num"

	NameKey = "name"

	// ValueKey represents a generic value attribute.
	ValueKey = "value"
)

func FNum(num int) zap.Field {
	return zap.Int(NumKey, num)
}

func FName(name string) zap.Field {
	return zap.String(NameKey, name)
}

func FValue(value interface{}) zap.Field {
	return zap.String(ValueKey, fmt.Sprintf("%v", value))
}
