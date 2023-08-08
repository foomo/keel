//go:build pprof
// +build pprof

package keel

import (
	"net/http"
	"net/http/pprof"

	keelconfig "github.com/foomo/keel/config"
	"github.com/foomo/keel/log"
	"go.uber.org/zap"
)

const (
	DefaultServiceHTTPPProfName = "pprof"
	DefaultServiceHTTPPProfAddr = "localhost:6060"
	DefaultServiceHTTPPProfPath = "/debug/pprof"
)

func NewServiceHTTPPProf(l *zap.Logger, name, addr, path string) *ServiceHTTP {
	handler := http.NewServeMux()
	handler.HandleFunc(path+"/", pprof.Index)
	handler.HandleFunc(path+"/cmdline", pprof.Cmdline)
	handler.HandleFunc(path+"/profile", pprof.Profile)
	handler.HandleFunc(path+"/symbol", pprof.Symbol)
	handler.HandleFunc(path+"/trace", pprof.Trace)
	return NewServiceHTTP(l, name, addr, handler)
}

func NewDefaultServiceHTTPPProf() *ServiceHTTP {
	return NewServiceHTTPPProf(
		log.Logger(),
		DefaultServiceHTTPPProfName,
		DefaultServiceHTTPPProfAddr,
		DefaultServiceHTTPPProfPath,
	)
}

// WithPProfService option with default value
func WithPProfService(enabled bool) Option {
	return func(inst *Server) {
		if keelconfig.GetBool(inst.Config(), "service.pprof.enabled", enabled)() {
			service := NewDefaultServiceHTTPPProf()
			inst.initServices = append(inst.initServices, service)
			inst.AddAlwaysHealthzers(service)
		}
	}
}
