//go:build !pprof

package keel

import (
	"net/http"

	"github.com/foomo/keel/log"
	"go.uber.org/zap"
)

const (
	DefaultServiceHTTPPProfName = "pprof"
	DefaultServiceHTTPPProfAddr = "localhost:6060"
	DefaultServiceHTTPPProfPath = "/debug/pprof"
)

func NewServiceHTTPPProf(l *zap.Logger, name, addr, path string) *ServiceHTTP {
	route := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		_, _ = w.Write([]byte("To enable pprof, you need to build your binary with the `-tags=pprof` flag"))
	}
	handler := http.NewServeMux()
	handler.HandleFunc(path+"/", route)
	handler.HandleFunc(path+"/cmdline", route)
	handler.HandleFunc(path+"/profile", route)
	handler.HandleFunc(path+"/symbol", route)
	handler.HandleFunc(path+"/trace", route)
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
