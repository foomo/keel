package service

import (
	"net/http"
	"net/http/pprof"

	pyroscope_pprof "github.com/grafana/pyroscope-go/http/pprof"
	"go.uber.org/zap"
)

// Default name, address and path of the service returned by
// [NewDefaultHTTPPProf].
var (
	DefaultHTTPPProfName = "pprof"
	DefaultHTTPPProfAddr = "localhost:6060"
	DefaultHTTPPProfPath = "/debug/pprof"
)

// NewHTTPPProf returns an [HTTP] service exposing the [net/http/pprof]
// handlers below path. The CPU profile endpoint is served by Pyroscope's
// pprof handler.
func NewHTTPPProf(l *zap.Logger, name, addr, path string) *HTTP {
	handler := http.NewServeMux()
	handler.HandleFunc(path+"/", pprof.Index)
	handler.HandleFunc(path+"/cmdline", pprof.Cmdline)
	handler.HandleFunc(path+"/profile", pyroscope_pprof.Profile)
	handler.HandleFunc(path+"/symbol", pprof.Symbol)
	handler.HandleFunc(path+"/trace", pprof.Trace)

	return NewHTTP(l, name, addr, handler)
}

// NewDefaultHTTPPProf returns [NewHTTPPProf] using [DefaultHTTPPProfName],
// [DefaultHTTPPProfAddr] and [DefaultHTTPPProfPath].
func NewDefaultHTTPPProf(l *zap.Logger) *HTTP {
	return NewHTTPPProf(
		l,
		DefaultHTTPPProfName,
		DefaultHTTPPProfAddr,
		DefaultHTTPPProfPath,
	)
}
