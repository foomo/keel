package runtimeutil

import (
	"runtime"
	"strings"
)

func Caller(skip int) (shortName, fullName, file string, line int, ok bool) { //nolint:nonamedreturns
	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return "unknown", "Unknown", "unknown", 0, false
	}

	fullName = runtime.FuncForPC(pc).Name()

	// Split fullName by last slash to separate package path and the rest
	lastSlash := strings.LastIndex(fullName, "/")
	if lastSlash != -1 {
		return fullName[lastSlash+1:], fullName, file, line, true
	}

	return fullName, fullName, file, line, true
}
