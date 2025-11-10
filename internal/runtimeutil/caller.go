package runtimeutil

import (
	"fmt"
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

func StackTrace(num, skip int) string {
	pcs := make([]uintptr, num)
	n := runtime.Callers(skip+1, pcs)
	pcs = pcs[:n]

	var ret string

	frames := runtime.CallersFrames(pcs)
	for {
		frame, more := frames.Next()

		ret += fmt.Sprintf("%s\n  %s:%d\n", frame.Function, frame.File, frame.Line)
		if !more || len(ret) == num {
			break
		}
	}

	return strings.Trim(ret, "\n")
}
