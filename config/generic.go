package config

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Supported is a type constraint for all config value types supported by viper.
type Supported interface {
	bool | int | int32 | int64 | uint | uint32 | uint64 |
		float64 | string | time.Time | time.Duration |
		[]int | []string |
		map[string]any | map[string]string | map[string][]string
}

// Get registers a config key with a fallback default and returns a getter closure.
func Get[T Supported](c *viper.Viper, key string, fallback T) func() T {
	setDefault(c, key, fmt.Sprintf("%T", fallback), fallback)

	return func() T {
		return getTyped[T](c, key)
	}
}

// MustGet registers a required config key and returns a getter closure.
// Panics if the key is not set.
func MustGet[T Supported](c *viper.Viper, key string) func() T {
	var zero T
	must(c, key, fmt.Sprintf("%T", zero))

	return func() T {
		return getTyped[T](c, key)
	}
}

// Watch polls the getter and calls the callback when the value changes.
func Watch[T comparable](ctx context.Context, fn func() T, callback func(T)) {
	current := fn()

	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			callback(current)
		}
	})
}

// WatchChan polls the getter and sends on ch when the value changes.
func WatchChan[T comparable](ctx context.Context, fn func() T, ch chan T) {
	current := fn()

	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			ch <- current
		}
	})
}

//nolint:forcetypeassert
func getTyped[T Supported](c *viper.Viper, key string) T {
	c = ensure(c)

	var zero T
	switch any(zero).(type) {
	case bool:
		return any(c.GetBool(key)).(T)
	case int:
		return any(c.GetInt(key)).(T)
	case int32:
		return any(c.GetInt32(key)).(T)
	case int64:
		return any(c.GetInt64(key)).(T)
	case uint:
		return any(c.GetUint(key)).(T)
	case uint32:
		return any(c.GetUint32(key)).(T)
	case uint64:
		return any(c.GetUint64(key)).(T)
	case float64:
		return any(c.GetFloat64(key)).(T)
	case string:
		return any(c.GetString(key)).(T)
	case time.Time:
		return any(c.GetTime(key)).(T)
	case time.Duration:
		return any(c.GetDuration(key)).(T)
	case []int:
		return any(c.GetIntSlice(key)).(T)
	case []string:
		return any(c.GetStringSlice(key)).(T)
	case map[string]any:
		return any(c.GetStringMap(key)).(T)
	case map[string]string:
		return any(c.GetStringMapString(key)).(T)
	case map[string][]string:
		return any(c.GetStringMapStringSlice(key)).(T)
	default:
		panic(fmt.Sprintf("unsupported config type: %T", zero))
	}
}
