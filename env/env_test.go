package env_test

import (
	"os"
	"testing"

	"github.com/foomo/keel/env"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	_ = os.Setenv("TEST_ENV_INT", "3")
	_ = os.Setenv("TEST_ENV_STRING", "test")

	m.Run()
}

func TextExists(t *testing.T) {
	t.Parallel()
	assert.True(t, env.Exists("TEST_ENV_EXISTS"))
	assert.False(t, env.Exists("TEST_ENV_NOOP"))
}

func TestMustExists(t *testing.T) {
	t.Parallel()
	assert.True(t, env.Exists("TEST_ENV_STRING"))
	assert.Panics(t, func() {
		env.MustExists("TEST_ENV_NOOP")
	})
}

func TestGet(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "test", env.Get("TEST_ENV_STRING", "fallback"))
	assert.Equal(t, "fallback", env.Get("TEST_ENV_NOOP", "fallback"))
}

func TestMustGet(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "test", env.Get("TEST_ENV_STRING", "fallback"))
	assert.Panics(t, func() {
		env.MustGet("TEST_ENV_NOOP")
	})
}

func TestGetInt(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 3, env.GetInt("TEST_ENV_INT", 4))
	assert.Equal(t, 4, env.GetInt("TEST_ENV_NOOP", 4))
}

func TestMustGetInt(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 3, env.GetInt("TEST_ENV_INT", 4))
	assert.Panics(t, func() {
		env.MustGet("TEST_ENV_NOOP")
	})
}

func TestRequiredKeys(t *testing.T) {
	t.Parallel()
	env.MustExists("TEST_ENV_STRING")
	assert.Contains(t, env.RequiredKeys(), "TEST_ENV_STRING")
}

func TestDefaults(t *testing.T) {
	t.Parallel()
	env.Get("TEST_ENV_STRING", "test")
	assert.Contains(t, env.Defaults(), "TEST_ENV_STRING")
}

func TestTypes(t *testing.T) {
	t.Parallel()
	env.Get("TEST_ENV_STRING", "test")
	env.GetInt("TEST_ENV_INT", 3)
	assert.NotEmpty(t, env.Types())
}

func TestTypeOf(t *testing.T) {
	t.Parallel()
	env.Get("TEST_ENV_STRING", "test")
	env.GetInt("TEST_ENV_INT", 3)
	assert.Equal(t, "string", env.TypeOf("TEST_ENV_STRING"))
	assert.Equal(t, "int", env.TypeOf("TEST_ENV_INT"))
}
