package stream

import (
	"context"
	"slices"
	"time"

	"github.com/foomo/gofuncy"
)

type Stream[T any] struct {
	source <-chan T
}

func (s Stream[T]) Chan() <-chan T {
	return s.source
}

func (s Stream[T]) Count() int {
	var count int
	for range s.source {
		count++
	}
	return count
}

func (s Stream[T]) Collect() []T {
	var out []T
	for v := range s.source {
		out = append(out, v)
	}
	return out
}

// Filter returns a stream containing only elements where fn returns true.
func (s Stream[T]) Filter(ctx context.Context, fn func(context.Context, T) bool, opts ...gofuncy.GoOption) Stream[T] {
	if ctx.Err() != nil {
		return Empty[T]()
	}

	source := make(chan T)
	gofuncy.Go(ctx, func(ctx context.Context) error {
		defer close(source)
		for item := range s.source {
			if !fn(ctx, item) {
				continue
			}
			select {
			case <-ctx.Done():
				return nil
			case source <- item:
			}
		}
		return nil
	}, opts...)
	return From[T](source)
}

// Sort collects all elements, sorts them using cmp, and emits in sorted order.
func (s Stream[T]) Sort(ctx context.Context, cmp func(T, T) int, opts ...gofuncy.GoOption) Stream[T] {
	if ctx.Err() != nil {
		return Empty[T]()
	}

	source := make(chan T)
	gofuncy.Go(ctx, func(ctx context.Context) error {
		defer close(source)
		items := s.Collect()
		slices.SortFunc(items, cmp)
		for _, item := range items {
			select {
			case <-ctx.Done():
				return nil
			case source <- item:
			}
		}
		return nil
	}, opts...)
	return From[T](source)
}

// Reverse collects all elements and emits them in reverse order.
func (s Stream[T]) Reverse(ctx context.Context, opts ...gofuncy.GoOption) Stream[T] {
	if ctx.Err() != nil {
		return Empty[T]()
	}

	source := make(chan T)
	gofuncy.Go(ctx, func(ctx context.Context) error {
		defer close(source)
		items := s.Collect()
		for i := len(items) - 1; i >= 0; i-- {
			select {
			case <-ctx.Done():
				return nil
			case source <- items[i]:
			}
		}
		return nil
	}, opts...)
	return From[T](source)
}

// Take emits the first n elements then closes the stream.
func (s Stream[T]) Take(ctx context.Context, n int, opts ...gofuncy.GoOption) Stream[T] {
	if n <= 0 || ctx.Err() != nil {
		return Empty[T]()
	}

	source := make(chan T)
	gofuncy.Go(ctx, func(ctx context.Context) error {
		defer close(source)
		count := 0
		for item := range s.source {
			if count >= n {
				return nil
			}
			select {
			case <-ctx.Done():
				return nil
			case source <- item:
				count++
			}
		}
		return nil
	}, opts...)
	return From[T](source)
}

// Skip drops the first n elements and emits the rest.
func (s Stream[T]) Skip(ctx context.Context, n int, opts ...gofuncy.GoOption) Stream[T] {
	if ctx.Err() != nil {
		return Empty[T]()
	}

	source := make(chan T)
	gofuncy.Go(ctx, func(ctx context.Context) error {
		defer close(source)
		count := 0
		for item := range s.source {
			if count < n {
				count++
				continue
			}
			select {
			case <-ctx.Done():
				return nil
			case source <- item:
			}
		}
		return nil
	}, opts...)
	return From[T](source)
}

// Peek calls fn for each element as a side-effect and forwards the element unchanged.
func (s Stream[T]) Peek(ctx context.Context, fn func(context.Context, T), opts ...gofuncy.GoOption) Stream[T] {
	if ctx.Err() != nil {
		return Empty[T]()
	}

	source := make(chan T)
	gofuncy.Go(ctx, func(ctx context.Context) error {
		defer close(source)
		for item := range s.source {
			fn(ctx, item)
			select {
			case <-ctx.Done():
				return nil
			case source <- item:
			}
		}
		return nil
	}, opts...)
	return From[T](source)
}

// Tee broadcasts every element to n output streams.
// Unlike FanOut which round-robins, Tee sends each element to all streams.
func (s Stream[T]) Tee(ctx context.Context, n int, opts ...gofuncy.GoOption) []Stream[T] {
	if n <= 0 || ctx.Err() != nil {
		return nil
	}

	sources := make([]chan T, n)
	for i := range sources {
		sources[i] = make(chan T)
	}

	gofuncy.Go(ctx, func(ctx context.Context) error {
		defer func() {
			for _, ch := range sources {
				close(ch)
			}
		}()
		for item := range s.source {
			for _, ch := range sources {
				select {
				case <-ctx.Done():
					return nil
				case ch <- item:
				}
			}
		}
		return nil
	}, opts...)

	streams := make([]Stream[T], n)
	for i, ch := range sources {
		streams[i] = From[T](ch)
	}
	return streams
}

// Distinct deduplicates elements using a key function. First occurrence wins.
func (s Stream[T]) Distinct(ctx context.Context, key func(T) string, opts ...gofuncy.GoOption) Stream[T] {
	if ctx.Err() != nil {
		return Empty[T]()
	}

	source := make(chan T)
	gofuncy.Go(ctx, func(ctx context.Context) error {
		defer close(source)
		seen := make(map[string]struct{})
		for item := range s.source {
			k := key(item)
			if _, ok := seen[k]; ok {
				continue
			}
			seen[k] = struct{}{}
			select {
			case <-ctx.Done():
				return nil
			case source <- item:
			}
		}
		return nil
	}, opts...)
	return From[T](source)
}

// Throttle rate-limits the stream to at most one element per duration d.
func (s Stream[T]) Throttle(ctx context.Context, d time.Duration, opts ...gofuncy.GoOption) Stream[T] {
	if ctx.Err() != nil {
		return Empty[T]()
	}

	source := make(chan T)
	gofuncy.Go(ctx, func(ctx context.Context) error {
		defer close(source)
		ticker := time.NewTicker(d)
		defer ticker.Stop()
		first := true
		for item := range s.source {
			if !first {
				select {
				case <-ctx.Done():
					return nil
				case <-ticker.C:
				}
			}
			first = false
			select {
			case <-ctx.Done():
				return nil
			case source <- item:
			}
		}
		return nil
	}, opts...)
	return From[T](source)
}

// ForEach consumes the stream, calling fn for each element.
// Returns the first error from fn or ctx, nil when fully consumed.
func (s Stream[T]) ForEach(ctx context.Context, fn func(context.Context, T) error) error {
	for item := range s.source {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := fn(ctx, item); err != nil {
			return err
		}
	}
	return nil
}

// Process consumes the stream, dispatching each element to a worker pool of size n.
// All errors are collected and returned via errors.Join.
func (s Stream[T]) Process(ctx context.Context, n int, fn func(context.Context, T) error, opts ...gofuncy.GroupOption) error {
	g := gofuncy.NewGroup(ctx, append([]gofuncy.GroupOption{
		gofuncy.WithLimit(n),
	}, opts...)...)
	for item := range s.source {
		g.Add(func(ctx context.Context) error {
			return fn(ctx, item)
		})
	}
	return g.Wait()
}

// Reduce folds all elements into a single value using fn.
// Returns the accumulated result or the first error from fn.
func Reduce[T, U any](ctx context.Context, s Stream[T], initial U, fn func(context.Context, U, T) (U, error)) (U, error) {
	acc := initial
	for item := range s.source {
		if ctx.Err() != nil {
			return acc, ctx.Err()
		}
		var err error
		acc, err = fn(ctx, acc, item)
		if err != nil {
			return acc, err
		}
	}
	return acc, nil
}

func Empty[T any]() Stream[T] {
	source := make(chan T)
	close(source)
	return Stream[T]{source: source}
}

// Of Returns a Stream based any element
func Of[T any](ctx context.Context, items ...T) Stream[T] {
	n := len(items)
	if n == 0 || ctx.Err() != nil {
		return Empty[T]()
	}

	source := make(chan T, n)
	gofuncy.Go(ctx, func(ctx context.Context) error {
		defer close(source)
		for _, item := range items {
			select {
			case <-ctx.Done():
				return nil
			case source <- item:
			}
		}
		return nil
	})
	return From[T](source)
}

// Pipe creates a writable stream entry point.
// Returns a send function and the readable stream.
// The send function returns ctx.Err() if the context is cancelled.
// The channel is closed when ctx is done.
func Pipe[T any](ctx context.Context, bufferSize ...int) (func(context.Context, T) error, Stream[T]) {
	size := 0
	if len(bufferSize) > 0 {
		size = bufferSize[0]
	}
	ch := make(chan T, size)
	go func() {
		<-ctx.Done()
		close(ch)
	}()
	send := func(ctx context.Context, v T) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ch <- v:
			return nil
		}
	}
	return send, From[T](ch)
}

// PipeFunc creates a Pipe and launches the consumer fn in a gofuncy.Go goroutine.
// Returns only the send handler.
func PipeFunc[T any](ctx context.Context, fn func(context.Context, Stream[T]) error, opts ...gofuncy.GoOption) func(context.Context, T) error {
	send, s := Pipe[T](ctx)
	gofuncy.Go(ctx, func(ctx context.Context) error {
		return fn(ctx, s)
	}, opts...)
	return send
}

func From[T any](source <-chan T) Stream[T] {
	return Stream[T]{source: source}
}

// FanOut distributes elements round-robin across n output streams.
func (s Stream[T]) FanOut(ctx context.Context, n int, opts ...gofuncy.GoOption) []Stream[T] {
	if n <= 0 || ctx.Err() != nil {
		return nil
	}

	sources := make([]chan T, n)
	for i := range sources {
		sources[i] = make(chan T)
	}

	gofuncy.Go(ctx, func(ctx context.Context) error {
		defer func() {
			for _, ch := range sources {
				close(ch)
			}
		}()
		i := 0
		for item := range s.source {
			select {
			case <-ctx.Done():
				return nil
			case sources[i%n] <- item:
			}
			i++
		}
		return nil
	}, append(opts, gofuncy.WithName("stream.fan-out"))...)

	streams := make([]Stream[T], n)
	for i, ch := range sources {
		streams[i] = From[T](ch)
	}
	return streams
}

// Split groups consecutive elements into batches of size n.
// The last batch may contain fewer than n elements.
func Split[T any](ctx context.Context, s Stream[T], n int, opts ...gofuncy.GoOption) Stream[[]T] {
	if n <= 0 || ctx.Err() != nil {
		return Empty[[]T]()
	}

	source := make(chan []T)
	gofuncy.Go(ctx, func(ctx context.Context) error {
		defer close(source)
		batch := make([]T, 0, n)
		for item := range s.source {
			batch = append(batch, item)
			if len(batch) == n {
				select {
				case <-ctx.Done():
					return nil
				case source <- batch:
				}
				batch = make([]T, 0, n)
			}
		}
		if len(batch) > 0 {
			select {
			case <-ctx.Done():
				return nil
			case source <- batch:
			}
		}
		return nil
	}, opts...)
	return From[[]T](source)
}

// Window emits sliding windows of n consecutive elements.
// If the source has fewer than n elements, no windows are emitted.
func Window[T any](ctx context.Context, s Stream[T], n int, opts ...gofuncy.GoOption) Stream[[]T] {
	if n <= 0 || ctx.Err() != nil {
		return Empty[[]T]()
	}

	source := make(chan []T)
	gofuncy.Go(ctx, func(ctx context.Context) error {
		defer close(source)
		buf := make([]T, 0, n)
		for item := range s.source {
			if len(buf) == n {
				buf = buf[1:]
			}
			buf = append(buf, item)
			if len(buf) == n {
				win := make([]T, n)
				copy(win, buf)
				select {
				case <-ctx.Done():
					return nil
				case source <- win:
				}
			}
		}
		return nil
	}, opts...)
	return From[[]T](source)
}

// FanIn combines multiple streams into a single stream.
// Elements arrive in non-deterministic order as they become available.
func FanIn[T any](ctx context.Context, streams []Stream[T], opts ...gofuncy.GoOption) Stream[T] {
	if len(streams) == 0 || ctx.Err() != nil {
		return Empty[T]()
	}

	source := make(chan T)
	g := gofuncy.NewGroup(ctx)

	for _, s := range streams {
		g.Add(func(ctx context.Context) error {
			for item := range s.source {
				select {
				case <-ctx.Done():
					return nil
				case source <- item:
				}
			}
			return nil
		})
	}

	gofuncy.Go(ctx, func(ctx context.Context) error {
		defer close(source)
		return g.Wait()
	}, append(opts, gofuncy.WithName("stream.fan-in"))...)

	return From[T](source)
}

// Flatten flattens a stream of slices into a stream of individual elements.
func Flatten[T any](ctx context.Context, s Stream[[]T], opts ...gofuncy.GoOption) Stream[T] {
	if ctx.Err() != nil {
		return Empty[T]()
	}

	source := make(chan T)
	gofuncy.Go(ctx, func(ctx context.Context) error {
		defer close(source)
		for batch := range s.source {
			for _, item := range batch {
				select {
				case <-ctx.Done():
					return nil
				case source <- item:
				}
			}
		}
		return nil
	}, opts...)
	return From[T](source)
}

// Map returns a new Stream by applying fn to each element of the source stream.
// If fn returns an error, the stream closes and the error is handled by gofuncy.Go.
func Map[T, U any](ctx context.Context, s Stream[T], fn func(context.Context, T) (U, error), opts ...gofuncy.GoOption) Stream[U] {
	if ctx.Err() != nil {
		return Empty[U]()
	}

	source := make(chan U)
	gofuncy.Go(ctx, func(ctx context.Context) error {
		defer close(source)
		for item := range s.source {
			v, err := fn(ctx, item)
			if err != nil {
				return err
			}
			select {
			case <-ctx.Done():
				return nil
			case source <- v:
			}
		}
		return nil
	}, append(opts, gofuncy.WithName("stream.map"))...)
	return From[U](source)
}

// MapFilter maps and filters elements. The bool controls emission:
// (val, true, nil) emits val; (_, false, nil) skips the item; (_, _, err) stops the stream.
func MapFilter[T, U any](ctx context.Context, s Stream[T], fn func(context.Context, T) (U, bool, error), opts ...gofuncy.GoOption) Stream[U] {
	if ctx.Err() != nil {
		return Empty[U]()
	}

	source := make(chan U)
	gofuncy.Go(ctx, func(ctx context.Context) error {
		defer close(source)
		for item := range s.source {
			v, ok, err := fn(ctx, item)
			if err != nil {
				return err
			}
			if !ok {
				continue
			}
			select {
			case <-ctx.Done():
				return nil
			case source <- v:
			}
		}
		return nil
	}, opts...)
	return From[U](source)
}

// MapFilterEach applies MapFilter to each stream in a slice.
func MapFilterEach[T, U any](ctx context.Context, streams []Stream[T], fn func(context.Context, T) (U, bool, error), opts ...gofuncy.GoOption) []Stream[U] {
	out := make([]Stream[U], len(streams))
	for i, s := range streams {
		out[i] = MapFilter(ctx, s, fn, opts...)
	}
	return out
}

// FanMapFilter fans out, applies MapFilter concurrently, and fans in the results.
func FanMapFilter[T, U any](ctx context.Context, s Stream[T], n int, fn func(context.Context, T) (U, bool, error), opts ...gofuncy.GoOption) Stream[U] {
	return FanIn(ctx, MapFilterEach(ctx, s.FanOut(ctx, n, opts...), fn, opts...))
}

// MapEach applies Map to each stream in a slice, returning a slice of transformed streams.
func MapEach[T, U any](ctx context.Context, streams []Stream[T], fn func(context.Context, T) (U, error), opts ...gofuncy.GoOption) []Stream[U] {
	out := make([]Stream[U], len(streams))
	for i, s := range streams {
		out[i] = Map(ctx, s, fn, opts...)
	}
	return out
}

// FanMap fans out a stream into n partitions, maps each concurrently, and fans in the results.
// Output order is non-deterministic.
func FanMap[T, U any](ctx context.Context, s Stream[T], n int, fn func(context.Context, T) (U, error), opts ...gofuncy.GoOption) Stream[U] {
	return FanIn(ctx, MapEach(ctx, s.FanOut(ctx, n, opts...), fn, opts...), opts...)
}
