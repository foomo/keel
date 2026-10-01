package markdown

import (
	"fmt"
)

// Code wraps v in backticks as inline code, or returns "" if v is empty.
func Code(v string) string {
	if v == "" {
		return ""
	}

	return "`" + v + "`"
}

// Name returns the result of v.Name() if v has such a method, or "".
func Name(v any) string {
	if i, ok := v.(interface {
		Name() string
	}); ok {
		return i.Name()
	}

	return ""
}

// String returns v.String() if v implements [fmt.Stringer], or "".
func String(v any) string {
	if i, ok := v.(fmt.Stringer); ok {
		return i.String()
	}

	return ""
}
