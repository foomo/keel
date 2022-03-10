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

// Probe's handler function
type ProbeHandler func(w http.ResponseWriter, r *http.Request)

func NewServiceHTTPProbes(l *zap.Logger, name, addr, path string, probes Probes) *ServiceHTTP {
	handler := http.NewServeMux()

	call := func(probe interface{}) bool {
		switch h := probe.(type) {
		case BoolProbeFn:
			return h()
		case ErrorProbeFn:
			if err := h(); err != nil {
				log.WithError(l, err).Error("failed to use probe ErrorHealthFn")
				return false
			}
		case ErrorPingProbe:
			if err := h.Ping(); err != nil {
				log.WithError(l, err).Error("failed to use probe ErrorPingProbe")
				return false
			}
		case ErrorPingProbeWithContext:
			if err := h.Ping(context.Background()); err != nil {
				log.WithError(l, err).Error("failed to use probe ErrorHealth")
				return false
			}
		case BoolProbeWithContextFn:
			return h(context.Background())
		case ErrorProbeWithContextFn:
			if err := h(context.Background()); err != nil {
				log.WithError(l, err).Error("failed to use probe ErrorHealthWithContext")
				return false
			}
		}
		return true
	}

	handler.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		for _, values := range probes {
			for _, p := range values {
				if !call(p) {
					httputils.InternalServiceUnavailable(l, w, r, errors.New("not ready yet"))
					return
				}
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	handler.HandleFunc(path+"/"+ProbeTypeLiveliness.String(), func(w http.ResponseWriter, r *http.Request) {
		var ps []interface{}
		if p, ok := probes[ProbeTypeAll]; ok {
			ps = append(ps, p...)
		}
		if p, ok := probes[ProbeTypeLiveliness]; ok {
			ps = append(ps, p...)
		}
		for _, p := range ps {
			if !call(p) {
				httputils.InternalServiceUnavailable(l, w, r, errors.New("not ready yet"))
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	handler.HandleFunc(path+"/"+ProbeTypeReadiness.String(), func(w http.ResponseWriter, r *http.Request) {
		var ps []interface{}
		if p, ok := probes[ProbeTypeAll]; ok {
			ps = append(ps, p...)
		}
		if p, ok := probes[ProbeTypeReadiness]; ok {
			ps = append(ps, p...)
		}
		for _, p := range ps {
			if !call(p) {
				httputils.InternalServiceUnavailable(l, w, r, errors.New("not ready yet"))
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	handler.HandleFunc(path+"/"+ProbeTypeStartup.String(), func(w http.ResponseWriter, r *http.Request) {
		var ps []interface{}
		if p, ok := probes[ProbeTypeAll]; ok {
			ps = append(ps, p...)
		}
		if p, ok := probes[ProbeTypeStartup]; ok {
			ps = append(ps, p...)
		}
		for _, p := range ps {
			if !call(p) {
				httputils.InternalServiceUnavailable(l, w, r, errors.New("not ready yet"))
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
