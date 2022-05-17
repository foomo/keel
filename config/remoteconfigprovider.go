package config

import (
	"bytes"
	"io"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

func init() {
	viper.RemoteConfig = &remoteConfigProvider{}
}

type remoteConfigProvider struct{}

// Get interface method
func (c *remoteConfigProvider) Get(rp viper.RemoteProvider) (io.Reader, error) {
	cm, err := getConfigManager(rp)
	if err != nil {
		return nil, err
	}
	b, err := cm.Get(rp.Path())
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

// Watch interface method
func (c *remoteConfigProvider) Watch(rp viper.RemoteProvider) (io.Reader, error) {
	cm, err := getConfigManager(rp)
	if err != nil {
		return nil, err
	}
	resp, err := cm.Get(rp.Path())
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(resp), nil
}

// WatchChannel interface method
func (c *remoteConfigProvider) WatchChannel(rp viper.RemoteProvider) (<-chan *viper.RemoteResponse, chan bool) {
	cm, err := getConfigManager(rp)
	if err != nil {
		return nil, nil
	}
	quit := make(chan bool)
	responseCh := cm.Watch(rp.Path(), quit)
	return responseCh, quit
}

func getConfigManager(rp viper.RemoteProvider) (remoteConfigManager, error) {
	var cm remoteConfigManager
	if rp.SecretKeyring() != "" {
		panic("implement me")
	} else {
		switch rp.Provider() {
		case "etcd":
			cm = NewEtcdConfigManager([]string{rp.Endpoint()})
		default:
			return nil, errors.New("implement me")
		}
	}
	return cm, nil
}
