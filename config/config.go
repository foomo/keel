package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// config holds the global configuration
var config *viper.Viper

// Init sets up the configuration
func init() {
	config = viper.New()
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.AutomaticEnv()
}

// Config return the config instance
func Config() *viper.Viper {
	return config
}

func GetBool(c *viper.Viper, key string, fallback bool) func() bool {
	c.SetDefault(key, fallback)
	return func() bool {
		return c.GetBool(key)
	}
}

func MustGetBool(c *viper.Viper, key string, fallback bool) func() bool {
	must(c, key)
	return func() bool {
		return c.GetBool(key)
	}
}

func GetString(c *viper.Viper, key, fallback string) func() string {
	c.SetDefault(key, fallback)
	return func() string {
		return c.GetString(key)
	}
}

func MustGetString(c *viper.Viper, key string) func() string {
	must(c, key)
	return func() string {
		return c.GetString(key)
	}
}

func GetStringSlice(c *viper.Viper, key string, fallback []string) func() []string {
	c.SetDefault(key, fallback)
	return func() []string {
		return c.GetStringSlice(key)
	}
}

func must(c *viper.Viper, key string) {
	if !c.InConfig(key) {
		panic(fmt.Sprintf("missing required config key: %s", key))
	}
}
