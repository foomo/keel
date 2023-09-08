package service

import (
	"encoding/json"
	"net/http"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/foomo/keel/config"
	"github.com/foomo/keel/log"
)

const (
	DefaultHTTPViperName = "viper"
	DefaultHTTPViperAddr = "localhost:9300"
	DefaultHTTPViperPath = "/config"
)

func NewHTTPViper(l *zap.Logger, c *viper.Viper, name, addr, path string) *HTTP {
	handler := http.NewServeMux()
	handler.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		type payload struct {
			Key   string      `json:"key"`
			Value interface{} `json:"value"`
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

func NewDefaultHTTPViper() *HTTP {
	return NewHTTPViper(
		log.Logger(),
		config.Config(),
		DefaultHTTPViperName,
		DefaultHTTPViperAddr,
		DefaultHTTPViperPath,
	)
}
