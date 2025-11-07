package service

import (
	"net/http"
	"net/http/pprof"

	pyroscope_pprof "github.com/grafana/pyroscope-go/http/pprof"
	"go.uber.org/zap"
)

var (
	DefaultHTTPPProfName = "pprof"
	DefaultHTTPPProfAddr = "localhost:6060"
	DefaultHTTPPProfPath = "/debug/pprof"
)

func NewHTTPPProf(l *zap.Logger, name, addr, path string) *HTTP {
	handler := http.NewServeMux()
	handler.HandleFunc(path+"/", pprof.Index)
	handler.HandleFunc(path+"/cmdline", pprof.Cmdline)
	handler.HandleFunc(path+"/profile", pyroscope_pprof.Profile)
	handler.HandleFunc(path+"/symbol", pprof.Symbol)
	handler.HandleFunc(path+"/trace", pprof.Trace)

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
