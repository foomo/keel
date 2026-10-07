package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/foomo/keel/healthz"
	"github.com/foomo/keel/interfaces"
	"github.com/foomo/keel/log"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.uber.org/zap"
)

// Default name, address and path of the service returned by
// [NewDefaultHTTPProbes].
var (
	DefaultHTTPHealthzName = "healthz"
	DefaultHTTPHealthzAddr = ":9400"
	DefaultHTTPHealthzPath = "/healthz"
)

// Errors reported by the health probe handlers.
var (
	// ErrUnhandledHealthzProbe is reported for a probe of an unsupported type.
	ErrUnhandledHealthzProbe = errors.New("unhandled healthz probe")
	// ErrProbeFailed is reported when a boolean probe fails on the aggregate path.
	ErrProbeFailed = errors.New("probe failed")
	// ErrLivenessProbeFailed is reported when a boolean liveness probe fails.
	ErrLivenessProbeFailed = errors.New("liveness probe failed")
	// ErrReadinessProbeFailed is reported when a boolean readiness probe fails.
	ErrReadinessProbeFailed = errors.New("readiness probe failed")
	// ErrStartupProbeFailed is reported when a boolean startup probe fails.
	ErrStartupProbeFailed = errors.New("startup probe failed")
)

// NewHealthz returns an [HTTP] service serving health checks for probes:
//
//   - path runs all probes except [healthz.TypeStartup] ones
//   - path/liveness runs [healthz.TypeAlways] and [healthz.TypeLiveness] probes
//   - path/readiness runs [healthz.TypeAlways] and [healthz.TypeReadiness] probes
//   - path/startup runs [healthz.TypeAlways] and [healthz.TypeStartup] probes
//
// Each endpoint responds 200 OK when all probes pass and 503 Service Unavailable
// at the first failing probe. probes is read on every request, so probes added
// to the map later are included.
func NewHealthz(l *zap.Logger, name, addr, path string, probes map[healthz.Type][]any) *HTTP {
	handler := http.NewServeMux()

	unavailable := func(l *zap.Logger, w http.ResponseWriter, r *http.Request, err error) {
		if err != nil {
			log.WithError(l, err).With(log.Attribute(semconv.URLFull(r.RequestURI))).Debug("healthz probe failed")
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		}
	}

	call := func(ctx context.Context, probe any) (bool, error) {
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
		var ps []any
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
		var ps []any
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
		var ps []any
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

// NewDefaultHTTPProbes returns [NewHealthz] using [DefaultHTTPHealthzName],
// [DefaultHTTPHealthzAddr] and [DefaultHTTPHealthzPath].
func NewDefaultHTTPProbes(l *zap.Logger, probes map[healthz.Type][]any) *HTTP {
	return NewHealthz(
		l,
		DefaultHTTPHealthzName,
		DefaultHTTPHealthzAddr,
		DefaultHTTPHealthzPath,
		probes,
	)
}
