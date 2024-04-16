package config

import (
	"context"
	"time"
)

// Watch callback

func WatchBool(ctx context.Context, fn func() bool, callback func(bool)) {
	current := fn()
	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			callback(current)
		}
	})
}

func WatchTime(ctx context.Context, fn func() time.Time, callback func(time.Time)) {
	current := fn()
	watch(ctx, func() {
		if value := fn(); !value.Equal(current) {
			current = value
			callback(current)
		}
	})
}

func WatchDuration(ctx context.Context, fn func() time.Duration, callback func(time.Duration)) {
	current := fn()
	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			callback(current)
		}
	})
}

func WatchInt(ctx context.Context, fn func() int, callback func(int)) {
	current := fn()
	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			callback(current)
		}
	})
}

func WatchInt32(ctx context.Context, fn func() int32, callback func(int32)) {
	current := fn()
	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			callback(current)
		}
	})
}

func WatchInt64(ctx context.Context, fn func() int64, callback func(int64)) {
	current := fn()
	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			callback(current)
		}
	})
}

func WatchFloat64(ctx context.Context, fn func() float64, callback func(float64)) {
	current := fn()
	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			callback(current)
		}
	})
}

func WatchString(ctx context.Context, fn func() string, callback func(string)) {
	current := fn()
	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			callback(current)
		}
	})
}

// Watch channel

func WatchBoolChan(ctx context.Context, fn func() bool, ch chan bool) {
	current := fn()
	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			ch <- current
		}
	})
}

func WatchTimeChan(ctx context.Context, fn func() time.Time, ch chan time.Time) {
	current := fn()
	watch(ctx, func() {
		if value := fn(); !value.Equal(current) {
			current = value
			ch <- current
		}
	})
}

func WatchDurationChan(ctx context.Context, fn func() time.Duration, ch chan time.Duration) {
	current := fn()
	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			ch <- current
		}
	})
}

func WatchIntChan(ctx context.Context, fn func() int, ch chan int) {
	current := fn()
	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			ch <- current
		}
	})
}

func WatchInt32Chan(ctx context.Context, fn func() int32, ch chan int32) {
	current := fn()
	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			ch <- current
		}
	})
}

func WatchInt64Chan(ctx context.Context, fn func() int64, ch chan int64) {
	current := fn()
	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			ch <- current
		}
	})
}

func WatchFloat64Chan(ctx context.Context, fn func() float64, ch chan float64) {
	current := fn()
	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			ch <- current
		}
	})
}

func WatchStringChan(ctx context.Context, fn func() string, ch chan string) {
	current := fn()
	watch(ctx, func() {
		if value := fn(); value != current {
			current = value
			ch <- current
		}
	})
}

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
