//go:build safe

package config_test

import (
	"testing"
	"time"

	"github.com/foomo/keel/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	t.Parallel()

	c := viper.New()

	t.Run("string", func(t *testing.T) {
		t.Parallel()

		fn := config.Get[string](c, "test.string", "fallback")
		assert.Equal(t, "fallback", fn())

		c.Set("test.string", "override")
		assert.Equal(t, "override", fn())
	})

	t.Run("int", func(t *testing.T) {
		t.Parallel()

		fn := config.Get[int](c, "test.int", 42)
		assert.Equal(t, 42, fn())

		c.Set("test.int", 99)
		assert.Equal(t, 99, fn())
	})

	t.Run("bool", func(t *testing.T) {
		t.Parallel()

		fn := config.Get[bool](c, "test.bool", true)
		assert.True(t, fn())

		c.Set("test.bool", false)
		assert.False(t, fn())
	})

	t.Run("duration", func(t *testing.T) {
		t.Parallel()

		fn := config.Get[time.Duration](c, "test.duration", 5*time.Second)
		assert.Equal(t, 5*time.Second, fn())

		c.Set("test.duration", 10*time.Second)
		assert.Equal(t, 10*time.Second, fn())
	})

	t.Run("string slice", func(t *testing.T) {
		t.Parallel()

		fn := config.Get[[]string](c, "test.strings", []string{"a", "b"})
		assert.Equal(t, []string{"a", "b"}, fn())

		c.Set("test.strings", []string{"c"})
		assert.Equal(t, []string{"c"}, fn())
	})
}

func TestMustGet(t *testing.T) {
	t.Parallel()

	t.Run("panics when missing", func(t *testing.T) {
		t.Parallel()

		c := viper.New()

		assert.Panics(t, func() {
			config.MustGet[string](c, "missing.key")
		})
	})

	t.Run("returns value when set", func(t *testing.T) {
		t.Parallel()

		c := viper.New()
		c.Set("present.key", "hello")
		fn := config.MustGet[string](c, "present.key")
		assert.Equal(t, "hello", fn())
	})
}

func TestWatch(t *testing.T) {
	t.Parallel()

	c := viper.New()
	fn := config.Get[string](c, "watch.test", "initial")

	ctx := t.Context()

	ch := make(chan string, 1)

	config.Watch(ctx, fn, func(v string) {
		ch <- v
	})

	c.Set("watch.test", "changed")

	select {
	case v := <-ch:
		assert.Equal(t, "changed", v)
	case <-time.After(3 * time.Second):
		require.Fail(t, "watch callback not called within timeout")
	}
}

func TestWatchChan(t *testing.T) {
	t.Parallel()

	c := viper.New()
	fn := config.Get[int](c, "watchch.test", 1)

	ctx := t.Context()

	ch := make(chan int, 1)
	config.WatchChan(ctx, fn, ch)

	c.Set("watchch.test", 2)

	select {
	case v := <-ch:
		assert.Equal(t, 2, v)
	case <-time.After(3 * time.Second):
		require.Fail(t, "watch chan not notified within timeout")
	}
}
