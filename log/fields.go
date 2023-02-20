package log

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

const (
	// NumKey - generic number attribute
	NumKey = "num"
	// NameKey - generic name attribute
	NameKey = "name"
	// ValueKey - generic value attribute
	ValueKey = "value"
	// JSONKey - generic json attribute
	JSONKey = "json"
)

// FNum - returns zap field
func FNum(num int) zap.Field {
	return zap.Int(NumKey, num)
}

// FName - returns zap field
func FName(name string) zap.Field {
	return zap.String(NameKey, name)
}

// FValue - returns zap field
func FValue(value interface{}) zap.Field {
	return zap.String(ValueKey, fmt.Sprintf("%v", value))
}

// FJSON - returns zap field
func FJSON(v interface{}) zap.Field {
	if out, err := json.Marshal(v); err != nil {
		return zap.String(JSONKey+"_error", err.Error())
	} else {
		raw := json.RawMessage(out)
		return zap.Any(JSONKey, &raw)
	}
}
