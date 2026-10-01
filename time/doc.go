// Package keeltime provides a replaceable clock for deterministic tests.
//
// Code reads the current time through [Now]. Tests can switch it to a fixed
// time with [Static] or to a strictly increasing time with [Incremental].
// The package variables are not synchronized; switch providers before
// concurrent use.
package keeltime
