package config

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

func WithRemoteConfig(c *viper.Viper, provider, endpoint string, path string) error {
	if err := c.AddRemoteProvider(provider, endpoint, path); err != nil {
		return err
	}

	if err := c.ReadRemoteConfig(); err != nil {
		return errors.Wrap(err, "failed to read remote config")
	}

	if err := c.WatchRemoteConfigOnChannel(); err != nil {
		return errors.Wrap(err, "failed to watch remote config")
	}

	return nil
}
