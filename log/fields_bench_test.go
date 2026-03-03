package log_test

import (
	"testing"

	"github.com/foomo/keel/log"
)

func BenchmarkFJSON(b *testing.B) {
	b.Run("simple", func(b *testing.B) {
		v := map[string]string{"key": "value"}

		for b.Loop() {
			log.FJSON(v)
		}
	})

	b.Run("nested", func(b *testing.B) {
		v := map[string]any{
			"key":    "value",
			"number": 42,
			"nested": map[string]any{
				"inner": "data",
				"list":  []int{1, 2, 3},
			},
		}

		for b.Loop() {
			log.FJSON(v)
		}
	})
}

func BenchmarkFValue(b *testing.B) {
	b.Run("string", func(b *testing.B) {
		for b.Loop() {
			log.FValue("hello")
		}
	})

	b.Run("int", func(b *testing.B) {
		for b.Loop() {
			log.FValue(42)
		}
	})

	b.Run("struct", func(b *testing.B) {
		v := struct {
			Name string
			Age  int
		}{"Alice", 30}

		for b.Loop() {
			log.FValue(v)
		}
	})
}
