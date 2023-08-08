//go:build !pprof

package keel

import (
	keelconfig "github.com/foomo/keel/config"
	"github.com/foomo/keel/log"
)

// WithPProfService option with default value
func WithPProfService(enabled bool) Option {
	return func(inst *Server) {
		if keelconfig.GetBool(inst.Config(), "service.pprof.enabled", enabled)() {
			log.Logger().Debug("build your binary with the `-tag pprof` to enable this service")
		}
	}
}
