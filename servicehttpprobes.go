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
		// TODO implement
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

func CreateProbeHandlers(s *Server) *http.ServeMux {
	handler := http.NewServeMux()
	for _, f := range s.probes {
		probeHandler := f
		handler.HandleFunc("/healthz/"+probeHandler.probeType, func(w http.ResponseWriter, r *http.Request) {
			switch h := probeHandler.handler.(type) {
			case BoolProbeFn:
				success := h()
				if success {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("OK"))
				} else {
					http.Error(w, "Failed to run probe ping", http.StatusInternalServerError)
				}
			case ErrorProbeFn:
				if success, err := h(); err != nil {
					log.WithError(s.l, err).Error("failed to use probe ErrorHealthFn")
					http.Error(w, err.Error(), http.StatusInternalServerError)
				} else if success {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("OK"))
				} else {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			case ErrorPingProbe:
				success := h.Ping()
				if success {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("OK"))
				} else {
					http.Error(w, "Failed to run probe ping", http.StatusInternalServerError)
				}
			case ErrorHealth:
				if _, err := h.Ping(); err != nil {
					log.WithError(s.l, err).Error("failed to use probe ErrorHealth")
				}
			case BoolProbeWithContextFn:
				h(context.Background())
			case HealthWithContext:
				h.Ping(context.Background())
			case ErrorHealthWithContext:
				if _, err := h.Ping(context.Background()); err != nil {
					log.WithError(s.l, err).Error("failed to use probe ErrorHealthWithContext")
				}
			}
		})
	}
	return handler
}
