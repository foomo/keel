package config

import (
	"github.com/spf13/viper"
)

func WithRemoteConfig(c *viper.Viper, provider, endpoint, path string) error {
	if err := c.AddRemoteProvider(provider, endpoint, path); err != nil {
		return err
	}

	if err := c.ReadRemoteConfig(); err != nil {
		return err
	}

	go func() {
		for {
			if err := c.WatchRemoteConfig(); err != nil {
				panic(err)
			}
			// TODO sth is broken on re-connect
			// if err := c.WatchRemoteConfigOnChannel(); err != nil {
			// 	 panic(err)
			// }
		}
	}()

	return nil
}
