package config

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote" // required import
)

// remotes records every remote provider added through [WithRemoteConfig].
var remotes []struct {
	provider string
	endpoint string
	path     string
}

// WithRemoteConfig adds a remote provider (e.g. "etcd3", "consul") at
// endpoint and path to c, reads it and starts watching it for changes. It
// returns an error if any of these steps fail.
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

	remotes = append(remotes, struct {
		provider string
		endpoint string
		path     string
	}{provider: provider, endpoint: endpoint, path: path})

	return nil
}
