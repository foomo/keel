package keel_test

import (
	"context"
	"testing"

	"github.com/foomo/keel"
)

type benchCloser struct{}

func (benchCloser) Close() {}

type benchErrorCloser struct{}

func (benchErrorCloser) Close() error { return nil }

type benchCloserWithContext struct{}

func (benchCloserWithContext) Close(context.Context) {}

type benchNonCloser struct{}

func BenchmarkIsCloser(b *testing.B) {
	b.Run("first-case", func(b *testing.B) {
		v := benchCloser{}

		for b.Loop() {
			keel.IsCloser(v)
		}
	})

	b.Run("error-closer", func(b *testing.B) {
		v := benchErrorCloser{}

		for b.Loop() {
			keel.IsCloser(v)
		}
	})

	b.Run("context-closer", func(b *testing.B) {
		v := benchCloserWithContext{}

		for b.Loop() {
			keel.IsCloser(v)
		}
	})

	b.Run("non-match", func(b *testing.B) {
		v := benchNonCloser{}

		for b.Loop() {
			keel.IsCloser(v)
		}
	})
}
