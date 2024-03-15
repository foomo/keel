package keel

import (
	"github.com/foomo/keel/healthz"
	"github.com/foomo/keel/interfaces"
)

func IsHealthz(v any) bool {
	switch v.(type) {
	case healthz.BoolHealthzer,
		healthz.BoolHealthzerWithContext,
		healthz.ErrorHealthzer,
		healthz.ErrorHealthzWithContext,
		interfaces.ErrorPinger,
		interfaces.ErrorPingerWithContext:
		return true
	default:
		return false
	}
}
