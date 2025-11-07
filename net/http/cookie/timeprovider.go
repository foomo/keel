package cookie

import (
	"time"
)

type TimeProvider func() time.Time

type (
	TimeProviderOptions struct {
		Offset time.Duration
	}
	TimeProviderOption func(options *TimeProviderOptions)
)

func GetDefaultTimeProviderOptions() TimeProviderOptions {
	return TimeProviderOptions{}
}

func TimeProviderWithOffset(v time.Duration) TimeProviderOption {
	return func(o *TimeProviderOptions) {
		o.Offset = v
	}
}

func NewTimeProvider(opts ...TimeProviderOption) TimeProvider {
	options := GetDefaultTimeProviderOptions()

	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return func() time.Time {
		return time.Now().Add(options.Offset)
	}
}
