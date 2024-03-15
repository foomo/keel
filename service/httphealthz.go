package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/foomo/keel/healthz"
	"github.com/foomo/keel/interfaces"
	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

const (
	DefaultHTTPHealthzName = "healthz"
	DefaultHTTPHealthzAddr = ":9400"
	DefaultHTTPHealthzPath = "/healthz"
)

var (
	ErrUnhandledHealthzProbe = errors.New("unhandled healthz probe")
	ErrProbeFailed           = errors.New("probe failed")
	ErrLivenessProbeFailed   = errors.New("liveness probe failed")
	ErrReadinessProbeFailed  = errors.New("readiness probe failed")
	ErrStartupProbeFailed    = errors.New("startup probe failed")
)

func NewHealthz(l *zap.Logger, name, addr, path string, probes map[healthz.Type][]interface{}) *HTTP {
	handler := http.NewServeMux()

	unavailable := func(l *zap.Logger, w http.ResponseWriter, r *http.Request, err error) {
		if err != nil {
			log.WithHTTPRequest(l, r).Info("http healthz server", log.FError(err), log.FHTTPStatusCode(http.StatusServiceUnavailable))
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		}
	}

	call := func(ctx context.Context, probe interface{}) (bool, error) {
		switch h := probe.(type) {
		case healthz.BoolHealthzer:
			return h.Healthz(), nil
		case healthz.BoolHealthzerWithContext:
			return h.Healthz(ctx), nil
		case healthz.ErrorHealthzer:
			return true, h.Healthz()
		case healthz.ErrorHealthzWithContext:
			return true, h.Healthz(ctx)
		case interfaces.ErrorPinger:
			return true, h.Ping()
		case interfaces.ErrorPingerWithContext:
			return true, h.Ping(ctx)
		default:
			return false, ErrUnhandledHealthzProbe
		}
	}

	handler.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		for typ, values := range probes {
			if typ == healthz.TypeStartup {
				continue
			}
			for _, p := range values {
				if ok, err := call(r.Context(), p); err != nil {
					unavailable(l, w, r, err)
					return
				} else if !ok {
					unavailable(l, w, r, ErrProbeFailed)
					return
				}
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	handler.HandleFunc(path+"/"+healthz.TypeLiveness.String(), func(w http.ResponseWriter, r *http.Request) {
		var ps []interface{}
		if p, ok := probes[healthz.TypeAlways]; ok {
			ps = append(ps, p...)
		}
		if p, ok := probes[healthz.TypeLiveness]; ok {
			ps = append(ps, p...)
		}
		for _, p := range ps {
			if ok, err := call(r.Context(), p); err != nil {
				unavailable(l, w, r, err)
				return
			} else if !ok {
				unavailable(l, w, r, ErrLivenessProbeFailed)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	handler.HandleFunc(path+"/"+healthz.TypeReadiness.String(), func(w http.ResponseWriter, r *http.Request) {
		var ps []interface{}
		if p, ok := probes[healthz.TypeAlways]; ok {
			ps = append(ps, p...)
		}
		if p, ok := probes[healthz.TypeReadiness]; ok {
			ps = append(ps, p...)
		}
		for _, p := range ps {
			if ok, err := call(r.Context(), p); err != nil {
				unavailable(l, w, r, err)
				return
			} else if !ok {
				unavailable(l, w, r, ErrReadinessProbeFailed)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	handler.HandleFunc(path+"/"+healthz.TypeStartup.String(), func(w http.ResponseWriter, r *http.Request) {
		var ps []interface{}
		if p, ok := probes[healthz.TypeAlways]; ok {
			ps = append(ps, p...)
		}
		if p, ok := probes[healthz.TypeStartup]; ok {
			ps = append(ps, p...)
		}
		for _, p := range ps {
			if ok, err := call(r.Context(), p); err != nil {
				unavailable(l, w, r, err)
				return
			} else if !ok {
				unavailable(l, w, r, ErrStartupProbeFailed)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	return NewHTTP(l, name, addr, handler)
}

func NewDefaultHTTPProbes(l *zap.Logger, probes map[healthz.Type][]interface{}) *HTTP {
	return NewHealthz(
		l,
		DefaultHTTPHealthzName,
		DefaultHTTPHealthzAddr,
		DefaultHTTPHealthzPath,
		probes,
	)
}
