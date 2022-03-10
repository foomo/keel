package keel

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

const (
	DefaultServiceHTTPProbesName = probesServiceName
	DefaultServiceHTTPProbesAddr = ":9400"
	DefaultServiceHTTPProbesPath = "/healthz"
)

// Probe's handler function
type ProbeHandler func(w http.ResponseWriter, r *http.Request)

func NewServiceHTTPProbes(l *zap.Logger, name, addr, path string) *ServiceHTTP {
	handler := http.NewServeMux()
	handler.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})
	return NewServiceHTTP(l, name, addr, handler)
}

func NewDefaultServiceHTTPProbes() *ServiceHTTP {
	return NewServiceHTTPProbes(
		log.Logger(),
		DefaultServiceHTTPProbesName,
		DefaultServiceHTTPProbesAddr,
		DefaultServiceHTTPProbesPath,
	)
}

func CreateProbeHandlers(s *Server) *http.ServeMux {
	handler := http.NewServeMux()
	for _, f := range s.probeHandlers {
		probeHandler := f
		handler.HandleFunc("/healthz/"+probeHandler.probeType, func(w http.ResponseWriter, r *http.Request) {
			switch h := probeHandler.handler.(type) {
			case Health:
				success := h.Ping()
				if success {
					w.WriteHeader(200)
					w.Write([]byte("ok"))
				} else {
					http.Error(w, "Failed to run probe ping", http.StatusInternalServerError)
				}
			case HealthFn:
				success := h()
				if success {
					w.WriteHeader(200)
					w.Write([]byte("ok"))
				} else {
					http.Error(w, "Failed to run probe ping", http.StatusInternalServerError)
				}
			case ErrorHealthFn:
				if success, err := h(); err != nil {
					log.WithError(s.l, err).Error("failed to use probe ErrorHealthFn")
					http.Error(w, err.Error(), http.StatusInternalServerError)
				} else if success {
					w.WriteHeader(200)
					w.Write([]byte("ok"))
				} else {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			case ErrorHealth:
				if _, err := h.Ping(); err != nil {
					log.WithError(s.l, err).Error("failed to use probe ErrorHealth")
				}
			case HealthWithContextFn:
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
