package cookie

import (
	"time"
)

// TimeProvider returns the current time used to compute cookie expiry.
type TimeProvider func() time.Time

type (
	// TimeProviderOptions configures [NewTimeProvider].
	TimeProviderOptions struct {
		// Offset is added to the current time.
		Offset time.Duration
	}
	// TimeProviderOption configures [TimeProviderOptions].
	TimeProviderOption func(options *TimeProviderOptions)
)

// GetDefaultTimeProviderOptions returns the default options, which use no
// offset.
func GetDefaultTimeProviderOptions() TimeProviderOptions {
	return TimeProviderOptions{}
}

// TimeProviderWithOffset sets the offset added to the current time. Defaults
// to 0.
func TimeProviderWithOffset(v time.Duration) TimeProviderOption {
	return func(o *TimeProviderOptions) {
		o.Offset = v
	}
}

// NewTimeProvider returns a [TimeProvider] reporting [time.Now] plus the
// configured offset.
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
