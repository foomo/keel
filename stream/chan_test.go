package stream_test

import (
	"cmp"
	"context"
	"errors"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/foomo/gofuncy"
	"github.com/foomo/keel/stream"
	"github.com/foomo/opentelemetry-go/exporters/glossy/glossytrace"
	oteltesting "github.com/foomo/opentelemetry-go/testing"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestEmpty(t *testing.T) {
	s := stream.Empty[int]()
	assert.Equal(t, 0, s.Count())
}

func TestEmptySource(t *testing.T) {
	s := stream.Empty[string]()
	assert.Empty(t, s.Collect())
}

func TestOf(t *testing.T) {
	data := []int{1, 2, 3, 4, 4, 22, 2, 1, 4}
	assert.Equal(t, data, stream.Of(t.Context(), data...).Collect())
}

func TestOfEmpty(t *testing.T) {
	s := stream.Of[int](t.Context())
	assert.Equal(t, 0, s.Count())
}

func TestOfCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	s := stream.Of(ctx, 1, 2, 3, 4, 5)
	assert.Equal(t, 0, s.Count())
}

func TestFrom(t *testing.T) {
	ch := make(chan int, 3)
	ch <- 10
	ch <- 20
	ch <- 30
	close(ch)

	assert.Equal(t, []int{10, 20, 30}, stream.From(ch).Collect())
}

func TestCollect(t *testing.T) {
	assert.Equal(t, []string{"a", "b"}, stream.Of(t.Context(), "a", "b").Collect())
	assert.Nil(t, stream.Empty[int]().Collect())
}

func TestMap(t *testing.T) {
	ctx := t.Context()
	got := stream.Map(ctx, stream.Of(ctx, 1, 2, 3), func(_ context.Context, n int) (string, error) {
		return strconv.Itoa(n), nil
	}).Collect()
	assert.Equal(t, []string{"1", "2", "3"}, got)
}

func TestMapSameType(t *testing.T) {
	ctx := t.Context()
	got := stream.Map(ctx, stream.Of(ctx, 1, 2, 3), func(_ context.Context, n int) (int, error) { return n * 2, nil }).Collect()
	assert.Equal(t, []int{2, 4, 6}, got)
}

func TestMapChain(t *testing.T) {
	ctx := t.Context()
	s := stream.Of(ctx, 1, 2, 3)
	doubled := stream.Map(ctx, s, func(_ context.Context, n int) (int, error) { return n * 2, nil })
	strs := stream.Map(ctx, doubled, func(_ context.Context, n int) (string, error) { return strconv.Itoa(n), nil })
	assert.Equal(t, []string{"2", "4", "6"}, strs.Collect())
}

func TestFilter(t *testing.T) {
	ctx := t.Context()
	got := stream.Of(ctx, 1, 2, 3, 4, 5, 6).Filter(ctx, func(_ context.Context, n int) bool {
		return n%2 == 0
	}).Collect()
	assert.Equal(t, []int{2, 4, 6}, got)
}

func TestFilterEmpty(t *testing.T) {
	got := stream.Empty[int]().Filter(t.Context(), func(_ context.Context, _ int) bool {
		return true
	}).Collect()
	assert.Nil(t, got)
}

func TestSort(t *testing.T) {
	ctx := t.Context()
	got := stream.Of(ctx, 3, 1, 4, 1, 5).Sort(ctx, cmp.Compare[int]).Collect()
	assert.Equal(t, []int{1, 1, 3, 4, 5}, got)
}

func TestReverse(t *testing.T) {
	ctx := t.Context()
	got := stream.Of(ctx, 1, 2, 3).Reverse(ctx).Collect()
	assert.Equal(t, []int{3, 2, 1}, got)
}

func TestReverseEmpty(t *testing.T) {
	got := stream.Empty[int]().Reverse(t.Context()).Collect()
	assert.Nil(t, got)
}

func TestSplit(t *testing.T) {
	ctx := t.Context()
	got := stream.Split(ctx, stream.Of(ctx, 1, 2, 3, 4, 5), 2).Collect()
	assert.Equal(t, [][]int{{1, 2}, {3, 4}, {5}}, got)
}

func TestSplitSingleBatch(t *testing.T) {
	ctx := t.Context()
	got := stream.Split(ctx, stream.Of(ctx, 1, 2), 5).Collect()
	assert.Equal(t, [][]int{{1, 2}}, got)
}

func TestFanOut(t *testing.T) {
	ctx := t.Context()
	parts := stream.Of(ctx, 1, 2, 3, 4, 5).FanOut(ctx, 2)
	assert.Len(t, parts, 2)

	// consume both partitions concurrently to avoid deadlock
	results := make([][]int, 2)
	done := make(chan struct{})
	go func() {
		results[1] = parts[1].Collect()
		close(done)
	}()
	results[0] = parts[0].Collect()
	<-done

	assert.Equal(t, []int{1, 3, 5}, results[0])
	assert.Equal(t, []int{2, 4}, results[1])
}

func TestFanOutCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	parts := stream.Of(ctx, 1, 2, 3).FanOut(ctx, 2)
	assert.Nil(t, parts)
}

func TestWindow(t *testing.T) {
	ctx := t.Context()
	got := stream.Window(ctx, stream.Of(ctx, 1, 2, 3, 4, 5), 3).Collect()
	assert.Equal(t, [][]int{{1, 2, 3}, {2, 3, 4}, {3, 4, 5}}, got)
}

func TestWindowLargerThanSource(t *testing.T) {
	ctx := t.Context()
	got := stream.Window(ctx, stream.Of(ctx, 1, 2), 5).Collect()
	assert.Nil(t, got)
}

func TestFanIn(t *testing.T) {
	ctx := t.Context()
	s1 := stream.Of(ctx, 1, 2, 3)
	s2 := stream.Of(ctx, 4, 5, 6)
	got := stream.FanIn(ctx, []stream.Stream[int]{s1, s2}).Collect()
	assert.Len(t, got, 6)
	assert.ElementsMatch(t, []int{1, 2, 3, 4, 5, 6}, got)
}

func TestFanInEmpty(t *testing.T) {
	got := stream.FanIn[int](t.Context(), nil).Collect()
	assert.Nil(t, got)
}

func TestFlatten(t *testing.T) {
	ctx := t.Context()
	chunked := stream.Split(ctx, stream.Of(ctx, 1, 2, 3, 4, 5), 2)
	got := stream.Flatten(ctx, chunked).Collect()
	assert.Equal(t, []int{1, 2, 3, 4, 5}, got)
}

func TestFlattenEmpty(t *testing.T) {
	got := stream.Flatten(t.Context(), stream.Empty[[]int]()).Collect()
	assert.Nil(t, got)
}

func TestForEach(t *testing.T) {
	ctx := t.Context()
	var got []int
	err := stream.Of(ctx, 1, 2, 3).ForEach(ctx, func(_ context.Context, v int) error {
		got = append(got, v)
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, got)
}

func TestForEachError(t *testing.T) {
	ctx := t.Context()
	errBoom := errors.New("boom")
	var count int
	err := stream.Of(ctx, 1, 2, 3, 4, 5).ForEach(ctx, func(_ context.Context, _ int) error {
		count++
		if count == 3 {
			return errBoom
		}
		return nil
	})
	assert.ErrorIs(t, err, errBoom)
	assert.Equal(t, 3, count)
}

func TestForEachEmpty(t *testing.T) {
	err := stream.Empty[int]().ForEach(t.Context(), func(_ context.Context, _ int) error {
		t.Fatal("should not be called")
		return nil
	})
	assert.NoError(t, err)
}

func TestProcess(t *testing.T) {
	ctx := t.Context()
	var mu sync.Mutex
	var got []int
	err := stream.Of(ctx, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10).Process(ctx, 3, func(_ context.Context, v int) error {
		mu.Lock()
		got = append(got, v)
		mu.Unlock()
		return nil
	})
	assert.NoError(t, err)
	assert.ElementsMatch(t, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, got)
}

func TestProcessError(t *testing.T) {
	ctx := t.Context()
	errBoom := errors.New("boom")
	err := stream.Of(ctx, 1, 2, 3).Process(ctx, 2, func(_ context.Context, v int) error {
		if v == 2 {
			return errBoom
		}
		return nil
	})
	assert.ErrorIs(t, err, errBoom)
}

func TestProcessLimit(t *testing.T) {
	ctx := t.Context()
	var active atomic.Int32
	var peak atomic.Int32
	err := stream.Of(ctx, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10).Process(ctx, 3, func(_ context.Context, _ int) error {
		cur := active.Add(1)
		for {
			old := peak.Load()
			if cur <= old || peak.CompareAndSwap(old, cur) {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
		active.Add(-1)
		return nil
	})
	assert.NoError(t, err)
	assert.LessOrEqual(t, peak.Load(), int32(3))
}

func TestPipe(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	send, s := stream.Pipe[int](ctx)

	go func() {
		send(ctx, 1)
		send(ctx, 2)
		send(ctx, 3)
		cancel()
	}()

	got := s.Collect()
	assert.Equal(t, []int{1, 2, 3}, got)
}

func TestPipeCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	send, _ := stream.Pipe[int](ctx)
	assert.ErrorIs(t, send(ctx, 1), context.Canceled)
}

func TestPipeBuffered(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	send, s := stream.Pipe[int](ctx, 3)

	// These should not block since buffer is 3
	assert.NoError(t, send(ctx, 10))
	assert.NoError(t, send(ctx, 20))
	assert.NoError(t, send(ctx, 30))
	cancel()

	got := s.Collect()
	assert.Equal(t, []int{10, 20, 30}, got)
}

func TestPipeFunc(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())

	var got []int
	var mu sync.Mutex

	send := stream.PipeFunc(ctx, func(ctx context.Context, s stream.Stream[int]) error {
		return s.ForEach(ctx, func(_ context.Context, v int) error {
			mu.Lock()
			got = append(got, v)
			mu.Unlock()
			return nil
		})
	})

	// unbuffered pipe — each send blocks until consumed
	go func() {
		send(ctx, 1)
		send(ctx, 2)
		send(ctx, 3)
		cancel() // close pipe after all items sent and consumed
	}()

	// wait for pipe to close
	<-ctx.Done()
	time.Sleep(10 * time.Millisecond) // let ForEach finish

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, []int{1, 2, 3}, got)
}

func TestTake(t *testing.T) {
	ctx := t.Context()
	got := stream.Of(ctx, 1, 2, 3, 4, 5).Take(ctx, 3).Collect()
	assert.Equal(t, []int{1, 2, 3}, got)
}

func TestTakeMoreThanSource(t *testing.T) {
	ctx := t.Context()
	got := stream.Of(ctx, 1, 2).Take(ctx, 10).Collect()
	assert.Equal(t, []int{1, 2}, got)
}

func TestSkip(t *testing.T) {
	ctx := t.Context()
	got := stream.Of(ctx, 1, 2, 3, 4, 5).Skip(ctx, 2).Collect()
	assert.Equal(t, []int{3, 4, 5}, got)
}

func TestSkipAll(t *testing.T) {
	ctx := t.Context()
	got := stream.Of(ctx, 1, 2, 3).Skip(ctx, 10).Collect()
	assert.Nil(t, got)
}

func TestPeek(t *testing.T) {
	ctx := t.Context()
	var peeked []int
	got := stream.Of(ctx, 1, 2, 3).Peek(ctx, func(_ context.Context, n int) {
		peeked = append(peeked, n)
	}).Collect()
	assert.Equal(t, []int{1, 2, 3}, got)
	assert.Equal(t, []int{1, 2, 3}, peeked)
}

func TestTee(t *testing.T) {
	ctx := t.Context()
	streams := stream.Of(ctx, 1, 2, 3).Tee(ctx, 2)
	assert.Len(t, streams, 2)

	// consume both concurrently to avoid deadlock
	results := make([][]int, 2)
	done := make(chan struct{})
	go func() {
		results[1] = streams[1].Collect()
		close(done)
	}()
	results[0] = streams[0].Collect()
	<-done

	assert.Equal(t, []int{1, 2, 3}, results[0])
	assert.Equal(t, []int{1, 2, 3}, results[1])
}

func TestDistinct(t *testing.T) {
	ctx := t.Context()
	got := stream.Of(ctx, "a", "b", "a", "c", "b").Distinct(ctx, func(s string) string {
		return s
	}).Collect()
	assert.Equal(t, []string{"a", "b", "c"}, got)
}

func TestThrottle(t *testing.T) {
	ctx := t.Context()
	start := time.Now()
	got := stream.Of(ctx, 1, 2, 3).Throttle(ctx, 20*time.Millisecond).Collect()
	elapsed := time.Since(start)
	assert.Equal(t, []int{1, 2, 3}, got)
	// 3 items with 20ms throttle: first immediate, then 2 waits = ~40ms minimum
	assert.GreaterOrEqual(t, elapsed, 40*time.Millisecond)
}

func TestReduce(t *testing.T) {
	ctx := t.Context()
	sum, err := stream.Reduce(ctx, stream.Of(ctx, 1, 2, 3, 4, 5), 0, func(_ context.Context, acc, n int) (int, error) {
		return acc + n, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 15, sum)
}

func TestReduceEmpty(t *testing.T) {
	sum, err := stream.Reduce(t.Context(), stream.Empty[int](), 42, func(_ context.Context, acc, n int) (int, error) {
		return acc + n, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 42, sum)
}

func TestReduceError(t *testing.T) {
	ctx := t.Context()
	errBoom := errors.New("boom")
	_, err := stream.Reduce(ctx, stream.Of(ctx, 1, 2, 3), 0, func(_ context.Context, acc, n int) (int, error) {
		if n == 2 {
			return 0, errBoom
		}
		return acc + n, nil
	})
	assert.ErrorIs(t, err, errBoom)
}

func TestMapEach(t *testing.T) {
	ctx := t.Context()
	s1 := stream.Of(ctx, 1, 2, 3)
	s2 := stream.Of(ctx, 4, 5, 6)
	mapped := stream.MapEach(ctx, []stream.Stream[int]{s1, s2}, func(_ context.Context, n int) (string, error) {
		return strconv.Itoa(n), nil
	})
	assert.Len(t, mapped, 2)
	assert.Equal(t, []string{"1", "2", "3"}, mapped[0].Collect())
	assert.Equal(t, []string{"4", "5", "6"}, mapped[1].Collect())
}

func TestFanMap(t *testing.T) {
	ctx := t.Context()
	got := stream.FanMap(ctx, stream.Of(ctx, 1, 2, 3, 4, 5, 6), 3, func(_ context.Context, n int) (int, error) {
		return n * 10, nil
	}).Collect()
	assert.Len(t, got, 6)
	assert.ElementsMatch(t, []int{10, 20, 30, 40, 50, 60}, got)
}

func TestFanMap_withTracing(t *testing.T) {
	ctx := t.Context()

	tp := oteltesting.ReportTraces(t, glossytrace.NewTest(t, glossytrace.WithSpanAttributes()))

	ctx, span := tp.Tracer("test").Start(ctx, "my-pipeline")
	defer span.End()

	got := stream.FanMap(ctx, stream.Of(ctx, 1, 2, 3, 4, 5, 6), 3,
		func(_ context.Context, n int) (int, error) {
			return n * 10, nil
		},
		gofuncy.WithTracerProvider(tp),
	).Collect()

	assert.Len(t, got, 6)
	assert.ElementsMatch(t, []int{10, 20, 30, 40, 50, 60}, got)
}

func TestMapFilter(t *testing.T) {
	ctx := t.Context()
	got := stream.MapFilter(ctx, stream.Of(ctx, 1, 2, 3, 4, 5), func(_ context.Context, n int) (string, bool, error) {
		if n%2 == 0 {
			return "", false, nil // skip evens
		}
		return strconv.Itoa(n), true, nil
	}).Collect()
	assert.Equal(t, []string{"1", "3", "5"}, got)
}

func TestMapFilterDeadLetter(t *testing.T) {
	ctx := t.Context()
	var deadLettered []int
	var mu sync.Mutex
	got := stream.MapFilter(ctx, stream.Of(ctx, 1, 2, 3, 4, 5), func(_ context.Context, n int) (int, bool, error) {
		if n == 3 {
			mu.Lock()
			deadLettered = append(deadLettered, n)
			mu.Unlock()
			return 0, false, nil // dead letter, skip
		}
		return n * 10, true, nil
	}).Collect()
	assert.Equal(t, []int{10, 20, 40, 50}, got)
	assert.Equal(t, []int{3}, deadLettered)
}

func TestMapFilterFatalError(t *testing.T) {
	ctx := t.Context()
	got := stream.MapFilter(ctx, stream.Of(ctx, 1, 2, 3, 4, 5), func(_ context.Context, n int) (int, bool, error) {
		if n == 3 {
			return 0, false, errors.New("fatal")
		}
		return n, true, nil
	}, gofuncy.WithErrorHandler(func(_ context.Context, _ error) {})).Collect()
	assert.Less(t, len(got), 5)
}

func TestMapError(t *testing.T) {
	ctx := t.Context()
	var gotErr error
	errBoom := errors.New("boom")
	got := stream.Map(ctx, stream.Of(ctx, 1, 2, 3, 4, 5), func(_ context.Context, n int) (int, error) {
		if n == 3 {
			return 0, errBoom
		}
		return n * 10, nil
	}, gofuncy.WithErrorHandler(func(_ context.Context, err error) {
		gotErr = err
	})).Collect()
	// Stream closes on error — we get elements before the error
	assert.Less(t, len(got), 5)
	assert.ErrorIs(t, gotErr, errBoom)
}

func TestMapCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	s := stream.Map(ctx, stream.Empty[int](), func(_ context.Context, n int) (string, error) { return strconv.Itoa(n), nil })
	assert.Equal(t, 0, s.Count())
}
