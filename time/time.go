package keeltime

import (
	"time"
)

var (
	Now                = time.Now
	NowStaticNSec      = int64(1609498800e9) // 2021-01-01 12:00:00
	NowIncrementalNSec = NowStaticNSec
)

// Static sets now to a static time provider
func Static() {
	Now = static
}

// Incremental sets now to a incremental time provider
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
func ResetIncremental() {
	NowIncrementalNSec = NowStaticNSec
}
