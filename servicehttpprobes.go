package keel

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/foomo/keel/log"
	httputils "github.com/foomo/keel/utils/net/http"
)

const (
	DefaultServiceHTTPProbesName = probesServiceName
	DefaultServiceHTTPProbesAddr = ":9400"
	DefaultServiceHTTPProbesPath = "/healthz"
)

func NewServiceHTTPProbes(l *zap.Logger, name, addr, path string, probes Probes) *ServiceHTTP {
	handler := http.NewServeMux()

	call := func(ctx context.Context, probe interface{}) (bool, error) {
		switch h := probe.(type) {
		case BoolHealthz:
			return h.Healthz(), nil
		case BoolHealthzWithContext:
			return h.Healthz(ctx), nil
		case ErrorHealthz:
			return true, h.Healthz()
		case ErrorHealthzWithContext:
			return true, h.Healthz(ctx)
		case ErrorPingProbe:
			return true, h.Ping()
		case ErrorPingProbeWithContext:
			return true, h.Ping(ctx)
		default:
			return false, errors.New("unhandled probe")
		}
	}

	handler.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		for typ, values := range probes {
			if typ == ProbeTypeStartup {
				continue
			}
			for _, p := range values {
				if ok, err := call(r.Context(), p); err != nil {
					httputils.InternalServiceUnavailable(l, w, r, err)
					return
				} else if !ok {
					httputils.InternalServiceUnavailable(l, w, r, errors.New("probe failed"))
					return
				}
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	handler.HandleFunc(path+"/"+ProbeTypeLiveness.String(), func(w http.ResponseWriter, r *http.Request) {
		var ps []interface{}
		if p, ok := probes[ProbeTypeAny]; ok {
			ps = append(ps, p...)
		}
		if p, ok := probes[ProbeTypeLiveness]; ok {
			ps = append(ps, p...)
		}
		for _, p := range ps {
			if ok, err := call(r.Context(), p); err != nil {
				httputils.InternalServiceUnavailable(l, w, r, err)
				return
			} else if !ok {
				httputils.InternalServiceUnavailable(l, w, r, errors.New("liveness probe failed"))
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	handler.HandleFunc(path+"/"+ProbeTypeReadiness.String(), func(w http.ResponseWriter, r *http.Request) {
		var ps []interface{}
		if p, ok := probes[ProbeTypeAny]; ok {
			ps = append(ps, p...)
		}
		if p, ok := probes[ProbeTypeReadiness]; ok {
			ps = append(ps, p...)
		}
		for _, p := range ps {
			if ok, err := call(r.Context(), p); err != nil {
				httputils.InternalServiceUnavailable(l, w, r, err)
				return
			} else if !ok {
				httputils.InternalServiceUnavailable(l, w, r, errors.New("readiness probe failed"))
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	handler.HandleFunc(path+"/"+ProbeTypeStartup.String(), func(w http.ResponseWriter, r *http.Request) {
		var ps []interface{}
		if p, ok := probes[ProbeTypeAny]; ok {
			ps = append(ps, p...)
		}
		if p, ok := probes[ProbeTypeStartup]; ok {
			ps = append(ps, p...)
		}
		for _, p := range ps {
			if ok, err := call(r.Context(), p); err != nil {
				httputils.InternalServiceUnavailable(l, w, r, err)
				return
			} else if !ok {
				httputils.InternalServiceUnavailable(l, w, r, errors.New("startup probe failed"))
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	return NewServiceHTTP(l, name, addr, handler)
}

func NewDefaultServiceHTTPProbes(probes Probes) *ServiceHTTP {
	return NewServiceHTTPProbes(
		log.Logger(),
		DefaultServiceHTTPProbesName,
		DefaultServiceHTTPProbesAddr,
		DefaultServiceHTTPProbesPath,
		probes,
	)
}
