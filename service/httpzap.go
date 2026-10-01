package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/foomo/keel/log"
)

// Default name, address and path of the service returned by
// [NewDefaultHTTPZap].
var (
	DefaultHTTPZapName = "zap"
	DefaultHTTPZapAddr = "localhost:9100"
	DefaultHTTPZapPath = "/log"
)

// NewHTTPZap returns an [HTTP] service controlling the global logger on path.
// GET responds with the level, disableCaller and disableStacktrace settings as
// JSON; PUT updates any of them from a JSON body of the same shape and responds
// with the resulting settings.
func NewHTTPZap(l *zap.Logger, name, addr, path string) *HTTP {
	handler := http.NewServeMux()
	handler.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		type errorResponse struct {
			Error string `json:"error"`
		}

		type payload struct {
			Level             *zapcore.Level `json:"level"`
			DisableCaller     *bool          `json:"disableCaller"`
			DisableStacktrace *bool          `json:"disableStacktrace"`
		}

		enc := json.NewEncoder(w)

		switch r.Method {
		case http.MethodGet:
			current := log.AtomicLevel().Level()
			disableCaller := log.IsDisableCaller()
			disableStacktrace := log.IsDisableStacktrace()
			_ = enc.Encode(payload{
				Level:             &current,
				DisableCaller:     &disableCaller,
				DisableStacktrace: &disableStacktrace,
			})

		case http.MethodPut:
			var req payload

			if errmess := func() string {
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					return fmt.Sprintf("Request body must be well-formed JSON: %v", err)
				}

				if req.Level == nil && req.DisableCaller == nil && req.DisableStacktrace == nil {
					return "Must specify a value."
				}

				return ""
			}(); errmess != "" {
				w.WriteHeader(http.StatusBadRequest)

				_ = enc.Encode(errorResponse{Error: errmess})

				return
			}

			if req.Level != nil {
				log.AtomicLevel().SetLevel(*req.Level)
			}

			if req.DisableCaller != nil {
				if err := log.SetDisableCaller(*req.DisableCaller); err != nil {
					w.WriteHeader(http.StatusInternalServerError)

					_ = enc.Encode(errorResponse{Error: err.Error()})

					return
				}
			}

			if req.DisableStacktrace != nil {
				if err := log.SetDisableStacktrace(*req.DisableStacktrace); err != nil {
					w.WriteHeader(http.StatusInternalServerError)

					_ = enc.Encode(errorResponse{Error: err.Error()})

					return
				}
			}

			current := log.AtomicLevel().Level()
			disableCaller := log.IsDisableCaller()
			disableStacktrace := log.IsDisableStacktrace()
			_ = enc.Encode(payload{
				Level:             &current,
				DisableCaller:     &disableCaller,
				DisableStacktrace: &disableStacktrace,
			})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)

			_ = enc.Encode(errorResponse{
				Error: "Only GET and PUT are supported.",
			})
		}
	})

	return NewHTTP(l, name, addr, handler)
}

// NewDefaultHTTPZap returns [NewHTTPZap] using [DefaultHTTPZapName],
// [DefaultHTTPZapAddr] and [DefaultHTTPZapPath].
func NewDefaultHTTPZap(l *zap.Logger) *HTTP {
	return NewHTTPZap(
		l,
		DefaultHTTPZapName,
		DefaultHTTPZapAddr,
		DefaultHTTPZapPath,
	)
}
