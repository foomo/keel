package markdown

import (
	"fmt"
)

func Code(v string) string {
	if v == "" {
		return ""
	}

	return "`" + v + "`"
}

func Name(v any) string {
	if i, ok := v.(interface {
		Name() string
	}); ok {
		return i.Name()
	}

	return ""
}

func String(v any) string {
	if i, ok := v.(fmt.Stringer); ok {
		return i.String()
	}

	return ""
}
