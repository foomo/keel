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
	config       *viper.Viper
	requiredKeys []string
	defaults     = map[string]interface{}{}
	types        = map[string]string{}
)

// Init sets up the configuration
func init() {
	config = viper.New()
	config.AutomaticEnv()
	config.SetConfigType("yaml")
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

// Config return the config instance
func Config() *viper.Viper {
	return config
}

func GetBool(c *viper.Viper, key string, fallback bool) func() bool {
	setDefault(c, key, "bool", fallback)

	return func() bool {
		return c.GetBool(key)
	}
}

func MustGetBool(c *viper.Viper, key string) func() bool {
	must(c, key, "bool")

	return func() bool {
		return c.GetBool(key)
	}
}

func GetInt(c *viper.Viper, key string, fallback int) func() int {
	setDefault(c, key, "int", fallback)

	return func() int {
		return c.GetInt(key)
	}
}

func MustGetInt(c *viper.Viper, key string) func() int {
	must(c, key, "int")

	return func() int {
		return c.GetInt(key)
	}
}

func GetInt32(c *viper.Viper, key string, fallback int32) func() int32 {
	setDefault(c, key, "int32", fallback)

	return func() int32 {
		return c.GetInt32(key)
	}
}

func MustGetInt32(c *viper.Viper, key string) func() int32 {
	must(c, key, "int32")

	return func() int32 {
		return c.GetInt32(key)
	}
}

func GetInt64(c *viper.Viper, key string, fallback int64) func() int64 {
	setDefault(c, key, "int64", fallback)

	return func() int64 {
		return c.GetInt64(key)
	}
}

func MustGetInt64(c *viper.Viper, key string) func() int64 {
	must(c, key, "int64")

	return func() int64 {
		return c.GetInt64(key)
	}
}

func GetUint(c *viper.Viper, key string, fallback uint) func() uint {
	setDefault(c, key, "uint", fallback)

	return func() uint {
		return c.GetUint(key)
	}
}

func MustGetUint(c *viper.Viper, key string) func() uint {
	must(c, key, "uint")

	return func() uint {
		return c.GetUint(key)
	}
}

func GetUint32(c *viper.Viper, key string, fallback uint32) func() uint32 {
	setDefault(c, key, "uint32", fallback)

	return func() uint32 {
		return c.GetUint32(key)
	}
}

func MustGetUint32(c *viper.Viper, key string) func() uint32 {
	must(c, key, "uint32")

	return func() uint32 {
		return c.GetUint32(key)
	}
}

func GetUint64(c *viper.Viper, key string, fallback uint64) func() uint64 {
	setDefault(c, key, "uint64", fallback)

	return func() uint64 {
		return c.GetUint64(key)
	}
}

func MustGetUint64(c *viper.Viper, key string) func() uint64 {
	must(c, key, "uint64")

	return func() uint64 {
		return c.GetUint64(key)
	}
}

func GetFloat64(c *viper.Viper, key string, fallback float64) func() float64 {
	setDefault(c, key, "float64", fallback)

	return func() float64 {
		return c.GetFloat64(key)
	}
}

func MustGetFloat64(c *viper.Viper, key string) func() float64 {
	must(c, key, "float64")

	return func() float64 {
		return c.GetFloat64(key)
	}
}

func GetString(c *viper.Viper, key, fallback string) func() string {
	setDefault(c, key, "string", fallback)

	return func() string {
		return c.GetString(key)
	}
}

func MustGetString(c *viper.Viper, key string) func() string {
	must(c, key, "string")

	return func() string {
		return c.GetString(key)
	}
}

func GetTime(c *viper.Viper, key string, fallback time.Time) func() time.Time {
	setDefault(c, key, "time.Time", fallback)

	return func() time.Time {
		return c.GetTime(key)
	}
}

func MustGetTime(c *viper.Viper, key string) func() time.Time {
	must(c, key, "time.Time")

	return func() time.Time {
		return c.GetTime(key)
	}
}

func GetDuration(c *viper.Viper, key string, fallback time.Duration) func() time.Duration {
	setDefault(c, key, "time.Duration", fallback)

	return func() time.Duration {
		return c.GetDuration(key)
	}
}

func MustGetDuration(c *viper.Viper, key string) func() time.Duration {
	must(c, key, "time.Duration")

	return func() time.Duration {
		return c.GetDuration(key)
	}
}

func GetIntSlice(c *viper.Viper, key string, fallback []int) func() []int {
	setDefault(c, key, "[]int", fallback)

	return func() []int {
		return c.GetIntSlice(key)
	}
}

func MustGetIntSlice(c *viper.Viper, key string) func() []int {
	must(c, key, "[]int")

	return func() []int {
		return c.GetIntSlice(key)
	}
}

func GetStringSlice(c *viper.Viper, key string, fallback []string) func() []string {
	setDefault(c, key, "[]string", fallback)

	return func() []string {
		return c.GetStringSlice(key)
	}
}

func MustGetStringSlice(c *viper.Viper, key string) func() []string {
	must(c, key, "[]string")

	return func() []string {
		return c.GetStringSlice(key)
	}
}

func GetStringMap(c *viper.Viper, key string, fallback map[string]interface{}) func() map[string]interface{} {
	setDefault(c, key, "map[string]interface{}", fallback)

	return func() map[string]interface{} {
		return c.GetStringMap(key)
	}
}

func MustGetStringMap(c *viper.Viper, key string) func() map[string]interface{} {
	must(c, key, "map[string]interface{}")

	return func() map[string]interface{} {
		return c.GetStringMap(key)
	}
}

func GetStringMapString(c *viper.Viper, key string, fallback map[string]string) func() map[string]string {
	setDefault(c, key, "map[string]string", fallback)

	return func() map[string]string {
		return c.GetStringMapString(key)
	}
}

func MustGetStringMapString(c *viper.Viper, key string) func() map[string]string {
	must(c, key, "map[string]string")

	return func() map[string]string {
		return c.GetStringMapString(key)
	}
}

func GetStringMapStringSlice(c *viper.Viper, key string, fallback map[string][]string) func() map[string][]string {
	setDefault(c, key, "map[string][]string", fallback)

	return func() map[string][]string {
		return c.GetStringMapStringSlice(key)
	}
}

func MustGetStringMapStringSlice(c *viper.Viper, key string) func() map[string][]string {
	must(c, key, "map[string][]string")

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

func RequiredKeys() []string {
	return requiredKeys
}

func Defaults() map[string]interface{} {
	return defaults
}

func Types() map[string]string {
	return types
}

func TypeOf(key string) string {
	if v, ok := types[key]; ok {
		return v
	}

	return ""
}

func ensure(c *viper.Viper) *viper.Viper {
	if c == nil {
		c = config
	}

	return c
}

func must(c *viper.Viper, key, typeof string) {
	c = ensure(c)
	types[key] = typeof

	requiredKeys = append(requiredKeys, key)
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

func setDefault(c *viper.Viper, key, typeof string, fallback any) {
	c = ensure(c)
	c.SetDefault(key, fallback)
	defaults[key] = fallback
	types[key] = typeof
}
