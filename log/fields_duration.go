package log

import (
	"time"

	"go.uber.org/zap"
)

const (
	DurationKey     = "duration"
	DurationSecKey  = "duration_sec"
	DurationMinKey  = "duration_min"
	DurationHourKey = "duration_hour"
)

func FDuration(duration time.Duration) zap.Field {
	return zap.Float64(DurationKey, float64(duration)/float64(time.Millisecond))
}

func FDurationSec(duration time.Duration) zap.Field {
	return zap.Float64(DurationSecKey, float64(duration)/float64(time.Second))
}

func FDurationMin(duration time.Duration) zap.Field {
	return zap.Float64(DurationMinKey, float64(duration)/float64(time.Minute))
}

func FDurationHour(duration time.Duration) zap.Field {
	return zap.Float64(DurationHourKey, float64(duration)/float64(time.Hour))
}

func FDurationFn() func() zap.Field {
	start := time.Now()
	return func() zap.Field {
		return FDuration(time.Since(start))
	}
}
