package log

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

const (
	NumKey = "num"

	NameKey = "name"

	DurationKey = "duration"

	// ValueKey represents a generic value attribute.
	ValueKey = "value"
)

func FNum(num int) zap.Field {
	return zap.Int(NumKey, num)
}

func FName(name string) zap.Field {
	return zap.String(NameKey, name)
}

func FDuration(duration time.Duration) zap.Field {
	return zap.Int64(DurationKey, duration.Milliseconds())
}

func FValue(value interface{}) zap.Field {
	return zap.String(ValueKey, fmt.Sprintf("%v", value))
}
