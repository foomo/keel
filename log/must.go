package log

import (
	"fmt"

	"go.uber.org/zap"
)

// Must logs a fatal error if given
func Must(l *zap.Logger, err error, msgAndArgs ...any) {
	if err != nil {
		if l == nil {
			l = Logger()
		}

		var msg = "Must"
		if len(msgAndArgs) > 0 {
			msg, msgAndArgs = fmt.Sprintf("%v", msgAndArgs[0]), msgAndArgs[1:]
		}

		l.WithOptions(zap.AddCallerSkip(1)).Fatal(fmt.Sprintf(msg, msgAndArgs...), FError(err))
	}
}
