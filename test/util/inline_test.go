package keeltestutil_test

import (
	"testing"

	keeltestutil "github.com/foomo/keel/test/util"
	"github.com/stretchr/testify/assert"
)

func TestInline(t *testing.T) {
	t.Run("read inline", func(t *testing.T) {
		value, ok := keeltestutil.Inline(t, 1) // INLINE: hello world
		assert.True(t, ok)
		assert.Equal(t, "hello world", value)
	})

	t.Run("read inline int", func(t *testing.T) {
		value, ok := keeltestutil.InlineInt(t, 1) // INLINE: 1
		assert.True(t, ok)
		assert.Equal(t, 1, value)
	})

	t.Run("read inline float", func(t *testing.T) {
		value, ok := keeltestutil.InlineFloat64(t, 1) // INLINE: 1.5
		assert.True(t, ok)
		assert.Equal(t, 1.5, value)
	})

	t.Run("read inline json", func(t *testing.T) {
		var x struct {
			Foo string `json:"foo"`
		}
		keeltestutil.InlineJSON(t, 1, &x) // INLINE: {"foo":"bar"}
		assert.Equal(t, "bar", x.Foo)
	})

	t.Run("write inline", func(t *testing.T) {
		value, ok := keeltestutil.Inline(t, 1, "hello %s", "world") // INLINE: hello world
		assert.True(t, ok)
		assert.Equal(t, "hello world", value)
	})
}
