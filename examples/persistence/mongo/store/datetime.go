package store

import (
	"time"
)

// DateTimeLayout is the ISO 8601 layout with millisecond precision.
const DateTimeLayout = "2006-01-02T15:04:05.000Z0700"

// DateTime is a time formatted with DateTimeLayout.
type DateTime string

// NewDateTime returns t formatted as DateTime.
func NewDateTime(t time.Time) DateTime {
	return DateTime(t.Format(DateTimeLayout))
}

// Time parses the date time into a time.Time.
func (d DateTime) Time() (time.Time, error) {
	return time.Parse(DateTimeLayout, string(d))
}

// MustTime is like Time but panics on failure.
func (d DateTime) MustTime() time.Time {
	t, err := d.Time()
	if err != nil {
		panic("failed to parse time: " + err.Error())
	}

	return t
}

// String returns the string representation.
func (d DateTime) String() string {
	return string(d)
}
