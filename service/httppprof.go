//go:build !pprof

package service

import (
	"net/http"

	"go.uber.org/zap"
)

var (
	DefaultHTTPPProfName = "pprof"
	DefaultHTTPPProfAddr = "localhost:6060"
	DefaultHTTPPProfPath = "/debug/pprof"
)

func NewHTTPPProf(l *zap.Logger, name, addr, path string) *HTTP {
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

	return NewHTTP(l, name, addr, handler)
}

func NewDefaultHTTPPProf(l *zap.Logger) *HTTP {
	return NewHTTPPProf(
		l,
		DefaultHTTPPProfName,
		DefaultHTTPPProfAddr,
		DefaultHTTPPProfPath,
	)
}
