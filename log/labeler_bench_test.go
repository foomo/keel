package log_test

import (
	"testing"

	"github.com/foomo/keel/log"
	"go.uber.org/zap"
)

func BenchmarkLabelerAdd(b *testing.B) {
	l := &log.Labeler{}
	field := zap.String("key", "value")

	for b.Loop() {
		l.Add(field)
	}
}

func BenchmarkLabelerGet(b *testing.B) {
	l := &log.Labeler{}
	for range 10 {
		l.Add(zap.String("key", "value"))
	}

	for b.Loop() {
		l.Get()
	}
}

func BenchmarkLabelerAddParallel(b *testing.B) {
	l := &log.Labeler{}
	field := zap.String("key", "value")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Add(field)
		}
	})
}
