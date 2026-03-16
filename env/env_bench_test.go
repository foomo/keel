package env_test

import (
	"testing"

	"github.com/foomo/keel/env"
)

func BenchmarkGet(b *testing.B) {
	b.Setenv("BENCH_ENV_STRING", "value")

	for b.Loop() {
		env.Get("BENCH_ENV_STRING", "fallback")
	}
}

func BenchmarkGetInt(b *testing.B) {
	b.Setenv("BENCH_ENV_INT", "42")

	for b.Loop() {
		env.GetInt("BENCH_ENV_INT", 0)
	}
}

func BenchmarkGetStringSlice(b *testing.B) {
	b.Setenv("BENCH_ENV_SLICE", "a,b,c,d,e")

	for b.Loop() {
		env.GetStringSlice("BENCH_ENV_SLICE", nil)
	}
}

func BenchmarkGetParallel(b *testing.B) {
	b.Setenv("BENCH_ENV_PARALLEL", "value")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			env.Get("BENCH_ENV_PARALLEL", "fallback")
		}
	})
}
