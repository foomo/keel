package store

import (
	"time"
)

// DateTimeLayout in the ISO8601 format with millisecond precision
const DateTimeLayout = "2006-01-02T15:04:05.000Z0700"

// DateTime type
type DateTime string

// NewDateTime constructor
func NewDateTime(t time.Time) DateTime {
	return DateTime(t.Format(DateTimeLayout))
}

// Time returns the date time as Time
func (d DateTime) Time() (time.Time, error) {
	return time.Parse(DateTimeLayout, string(d))
}

// MustTime returns the date time as Time and panics on failure
func (d DateTime) MustTime() time.Time {
	t, err := d.Time()
	if err != nil {
		panic("failed to parse time: " + err.Error())
	}
	return t
}

// String returns the string representation
func (d DateTime) String() string {
	return string(d)
}
