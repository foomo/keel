package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

var (
	types        = sync.Map{}
	defaults     = sync.Map{}
	requiredKeys = sync.Map{}
)

// Exists return true if env var is defined
func Exists(key string) bool {
	_, ok := os.LookupEnv(key)
	return ok
}

// MustExists panics if not exists
func MustExists(key string) {
	if !Exists(key) {
		panic(fmt.Sprintf("required environment variable `%s` does not exist", key))
	}

	if _, ok := requiredKeys.Load(key); !ok {
		requiredKeys.Store(key, true)
	}
}

// Get env var or fallback
func Get(key, fallback string) string {
	defaults.Store(key, fallback)

	if _, ok := types.Load(key); !ok {
		types.Store(key, "string")
	}

	if v, ok := os.LookupEnv(key); ok {
		return v
	}

	return fallback
}

// MustGet env var or panic
func MustGet(key string) string {
	MustExists(key)
	return Get(key, "")
}

// GetInt env var or fallback as int
func GetInt(key string, fallback int) int {
	if _, ok := types.Load(key); !ok {
		types.Store(key, "int")
	}

	if value, err := strconv.Atoi(Get(key, "")); err == nil {
		return value
	}

	return fallback
}

// MustGetInt env var as int or panic
func MustGetInt(key string) int {
	MustExists(key)
	return GetInt(key, 0)
}

// GetInt64 env var or fallback as int64
func GetInt64(key string, fallback int64) int64 {
	if _, ok := types.Load(key); !ok {
		types.Store(key, "int64")
	}

	if value, err := strconv.ParseInt(Get(key, ""), 10, 64); err == nil {
		return value
	}

	return fallback
}

// MustGetInt64 env var as int64 or panic
func MustGetInt64(key string) int64 {
	MustExists(key)
	return GetInt64(key, 0)
}

// GetFloat64 env var or fallback as float64
func GetFloat64(key string, fallback float64) float64 {
	if _, ok := types.Load(key); !ok {
		types.Store(key, "float64")
	}

	if value, err := strconv.ParseFloat(Get(key, ""), 64); err == nil {
		return value
	}

	return fallback
}

// MustGetFloat64 env var as float64 or panic
func MustGetFloat64(key string) float64 {
	MustExists(key)
	return GetFloat64(key, 0)
}

// GetBool env var or fallback as bool
func GetBool(key string, fallback bool) bool {
	if _, ok := types.Load(key); !ok {
		types.Store(key, "bool")
	}

	if val, err := strconv.ParseBool(Get(key, "")); err == nil {
		return val
	}

	return fallback
}

// MustGetBool env var as bool or panic
func MustGetBool(key string) bool {
	MustExists(key)
	return GetBool(key, false)
}

// GetStringSlice env var or fallback as []string
func GetStringSlice(key string, fallback []string) []string {
	if _, ok := types.Load(key); !ok {
		types.Store(key, "[]string")
	}

	if v := Get(key, ""); v != "" {
		return strings.Split(v, ",")
	}

	return fallback
}

// MustGetStringSlice env var as bool or panic
func MustGetStringSlice(key string) []string {
	MustExists(key)
	return GetStringSlice(key, nil)
}

// GetIntSlice env var or fallback as []string
func GetIntSlice(key string, fallback []int) []int {
	if _, ok := types.Load(key); !ok {
		types.Store(key, "[]int")
	}

	if v := Get(key, ""); v != "" {
		elements := strings.Split(v, ",")

		ret := make([]int, len(elements))
		for i, stringVal := range elements {
			intVal, err := strconv.Atoi(stringVal)
			if err != nil {
				return fallback
			}

			ret[i] = intVal
		}
	}

	return fallback
}

// MustGetGetIntSlice env var as bool or panic
func MustGetGetIntSlice(key string) []int {
	MustExists(key)
	return GetIntSlice(key, nil)
}

func RequiredKeys() []string {
	var ret []string

	requiredKeys.Range(func(key, value interface{}) bool {
		if v, ok := key.(string); ok {
			ret = append(ret, v)
		}

		return true
	})

	return ret
}

func Defaults() map[string]interface{} {
	ret := map[string]interface{}{}

	defaults.Range(func(key, value interface{}) bool {
		if k, ok := key.(string); ok {
			ret[k] = value
		}

		return true
	})

	return ret
}

func Types() map[string]string {
	ret := map[string]string{}

	types.Range(func(key, value interface{}) bool {
		if v, ok := value.(string); ok {
			if k, ok := key.(string); ok {
				ret[k] = v
			}
		}

		return true
	})

	return ret
}

func TypeOf(key string) string {
	if v, ok := types.Load(key); ok {
		if s, ok := v.(string); ok {
			return s
		}

		return ""
	}

	return ""
}
