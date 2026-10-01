package log

import (
	"time"

	"go.uber.org/zap"
)

const (
	// DurationKey is the log field key for a duration in milliseconds.
	DurationKey = "duration"
	// DurationSecKey is the log field key for a duration in seconds.
	DurationSecKey = "duration_sec"
	// DurationMinKey is the log field key for a duration in minutes.
	DurationMinKey = "duration_min"
	// DurationHourKey is the log field key for a duration in hours.
	DurationHourKey = "duration_hour"
)

// FDuration creates a zap.Field with a given time.Duration converted to milliseconds under the key "duration".
func FDuration(duration time.Duration) zap.Field {
	return zap.Int64(DurationKey, duration.Milliseconds())
}

// FDurationSec creates a zap.Field with a given time.Duration converted to seconds under the key "duration_sec".
func FDurationSec(duration time.Duration) zap.Field {
	return zap.Float64(DurationSecKey, duration.Seconds())
}

// FDurationMin creates a zap.Field with a given time.Duration converted to minutes under the key "duration_min".
func FDurationMin(duration time.Duration) zap.Field {
	return zap.Float64(DurationMinKey, duration.Minutes())
}

// FDurationHour creates a zap.Field with a given time.Duration converted to hours under the key "duration_hour".
func FDurationHour(duration time.Duration) zap.Field {
	return zap.Float64(DurationHourKey, duration.Hours())
}

// FDurationFn starts a timer and returns a function that creates a zap.Field
// with the time elapsed since the call to FDurationFn in milliseconds under the
// key "duration".
func FDurationFn() func() zap.Field {
	start := time.Now()

	return func() zap.Field {
		return FDuration(time.Since(start))
	}
}
