package keeltime

import (
	"time"
)

var (
	// Now returns the current time. It defaults to [time.Now] and is replaced
	// by [Static] and [Incremental].
	Now = time.Now
	// NowStaticNSec is the Unix time in nanoseconds returned by the static
	// provider (2021-01-01 11:00:00 UTC).
	NowStaticNSec = int64(1609498800e9)
	// NowIncrementalNSec is the Unix time in nanoseconds the incremental
	// provider returns next.
	NowIncrementalNSec = NowStaticNSec
)

// Static sets [Now] to always return [NowStaticNSec].
func Static() {
	Now = static
}

// Incremental sets [Now] to return [NowIncrementalNSec] and advance it by one
// nanosecond on each call. It is not safe for concurrent use.
func Incremental() {
	Now = incremental
}

// static returns the time at NowStaticNSec.
func static() time.Time {
	return time.Unix(0, NowStaticNSec)
}

// incremental returns the time at NowIncrementalNSec and increments it.
func incremental() time.Time {
	t := time.Unix(0, NowIncrementalNSec)
	NowIncrementalNSec++

	return t
}

// ResetIncremental resets [NowIncrementalNSec] to [NowStaticNSec].
func ResetIncremental() {
	NowIncrementalNSec = NowStaticNSec
}
