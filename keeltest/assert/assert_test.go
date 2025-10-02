package keelassert_test

import (
	"testing"

	keelassert "github.com/foomo/keel/keeltest/assert"
	"github.com/stretchr/testify/assert"
)

func TestEqualInline(t *testing.T) {
	tests := []struct {
		name string
		when func(t *testing.T) bool
	}{
		{
			name: "equal int",
			when: func(t *testing.T) bool { //nolint:thelper
				return keelassert.InlineEqual(t, 15) // INLINE: 15
			},
		},
		{
			name: "equal bool",
			when: func(t *testing.T) bool { //nolint:thelper
				return keelassert.InlineEqual(t, true) // INLINE: true
			},
		},
		{
			name: "equal string",
			when: func(t *testing.T) bool { //nolint:thelper
				return keelassert.InlineEqual(t, "foo bar") // INLINE: foo bar
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, tt.when(t))
		})
	}
}

func TestEqualInlineJSONEq(t *testing.T) {
	tests := []struct {
		name string
		when func(t *testing.T) bool
	}{
		{
			name: "equal json",
			when: func(t *testing.T) bool { //nolint:thelper
				x := struct {
					Foo string `json:"foo"`
				}{
					Foo: "bar",
				}

				return keelassert.InlineJSONEq(t, x) // INLINE: {"foo":"bar"}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, tt.when(t))
		})
	}
}
