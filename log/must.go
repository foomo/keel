package log

import (
	"fmt"

	"go.uber.org/zap"
)

// Must logs err at fatal level and exits the process if err is not nil. The
// optional msgAndArgs are a format string followed by its arguments and
// default to "Must". A nil l selects [Logger].
func Must(l *zap.Logger, err error, msgAndArgs ...any) {
	if err != nil {
		if l == nil {
			l = Logger()
		}

		var msg = "Must"
		if len(msgAndArgs) > 0 {
			msg, msgAndArgs = fmt.Sprintf("%v", msgAndArgs[0]), msgAndArgs[1:]
		}

		WithError(l, err).WithOptions(zap.AddCallerSkip(1)).Fatal(fmt.Sprintf(msg, msgAndArgs...))
	}
}
