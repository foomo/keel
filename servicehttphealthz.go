package keel

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

const (
	DefaultServiceHTTPProbesName = healthzServiceName
	DefaultServiceHTTPProbesAddr = ":9400"
	DefaultServiceHTTPProbesPath = "/healthz"
)

func NewServiceHTTPHealthz(l *zap.Logger, name, addr, path string, probes map[HealthzType][]interface{}) *ServiceHTTP {
	handler := http.NewServeMux()

	unavailable := func(l *zap.Logger, w http.ResponseWriter, r *http.Request, err error) {
		if err != nil {
			log.WithHTTPRequest(l, r).Info("http healthz server", log.FError(err), log.FHTTPStatusCode(http.StatusServiceUnavailable))
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		}
	}

	call := func(ctx context.Context, probe interface{}) (bool, error) {
		switch h := probe.(type) {
		case BoolHealthzer:
			return h.Healthz(), nil
		case BoolHealthzerWithContext:
			return h.Healthz(ctx), nil
		case ErrorHealthzer:
			return true, h.Healthz()
		case ErrorHealthzWithContext:
			return true, h.Healthz(ctx)
		case ErrorPinger:
			return true, h.Ping()
		case ErrorPingerWithContext:
			return true, h.Ping(ctx)
		default:
			return false, errors.New("unhandled healthz probe")
		}
	}

	handler.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		for typ, values := range probes {
			if typ == HealthzTypeStartup {
				continue
			}
			for _, p := range values {
				if ok, err := call(r.Context(), p); err != nil {
					unavailable(l, w, r, err)
					return
				} else if !ok {
					unavailable(l, w, r, errors.New("probe failed"))
					return
				}
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	handler.HandleFunc(path+"/"+HealthzTypeLiveness.String(), func(w http.ResponseWriter, r *http.Request) {
		var ps []interface{}
		if p, ok := probes[HealthzTypeAny]; ok {
			ps = append(ps, p...)
		}
		if p, ok := probes[HealthzTypeLiveness]; ok {
			ps = append(ps, p...)
		}
		for _, p := range ps {
			if ok, err := call(r.Context(), p); err != nil {
				unavailable(l, w, r, err)
				return
			} else if !ok {
				unavailable(l, w, r, errors.New("liveness probe failed"))
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	handler.HandleFunc(path+"/"+HealthzTypeReadiness.String(), func(w http.ResponseWriter, r *http.Request) {
		var ps []interface{}
		if p, ok := probes[HealthzTypeAny]; ok {
			ps = append(ps, p...)
		}
		if p, ok := probes[HealthzTypeReadiness]; ok {
			ps = append(ps, p...)
		}
		for _, p := range ps {
			if ok, err := call(r.Context(), p); err != nil {
				unavailable(l, w, r, err)
				return
			} else if !ok {
				unavailable(l, w, r, errors.New("readiness probe failed"))
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	handler.HandleFunc(path+"/"+HealthzTypeStartup.String(), func(w http.ResponseWriter, r *http.Request) {
		var ps []interface{}
		if p, ok := probes[HealthzTypeAny]; ok {
			ps = append(ps, p...)
		}
		if p, ok := probes[HealthzTypeStartup]; ok {
			ps = append(ps, p...)
		}
		for _, p := range ps {
			if ok, err := call(r.Context(), p); err != nil {
				unavailable(l, w, r, err)
				return
			} else if !ok {
				unavailable(l, w, r, errors.New("startup probe failed"))
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	return NewServiceHTTP(l, name, addr, handler)
}

func NewDefaultServiceHTTPProbes(probes map[HealthzType][]interface{}) *ServiceHTTP {
	return NewServiceHTTPHealthz(
		log.Logger(),
		DefaultServiceHTTPProbesName,
		DefaultServiceHTTPProbesAddr,
		DefaultServiceHTTPProbesPath,
		probes,
	)
}
