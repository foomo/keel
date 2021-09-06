package keel

import (
	"encoding/json"
	"net/http"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/foomo/keel/config"
	"github.com/foomo/keel/log"
)

const (
	DefaultServiceHTTPViperName = "viper"
	DefaultServiceHTTPViperAddr = "localhost:9300"
	DefaultServiceHTTPViperPath = "/config"
)

func NewServiceHTTPViper(l *zap.Logger, c *viper.Viper, name, addr, path string) *ServiceHTTP {
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
	return NewServiceHTTP(l, name, addr, handler)
}

func NewDefaultServiceHTTPViper() *ServiceHTTP {
	return NewServiceHTTPViper(
		log.Logger(),
		config.Config(),
		DefaultServiceHTTPViperName,
		DefaultServiceHTTPViperAddr,
		DefaultServiceHTTPViperPath,
	)
}
