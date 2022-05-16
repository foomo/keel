package config

import (
	"github.com/spf13/viper"
)

type remoteConfigManager interface {
	Get(key string) ([]byte, error)
	Watch(key string, stop chan bool) <-chan *viper.RemoteResponse
}
