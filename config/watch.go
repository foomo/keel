package config

import (
	"context"
	"time"
)

func WatchBool(ctx context.Context, fn func() bool, callback func(bool)) {
	go func(ctx context.Context, fn func() bool, callback func(bool)) {
		current := fn()
		for {
			select {
			case <-time.After(time.Second):
				if value := fn(); value != current {
					current = value
					callback(current)
				}
			case <-ctx.Done():
				return
			}
		}
	}(ctx, fn, callback)
}

func WatchString(ctx context.Context, fn func() string, callback func(string)) {
	go func(ctx context.Context, fn func() string, callback func(string)) {
		current := fn()
		for {
			select {
			case <-time.After(time.Second):
				if value := fn(); value != current {
					current = value
					callback(current)
				}
			case <-ctx.Done():
				return
			}
		}
	}(ctx, fn, callback)
}
