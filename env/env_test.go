package env_test

import (
	"testing"

	"github.com/foomo/keel/env"
	"github.com/stretchr/testify/assert"
)

func TestExists(t *testing.T) {
	t.Setenv("TEST_ENV_STRING", "test")

	assert.True(t, env.Exists("TEST_ENV_STRING"))
	assert.False(t, env.Exists("TEST_ENV_NOOP"))
}

func TestMustExists(t *testing.T) {
	t.Setenv("TEST_ENV_STRING", "test")

	assert.NotPanics(t, func() {
		env.MustExists("TEST_ENV_STRING")
	})
	assert.Panics(t, func() {
		env.MustExists("TEST_ENV_NOOP")
	})
}

func TestGet(t *testing.T) {
	t.Setenv("TEST_ENV_STRING", "test")

	assert.Equal(t, "test", env.Get("TEST_ENV_STRING", "fallback"))
	assert.Equal(t, "fallback", env.Get("TEST_ENV_NOOP", "fallback"))
}

func TestMustGet(t *testing.T) {
	t.Setenv("TEST_ENV_STRING", "test")

	assert.Equal(t, "test", env.Get("TEST_ENV_STRING", "fallback"))
	assert.Panics(t, func() {
		env.MustGet("TEST_ENV_NOOP")
	})
}

func TestGetInt(t *testing.T) {
	t.Setenv("TEST_ENV_INT", "3")

	assert.Equal(t, 3, env.GetInt("TEST_ENV_INT", 4))
	assert.Equal(t, 4, env.GetInt("TEST_ENV_NOOP", 4))
}

func TestMustGetInt(t *testing.T) {
	t.Setenv("TEST_ENV_INT", "3")

	assert.Equal(t, 3, env.GetInt("TEST_ENV_INT", 4))
	assert.Panics(t, func() {
		env.MustGet("TEST_ENV_NOOP")
	})
}

func TestRequiredKeys(t *testing.T) {
	t.Setenv("TEST_ENV_STRING", "test")

	env.MustExists("TEST_ENV_STRING")
	assert.Contains(t, env.RequiredKeys(), "TEST_ENV_STRING")
}

func TestDefaults(t *testing.T) {
	t.Setenv("TEST_ENV_STRING", "test")

	env.Get("TEST_ENV_STRING", "test")
	assert.Contains(t, env.Defaults(), "TEST_ENV_STRING")
}

func TestTypes(t *testing.T) {
	t.Setenv("TEST_ENV_INT", "3")
	t.Setenv("TEST_ENV_STRING", "test")

	env.Get("TEST_ENV_STRING", "test")
	env.GetInt("TEST_ENV_INT", 3)
	assert.NotEmpty(t, env.Types())
}

func TestTypeOf(t *testing.T) {
	t.Setenv("TEST_ENV_INT", "3")
	t.Setenv("TEST_ENV_STRING", "test")

	env.Get("TEST_ENV_STRING", "test")
	env.GetInt("TEST_ENV_INT", 3)
	assert.Equal(t, "string", env.TypeOf("TEST_ENV_STRING"))
	assert.Equal(t, "int", env.TypeOf("TEST_ENV_INT"))
}
