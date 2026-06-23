package keel_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/foomo/keel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// recordCloser implements interfaces.ErrorCloserWithContext and records calls.
type recordCloser struct {
	closed atomic.Bool
}

func (c *recordCloser) Close(_ context.Context) error {
	c.closed.Store(true)
	return nil
}

func TestJob_StepsRunInOrder(t *testing.T) {
	t.Parallel()

	var order []string

	j := keel.NewJob(keel.JobWithLogger(zaptest.NewLogger(t)))
	for _, name := range []string{"a", "b", "c"} {
		j.AddStep(name, func(_ context.Context, _ *zap.Logger) error {
			order = append(order, name)
			return nil
		})
	}

	require.NoError(t, j.RunE())
	assert.Equal(t, []string{"a", "b", "c"}, order)
}

func TestJob_StepErrorStopsChain(t *testing.T) {
	t.Parallel()

	boom := errors.New("boom")

	var ran []string

	j := keel.NewJob(keel.JobWithLogger(zaptest.NewLogger(t)))
	j.AddStep("first", func(_ context.Context, _ *zap.Logger) error {
		ran = append(ran, "first")
		return nil
	})
	j.AddStep("second", func(_ context.Context, _ *zap.Logger) error {
		ran = append(ran, "second")
		return boom
	})
	j.AddStep("third", func(_ context.Context, _ *zap.Logger) error {
		ran = append(ran, "third")
		return nil
	})

	err := j.RunE()
	require.ErrorIs(t, err, boom)
	assert.Equal(t, []string{"first", "second"}, ran, "third step must not run")
}

func TestJob_TimeoutCancelsStep(t *testing.T) {
	t.Parallel()

	j := keel.NewJob(
		keel.JobWithLogger(zaptest.NewLogger(t)),
		keel.JobWithTimeout(50*time.Millisecond),
	)
	j.AddStep("slow", func(ctx context.Context, _ *zap.Logger) error {
		<-ctx.Done()
		return ctx.Err()
	})

	err := j.RunE()
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestJob_ContextCancelInterrupts(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	j := keel.NewJob(
		keel.JobWithLogger(zaptest.NewLogger(t)),
		keel.JobWithContext(ctx),
	)
	j.AddStep("blocking", func(stepCtx context.Context, _ *zap.Logger) error {
		<-stepCtx.Done()
		return stepCtx.Err()
	})

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := j.RunE()
	require.ErrorIs(t, err, context.Canceled)
}

func TestJob_FinalizeRunsClosersAndPushersOnError(t *testing.T) {
	t.Parallel()

	closer := &recordCloser{}

	var pushed atomic.Bool

	j := keel.NewJob(
		keel.JobWithLogger(zaptest.NewLogger(t)),
		keel.JobWithCloser(closer),
	)
	j.AddPusherForTest(func(_ context.Context) error {
		pushed.Store(true)
		return nil
	})
	j.AddStep("fail", func(_ context.Context, _ *zap.Logger) error {
		return errors.New("boom")
	})

	require.Error(t, j.RunE())
	assert.True(t, closer.closed.Load(), "closer must run during finalize even on error")
	assert.True(t, pushed.Load(), "pusher must run during finalize even on error")
}

func TestJob_NoStepsSucceeds(t *testing.T) {
	t.Parallel()

	j := keel.NewJob(keel.JobWithName("empty"), keel.JobWithLogger(zaptest.NewLogger(t)))
	require.NoError(t, j.RunE())
}

func TestJob_ParallelRunsConcurrently(t *testing.T) {
	t.Parallel()

	var (
		ran    atomic.Int32
		active atomic.Int32
		peak   atomic.Int32
	)

	j := keel.NewJob(
		keel.JobWithName("parallel"),
		keel.JobWithLogger(zaptest.NewLogger(t)),
		keel.JobWithParallel(0),
	)
	for _, name := range []string{"a", "b", "c"} {
		j.AddStep(name, func(_ context.Context, _ *zap.Logger) error {
			ran.Add(1)

			cur := active.Add(1)

			for {
				p := peak.Load()
				if cur <= p || peak.CompareAndSwap(p, cur) {
					break
				}
			}

			time.Sleep(30 * time.Millisecond)
			active.Add(-1)

			return nil
		})
	}

	require.NoError(t, j.RunE())
	assert.Equal(t, int32(3), ran.Load(), "all steps must run")
	assert.GreaterOrEqual(t, peak.Load(), int32(2), "steps must overlap (run concurrently)")
}

func TestJob_ParallelFailFastReturnsError(t *testing.T) {
	t.Parallel()

	boom := errors.New("boom")

	j := keel.NewJob(
		keel.JobWithName("parallel-fail"),
		keel.JobWithLogger(zaptest.NewLogger(t)),
		keel.JobWithParallel(2),
	)
	j.AddStep("ok", func(ctx context.Context, _ *zap.Logger) error {
		<-ctx.Done() // cancelled by fail-fast when the sibling errors
		return ctx.Err()
	})
	j.AddStep("fail", func(_ context.Context, _ *zap.Logger) error {
		return boom
	})

	require.ErrorIs(t, j.RunE(), boom)
}

func TestJob_PanicInStepRecovered(t *testing.T) {
	t.Parallel()

	closer := &recordCloser{}

	j := keel.NewJob(
		keel.JobWithName("panic"),
		keel.JobWithLogger(zaptest.NewLogger(t)),
		keel.JobWithCloser(closer),
	)
	j.AddStep("boom", func(_ context.Context, _ *zap.Logger) error {
		panic("kaboom")
	})

	err := j.RunE()
	require.Error(t, err, "panic must surface as an error, not crash")
	assert.True(t, closer.closed.Load(), "finalize must still run after a panic")
}

func TestJob_NameDefaultsAndOverride(t *testing.T) {
	t.Parallel()

	override := keel.NewJob(keel.JobWithName("custom"), keel.JobWithLogger(zaptest.NewLogger(t)))
	assert.Equal(t, "custom", override.NameForTest())

	def := keel.NewJob(keel.JobWithLogger(zaptest.NewLogger(t)))
	assert.NotEmpty(t, def.NameForTest(), "name defaults to OTEL_SERVICE_NAME / DefaultServiceName")
}
