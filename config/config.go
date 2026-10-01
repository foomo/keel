package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Package-level registry state. config is the global [viper.Viper] instance
// used whenever a nil instance is passed; requiredKeys, defaults and types
// record every key registered through the getter constructors.
var (
	config       *viper.Viper
	requiredKeys []string
	defaults     = map[string]any{}
	types        = map[string]string{}
)

// init creates the global viper instance with automatic environment lookup,
// YAML config type and "." to "_" env key replacement.
func init() {
	config = viper.New()
	config.AutomaticEnv()
	config.SetConfigType("yaml")
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

// Config returns the global [viper.Viper] instance.
func Config() *viper.Viper {
	return config
}

// GetBool registers key with fallback as its default on c and returns a
// getter that reads the current bool value on each call.
func GetBool(c *viper.Viper, key string, fallback bool) func() bool {
	setDefault(c, key, "bool", fallback)

	return func() bool {
		return c.GetBool(key)
	}
}

// MustGetBool registers key as required on c and returns a getter that reads
// the current bool value on each call. It panics if key is not set.
func MustGetBool(c *viper.Viper, key string) func() bool {
	must(c, key, "bool")

	return func() bool {
		return c.GetBool(key)
	}
}

// GetInt registers key with fallback as its default on c and returns a
// getter that reads the current int value on each call.
func GetInt(c *viper.Viper, key string, fallback int) func() int {
	setDefault(c, key, "int", fallback)

	return func() int {
		return c.GetInt(key)
	}
}

// MustGetInt registers key as required on c and returns a getter that reads
// the current int value on each call. It panics if key is not set.
func MustGetInt(c *viper.Viper, key string) func() int {
	must(c, key, "int")

	return func() int {
		return c.GetInt(key)
	}
}

// GetInt32 registers key with fallback as its default on c and returns a
// getter that reads the current int32 value on each call.
func GetInt32(c *viper.Viper, key string, fallback int32) func() int32 {
	setDefault(c, key, "int32", fallback)

	return func() int32 {
		return c.GetInt32(key)
	}
}

// MustGetInt32 registers key as required on c and returns a getter that reads
// the current int32 value on each call. It panics if key is not set.
func MustGetInt32(c *viper.Viper, key string) func() int32 {
	must(c, key, "int32")

	return func() int32 {
		return c.GetInt32(key)
	}
}

// GetInt64 registers key with fallback as its default on c and returns a
// getter that reads the current int64 value on each call.
func GetInt64(c *viper.Viper, key string, fallback int64) func() int64 {
	setDefault(c, key, "int64", fallback)

	return func() int64 {
		return c.GetInt64(key)
	}
}

// MustGetInt64 registers key as required on c and returns a getter that reads
// the current int64 value on each call. It panics if key is not set.
func MustGetInt64(c *viper.Viper, key string) func() int64 {
	must(c, key, "int64")

	return func() int64 {
		return c.GetInt64(key)
	}
}

// GetUint registers key with fallback as its default on c and returns a
// getter that reads the current uint value on each call.
func GetUint(c *viper.Viper, key string, fallback uint) func() uint {
	setDefault(c, key, "uint", fallback)

	return func() uint {
		return c.GetUint(key)
	}
}

// MustGetUint registers key as required on c and returns a getter that reads
// the current uint value on each call. It panics if key is not set.
func MustGetUint(c *viper.Viper, key string) func() uint {
	must(c, key, "uint")

	return func() uint {
		return c.GetUint(key)
	}
}

// GetUint32 registers key with fallback as its default on c and returns a
// getter that reads the current uint32 value on each call.
func GetUint32(c *viper.Viper, key string, fallback uint32) func() uint32 {
	setDefault(c, key, "uint32", fallback)

	return func() uint32 {
		return c.GetUint32(key)
	}
}

// MustGetUint32 registers key as required on c and returns a getter that reads
// the current uint32 value on each call. It panics if key is not set.
func MustGetUint32(c *viper.Viper, key string) func() uint32 {
	must(c, key, "uint32")

	return func() uint32 {
		return c.GetUint32(key)
	}
}

// GetUint64 registers key with fallback as its default on c and returns a
// getter that reads the current uint64 value on each call.
func GetUint64(c *viper.Viper, key string, fallback uint64) func() uint64 {
	setDefault(c, key, "uint64", fallback)

	return func() uint64 {
		return c.GetUint64(key)
	}
}

// MustGetUint64 registers key as required on c and returns a getter that reads
// the current uint64 value on each call. It panics if key is not set.
func MustGetUint64(c *viper.Viper, key string) func() uint64 {
	must(c, key, "uint64")

	return func() uint64 {
		return c.GetUint64(key)
	}
}

// GetFloat64 registers key with fallback as its default on c and returns a
// getter that reads the current float64 value on each call.
func GetFloat64(c *viper.Viper, key string, fallback float64) func() float64 {
	setDefault(c, key, "float64", fallback)

	return func() float64 {
		return c.GetFloat64(key)
	}
}

// MustGetFloat64 registers key as required on c and returns a getter that reads
// the current float64 value on each call. It panics if key is not set.
func MustGetFloat64(c *viper.Viper, key string) func() float64 {
	must(c, key, "float64")

	return func() float64 {
		return c.GetFloat64(key)
	}
}

// GetString registers key with fallback as its default on c and returns a
// getter that reads the current string value on each call.
func GetString(c *viper.Viper, key, fallback string) func() string {
	setDefault(c, key, "string", fallback)

	return func() string {
		return c.GetString(key)
	}
}

// MustGetString registers key as required on c and returns a getter that reads
// the current string value on each call. It panics if key is not set.
func MustGetString(c *viper.Viper, key string) func() string {
	must(c, key, "string")

	return func() string {
		return c.GetString(key)
	}
}

// GetTime registers key with fallback as its default on c and returns a
// getter that reads the current [time.Time] value on each call.
func GetTime(c *viper.Viper, key string, fallback time.Time) func() time.Time {
	setDefault(c, key, "time.Time", fallback)

	return func() time.Time {
		return c.GetTime(key)
	}
}

// MustGetTime registers key as required on c and returns a getter that reads
// the current [time.Time] value on each call. It panics if key is not set.
func MustGetTime(c *viper.Viper, key string) func() time.Time {
	must(c, key, "time.Time")

	return func() time.Time {
		return c.GetTime(key)
	}
}

// GetDuration registers key with fallback as its default on c and returns a
// getter that reads the current [time.Duration] value on each call.
func GetDuration(c *viper.Viper, key string, fallback time.Duration) func() time.Duration {
	setDefault(c, key, "time.Duration", fallback)

	return func() time.Duration {
		return c.GetDuration(key)
	}
}

// MustGetDuration registers key as required on c and returns a getter that reads
// the current [time.Duration] value on each call. It panics if key is not set.
func MustGetDuration(c *viper.Viper, key string) func() time.Duration {
	must(c, key, "time.Duration")

	return func() time.Duration {
		return c.GetDuration(key)
	}
}

// GetIntSlice registers key with fallback as its default on c and returns a
// getter that reads the current []int value on each call.
func GetIntSlice(c *viper.Viper, key string, fallback []int) func() []int {
	setDefault(c, key, "[]int", fallback)

	return func() []int {
		return c.GetIntSlice(key)
	}
}

// MustGetIntSlice registers key as required on c and returns a getter that reads
// the current []int value on each call. It panics if key is not set.
func MustGetIntSlice(c *viper.Viper, key string) func() []int {
	must(c, key, "[]int")

	return func() []int {
		return c.GetIntSlice(key)
	}
}

// GetStringSlice registers key with fallback as its default on c and returns a
// getter that reads the current []string value on each call.
func GetStringSlice(c *viper.Viper, key string, fallback []string) func() []string {
	setDefault(c, key, "[]string", fallback)

	return func() []string {
		return c.GetStringSlice(key)
	}
}

// MustGetStringSlice registers key as required on c and returns a getter that reads
// the current []string value on each call. It panics if key is not set.
func MustGetStringSlice(c *viper.Viper, key string) func() []string {
	must(c, key, "[]string")

	return func() []string {
		return c.GetStringSlice(key)
	}
}

// GetStringMap registers key with fallback as its default on c and returns a
// getter that reads the current map[string]any value on each call.
func GetStringMap(c *viper.Viper, key string, fallback map[string]any) func() map[string]any {
	setDefault(c, key, "map[string]interface{}", fallback)

	return func() map[string]any {
		return c.GetStringMap(key)
	}
}

// MustGetStringMap registers key as required on c and returns a getter that reads
// the current map[string]any value on each call. It panics if key is not set.
func MustGetStringMap(c *viper.Viper, key string) func() map[string]any {
	must(c, key, "map[string]interface{}")

	return func() map[string]any {
		return c.GetStringMap(key)
	}
}

// GetStringMapString registers key with fallback as its default on c and returns a
// getter that reads the current map[string]string value on each call.
func GetStringMapString(c *viper.Viper, key string, fallback map[string]string) func() map[string]string {
	setDefault(c, key, "map[string]string", fallback)

	return func() map[string]string {
		return c.GetStringMapString(key)
	}
}

// MustGetStringMapString registers key as required on c and returns a getter that reads
// the current map[string]string value on each call. It panics if key is not set.
func MustGetStringMapString(c *viper.Viper, key string) func() map[string]string {
	must(c, key, "map[string]string")

	return func() map[string]string {
		return c.GetStringMapString(key)
	}
}

// GetStringMapStringSlice registers key with fallback as its default on c and returns a
// getter that reads the current map[string][]string value on each call.
func GetStringMapStringSlice(c *viper.Viper, key string, fallback map[string][]string) func() map[string][]string {
	setDefault(c, key, "map[string][]string", fallback)

	return func() map[string][]string {
		return c.GetStringMapStringSlice(key)
	}
}

// MustGetStringMapStringSlice registers key as required on c and returns a getter that reads
// the current map[string][]string value on each call. It panics if key is not set.
func MustGetStringMapStringSlice(c *viper.Viper, key string) func() map[string][]string {
	must(c, key, "map[string][]string")

	return func() map[string][]string {
		return c.GetStringMapStringSlice(key)
	}
}

// GetStruct decodes fallback (using "yaml" struct tags) into defaults below
// key, merges them into c and returns a function that decodes the current
// value at key into v. A nil c selects the global instance. It returns an
// error if fallback cannot be decoded or merged.
func GetStruct(c *viper.Viper, key string, fallback any) (func(v any) error, error) {
	c = ensure(c)

	// decode default
	var decoded map[string]any
	if err := decode(fallback, &decoded); err != nil {
		return nil, err
	}

	// prefix key
	configMap := make(map[string]any, len(decoded))
	for s, i := range decoded {
		configMap[key+"."+s] = i
	}

	if err := c.MergeConfigMap(configMap); err != nil {
		return nil, err
	}

	return func(v any) error {
		var cfg map[string]any
		if err := c.Unmarshal(&cfg); err != nil {
			return err
		}

		for keyPart := range strings.SplitSeq(key, ".") {
			if cfgPart, ok := cfg[keyPart]; ok {
				if o, ok := cfgPart.(map[string]any); ok {
					cfg = o
				}
			}
		}

		return decode(cfg, v)
	}, nil
}

// RequiredKeys returns all keys registered as required.
func RequiredKeys() []string {
	return requiredKeys
}

// Defaults returns the registered default values by key.
func Defaults() map[string]any {
	return defaults
}

// Types returns the registered type names by key.
func Types() map[string]string {
	return types
}

// TypeOf returns the registered type name of key, or "" if key is unknown.
func TypeOf(key string) string {
	if v, ok := types[key]; ok {
		return v
	}

	return ""
}

// ensure returns c, or the global instance if c is nil.
func ensure(c *viper.Viper) *viper.Viper {
	if c == nil {
		c = config
	}

	return c
}

// must records key as required with type typeof and panics if it is not set.
func must(c *viper.Viper, key, typeof string) {
	c = ensure(c)
	types[key] = typeof

	requiredKeys = append(requiredKeys, key)
	if !c.IsSet(key) {
		panic(fmt.Sprintf("missing required config key: %s", key))
	}
}

// decode converts input into output using mapstructure with "yaml" tags.
func decode(input, output any) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "yaml",
		Result:  output,
	})
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}

// setDefault sets fallback as the default for key and records its type.
func setDefault(c *viper.Viper, key, typeof string, fallback any) {
	c = ensure(c)
	c.SetDefault(key, fallback)
	defaults[key] = fallback
	types[key] = typeof
}
