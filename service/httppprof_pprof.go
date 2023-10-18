//go:build pprof
// +build pprof

package service

import (
	"net/http"
	"net/http/pprof"

	"github.com/foomo/keel/log"
	"go.uber.org/zap"
)

const (
	DefaultHTTPPProfName = "pprof"
	DefaultHTTPPProfAddr = "localhost:6060"
	DefaultHTTPPProfPath = "/debug/pprof"
)

func NewHTTPPProf(l *zap.Logger, name, addr, path string) *ServiceHTTP {
	handler := http.NewServeMux()
	handler.HandleFunc(path+"/", pprof.Index)
	handler.HandleFunc(path+"/cmdline", pprof.Cmdline)
	handler.HandleFunc(path+"/profile", pprof.Profile)
	handler.HandleFunc(path+"/symbol", pprof.Symbol)
	handler.HandleFunc(path+"/trace", pprof.Trace)
	return NewServiceHTTP(l, name, addr, handler)
}

func NewDefaultHTTPPProf() *ServiceHTTP {
	return NewHTTPPProf(
		log.Logger(),
		DefaultHTTPPProfName,
		DefaultHTTPPProfAddr,
		DefaultHTTPPProfPath,
	)
}
