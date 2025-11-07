package log

import (
	"time"

	"go.uber.org/zap"
)

const (
	// DurationKey - generic duration attribute
	DurationKey = "duration"
	// DurationSecKey - duration in seconds
	DurationSecKey = "duration_sec"
	// DurationMinKey - duration in minutes
	DurationMinKey = "duration_min"
	// DurationHourKey - duration in hours
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

// FDurationFn returns a function that returns a zap.Field with a given time.Duration converted to milliseconds under the key "duration".
func FDurationFn() func() zap.Field {
	start := time.Now()

	return func() zap.Field {
		return FDuration(time.Since(start))
	}
}
