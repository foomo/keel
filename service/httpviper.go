package service

import (
	"encoding/json"
	"net/http"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/foomo/keel/config"
)

// Default name, address and path of the service returned by
// [NewDefaultHTTPViper].
var (
	DefaultHTTPViperName = "viper"
	DefaultHTTPViperAddr = "localhost:9300"
	DefaultHTTPViperPath = "/config"
)

// NewHTTPViper returns an [HTTP] service exposing c on path. GET responds with
// all settings as JSON; PUT sets a single value from a {"key": ..., "value": ...}
// JSON body.
func NewHTTPViper(l *zap.Logger, c *viper.Viper, name, addr, path string) *HTTP {
	handler := http.NewServeMux()
	handler.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		type payload struct {
			Key   string `json:"key"`
			Value any    `json:"value"`
		}

		enc := json.NewEncoder(w)

		switch r.Method {
		case http.MethodGet:
			if err := enc.Encode(c.AllSettings()); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case http.MethodPut:
			var req payload

			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			c.Set(req.Key, req.Value)
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	})

	return NewHTTP(l, name, addr, handler)
}

// NewDefaultHTTPViper returns [NewHTTPViper] for [config.Config] using
// [DefaultHTTPViperName], [DefaultHTTPViperAddr] and [DefaultHTTPViperPath].
func NewDefaultHTTPViper(l *zap.Logger) *HTTP {
	return NewHTTPViper(
		l,
		config.Config(),
		DefaultHTTPViperName,
		DefaultHTTPViperAddr,
		DefaultHTTPViperPath,
	)
}
