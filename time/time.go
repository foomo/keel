package keeltime

import (
	"time"
)

var (
	// Deprecated: Use time.Now instead
	Now = time.Now
	// Deprecated: Use stdlib time package instead
	NowStaticNSec = int64(1609498800e9) // 2021-01-01 12:00:00
	// Deprecated: Use stdlib time package instead
	NowIncrementalNSec = NowStaticNSec
)

// Static sets now to a static time provider
// Deprecated: Use stdlib time package instead
func Static() {
	Now = static
}

// Incremental sets now to a incremental time provider
// Deprecated: Use stdlib time package instead
func Incremental() {
	Now = incremental
}

func static() time.Time {
	return time.Unix(0, NowStaticNSec)
}

func incremental() time.Time {
	t := time.Unix(0, NowIncrementalNSec)
	NowIncrementalNSec++
	return t
}

// ResetIncremental sets the incremental time to the static default
// Deprecated: Use stdlib time package instead
func ResetIncremental() {
	NowIncrementalNSec = NowStaticNSec
}
