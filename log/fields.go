package log

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

// Generic log field keys.
const (
	NumKey   = "num"
	NameKey  = "name"
	ValueKey = "value"
	JSONKey  = "json"
)

// FNum creates a zap.Field with a given number under the key "num".
func FNum(num int) zap.Field {
	return zap.Int(NumKey, num)
}

// FName creates a zap.Field with a given string under the key "name".
func FName(name string) zap.Field {
	return zap.String(NameKey, name)
}

// FValue creates a zap.Field with value formatted by %v under the key "value".
func FValue(value any) zap.Field {
	return zap.String(ValueKey, fmt.Sprintf("%v", value))
}

// FJSON creates a zap.Field with v marshaled as raw JSON under the key "json".
// If marshaling fails, the error message is logged under "json_error" instead.
func FJSON(v any) zap.Field {
	if out, err := json.Marshal(v); err != nil {
		return zap.String(JSONKey+"_error", err.Error())
	} else {
		raw := json.RawMessage(out)
		return zap.Any(JSONKey, &raw)
	}
}
