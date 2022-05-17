package config

import (
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func WithRemoteConfig(c *viper.Viper, provider, endpoint, path string) error {
	if err := c.AddRemoteProvider(provider, endpoint, path); err != nil {
		return errors.Wrap(err, "failed to add remote provider")
	}

	var i int
	for {
		time.Sleep(time.Second)
		if err := c.ReadRemoteConfig(); err != nil && i < 5 {
			i++
		} else if err != nil {
			return errors.Wrap(err, "failed to read remote config")
		} else {
			break
		}
	}

	if err := c.WatchRemoteConfigOnChannel(); err != nil {
		return errors.Wrap(err, "failed to watch remote config")
	}

	return nil
}
