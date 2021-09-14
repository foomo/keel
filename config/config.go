package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// config holds the global configuration
var (
	config *viper.Viper
)

// Init sets up the configuration
func init() {
	config = viper.New()
	config.AutomaticEnv()
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

// Config return the config instance
func Config() *viper.Viper {
	return config
}

func GetBool(c *viper.Viper, key string, fallback bool) func() bool {
	c = ensure(c)
	c.SetDefault(key, fallback)
	return func() bool {
		return c.GetBool(key)
	}
}

func MustGetBool(c *viper.Viper, key string, fallback bool) func() bool {
	c = ensure(c)
	must(c, key)
	return func() bool {
		return c.GetBool(key)
	}
}

func GetInt(c *viper.Viper, key string, fallback int) func() int {
	c.SetDefault(key, fallback)
	return func() int {
		return c.GetInt(key)
	}
}

func MustGetInt(c *viper.Viper, key string) func() int {
	must(c, key)
	return func() int {
		return c.GetInt(key)
	}
}

func GetInt32(c *viper.Viper, key string, fallback int32) func() int32 {
	c = ensure(c)
	c.SetDefault(key, fallback)
	return func() int32 {
		return c.GetInt32(key)
	}
}

func MustGetInt32(c *viper.Viper, key string) func() int32 {
	c = ensure(c)
	must(c, key)
	return func() int32 {
		return c.GetInt32(key)
	}
}

func GetInt64(c *viper.Viper, key string, fallback int64) func() int64 {
	c = ensure(c)
	c.SetDefault(key, fallback)
	return func() int64 {
		return c.GetInt64(key)
	}
}

func MustGetInt64(c *viper.Viper, key string) func() int64 {
	c = ensure(c)
	must(c, key)
	return func() int64 {
		return c.GetInt64(key)
	}
}

func GetUint(c *viper.Viper, key string, fallback uint) func() uint {
	c = ensure(c)
	c.SetDefault(key, fallback)
	return func() uint {
		return c.GetUint(key)
	}
}

func MustGetUint(c *viper.Viper, key string) func() uint {
	c = ensure(c)
	must(c, key)
	return func() uint {
		return c.GetUint(key)
	}
}

func GetUint32(c *viper.Viper, key string, fallback uint32) func() uint32 {
	c = ensure(c)
	c.SetDefault(key, fallback)
	return func() uint32 {
		return c.GetUint32(key)
	}
}

func MustGetUint32(c *viper.Viper, key string) func() uint32 {
	c = ensure(c)
	must(c, key)
	return func() uint32 {
		return c.GetUint32(key)
	}
}

func GetUint64(c *viper.Viper, key string, fallback uint64) func() uint64 {
	c = ensure(c)
	c.SetDefault(key, fallback)
	return func() uint64 {
		return c.GetUint64(key)
	}
}

func MustGetUint64(c *viper.Viper, key string) func() uint64 {
	c = ensure(c)
	must(c, key)
	return func() uint64 {
		return c.GetUint64(key)
	}
}

func GetFloat64(c *viper.Viper, key string, fallback float64) func() float64 {
	c = ensure(c)
	c.SetDefault(key, fallback)
	return func() float64 {
		return c.GetFloat64(key)
	}
}

func MustGetFloat64(c *viper.Viper, key string) func() float64 {
	c = ensure(c)
	must(c, key)
	return func() float64 {
		return c.GetFloat64(key)
	}
}

func GetString(c *viper.Viper, key, fallback string) func() string {
	c = ensure(c)
	c.SetDefault(key, fallback)
	return func() string {
		return c.GetString(key)
	}
}

func MustGetString(c *viper.Viper, key string) func() string {
	c = ensure(c)
	must(c, key)
	return func() string {
		return c.GetString(key)
	}
}

func GetTime(c *viper.Viper, key string, fallback time.Time) func() time.Time {
	c = ensure(c)
	c.SetDefault(key, fallback)
	return func() time.Time {
		return c.GetTime(key)
	}
}

func MustGetTime(c *viper.Viper, key string) func() time.Time {
	c = ensure(c)
	must(c, key)
	return func() time.Time {
		return c.GetTime(key)
	}
}

func GetDuration(c *viper.Viper, key string, fallback time.Duration) func() time.Duration {
	c = ensure(c)
	c.SetDefault(key, fallback)
	return func() time.Duration {
		return c.GetDuration(key)
	}
}

func MustGetDuration(c *viper.Viper, key string) func() time.Duration {
	c = ensure(c)
	must(c, key)
	return func() time.Duration {
		return c.GetDuration(key)
	}
}

func GetIntSlice(c *viper.Viper, key string, fallback []int) func() []int {
	c = ensure(c)
	c.SetDefault(key, fallback)
	return func() []int {
		return c.GetIntSlice(key)
	}
}

func MustGetIntSlice(c *viper.Viper, key string) func() []int {
	c = ensure(c)
	must(c, key)
	return func() []int {
		return c.GetIntSlice(key)
	}
}

func GetStringSlice(c *viper.Viper, key string, fallback []string) func() []string {
	c = ensure(c)
	c.SetDefault(key, fallback)
	return func() []string {
		return c.GetStringSlice(key)
	}
}

func MustGetStringSlice(c *viper.Viper, key string) func() []string {
	c = ensure(c)
	must(c, key)
	return func() []string {
		return c.GetStringSlice(key)
	}
}

func GetStringMap(c *viper.Viper, key string, fallback map[string]interface{}) func() map[string]interface{} {
	c = ensure(c)
	c.SetDefault(key, fallback)
	return func() map[string]interface{} {
		return c.GetStringMap(key)
	}
}

func MustGetStringMap(c *viper.Viper, key string) func() map[string]interface{} {
	c = ensure(c)
	must(c, key)
	return func() map[string]interface{} {
		return c.GetStringMap(key)
	}
}

func GetStringMapString(c *viper.Viper, key string, fallback map[string]string) func() map[string]string {
	c = ensure(c)
	c.SetDefault(key, fallback)
	return func() map[string]string {
		return c.GetStringMapString(key)
	}
}

func MustGetStringMapString(c *viper.Viper, key string) func() map[string]string {
	c = ensure(c)
	must(c, key)
	return func() map[string]string {
		return c.GetStringMapString(key)
	}
}

func GetStringMapStringSlice(c *viper.Viper, key string, fallback map[string][]string) func() map[string][]string {
	c = ensure(c)
	c.SetDefault(key, fallback)
	return func() map[string][]string {
		return c.GetStringMapStringSlice(key)
	}
}

func MustGetStringMapStringSlice(c *viper.Viper, key string) func() map[string][]string {
	c = ensure(c)
	must(c, key)
	return func() map[string][]string {
		return c.GetStringMapStringSlice(key)
	}
}

func GetStruct(c *viper.Viper, key string, fallback interface{}) (func(v interface{}) error, error) {
	c = ensure(c)

	// decode default
	var decoded map[string]interface{}
	if err := decode(fallback, &decoded); err != nil {
		return nil, err
	}

	// prefix key
	configMap := make(map[string]interface{}, len(decoded))
	for s, i := range decoded {
		configMap[key+"."+s] = i
	}

	if err := c.MergeConfigMap(configMap); err != nil {
		return nil, err
	}

	return func(v interface{}) error {
		var cfg map[string]interface{}
		if err := c.Unmarshal(&cfg); err != nil {
			return err
		}
		for _, keyPart := range strings.Split(key, ".") {
			if cfgPart, ok := cfg[keyPart]; ok {
				if o, ok := cfgPart.(map[string]interface{}); ok {
					cfg = o
				}
			}
		}
		return decode(cfg, v)
	}, nil
}

func ensure(c *viper.Viper) *viper.Viper {
	if c == nil {
		c = config
	}
	return c
}

func must(c *viper.Viper, key string) {
	if !c.IsSet(key) {
		panic(fmt.Sprintf("missing required config key: %s", key))
	}
}

func decode(input, output interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "yaml",
		Result:  output,
	})
	if err != nil {
		return err
	}
	return decoder.Decode(input)
}
