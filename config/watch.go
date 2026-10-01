package config

import (
	"context"
	"time"
)

// WatchBool is [Watch] for bool values.
func WatchBool(ctx context.Context, fn func() bool, callback func(bool)) {
	current := fn()

	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			callback(current)
		}
	})
}

// WatchTime is [Watch] for time.Time values.
func WatchTime(ctx context.Context, fn func() time.Time, callback func(time.Time)) {
	current := fn()

	watch(ctx, func() {
		if value := fn(); !value.Equal(current) {
			current = value
			callback(current)
		}
	})
}

// WatchDuration is [Watch] for time.Duration values.
func WatchDuration(ctx context.Context, fn func() time.Duration, callback func(time.Duration)) {
	current := fn()

	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			callback(current)
		}
	})
}

// WatchInt is [Watch] for int values.
func WatchInt(ctx context.Context, fn func() int, callback func(int)) {
	current := fn()

	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			callback(current)
		}
	})
}

// WatchInt32 is [Watch] for int32 values.
func WatchInt32(ctx context.Context, fn func() int32, callback func(int32)) {
	current := fn()

	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			callback(current)
		}
	})
}

// WatchInt64 is [Watch] for int64 values.
func WatchInt64(ctx context.Context, fn func() int64, callback func(int64)) {
	current := fn()

	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			callback(current)
		}
	})
}

// WatchFloat64 is [Watch] for float64 values.
func WatchFloat64(ctx context.Context, fn func() float64, callback func(float64)) {
	current := fn()

	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			callback(current)
		}
	})
}

// WatchString is [Watch] for string values.
func WatchString(ctx context.Context, fn func() string, callback func(string)) {
	current := fn()

	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			callback(current)
		}
	})
}

// WatchBoolChan is [WatchChan] for bool values.
func WatchBoolChan(ctx context.Context, fn func() bool, ch chan bool) {
	current := fn()

	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			ch <- current
		}
	})
}

// WatchTimeChan is [WatchChan] for time.Time values.
func WatchTimeChan(ctx context.Context, fn func() time.Time, ch chan time.Time) {
	current := fn()

	watch(ctx, func() {
		if value := fn(); !value.Equal(current) {
			current = value
			ch <- current
		}
	})
}

// WatchDurationChan is [WatchChan] for time.Duration values.
func WatchDurationChan(ctx context.Context, fn func() time.Duration, ch chan time.Duration) {
	current := fn()

	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			ch <- current
		}
	})
}

// WatchIntChan is [WatchChan] for int values.
func WatchIntChan(ctx context.Context, fn func() int, ch chan int) {
	current := fn()

	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			ch <- current
		}
	})
}

// WatchInt32Chan is [WatchChan] for int32 values.
func WatchInt32Chan(ctx context.Context, fn func() int32, ch chan int32) {
	current := fn()

	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			ch <- current
		}
	})
}

// WatchInt64Chan is [WatchChan] for int64 values.
func WatchInt64Chan(ctx context.Context, fn func() int64, ch chan int64) {
	current := fn()

	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			ch <- current
		}
	})
}

// WatchFloat64Chan is [WatchChan] for float64 values.
func WatchFloat64Chan(ctx context.Context, fn func() float64, ch chan float64) {
	current := fn()

	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			ch <- current
		}
	})
}

// WatchStringChan is [WatchChan] for string values.
func WatchStringChan(ctx context.Context, fn func() string, ch chan string) {
	current := fn()

	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			ch <- current
		}
	})
}

// watch calls fn every second in a new goroutine until ctx is done.
func watch(ctx context.Context, fn func()) {
	go func(ctx context.Context, fn func()) {
		for {
			select {
			case <-time.After(time.Second):
				fn()
			case <-ctx.Done():
				return
			}
		}
	}(ctx, fn)
}
