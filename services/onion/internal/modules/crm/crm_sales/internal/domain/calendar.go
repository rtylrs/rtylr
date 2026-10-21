package domain

import "time"

// CallerDay is the caller's calendar day: follow-up dates are compared to
// Date, and "today" counts start at StartUTC, the caller's midnight in UTC.
type CallerDay struct {
	Date     time.Time
	StartUTC time.Time
}

// MonthStartUTC is the caller's first of the month, as a UTC instant.
func (d CallerDay) MonthStartUTC() time.Time {
	return d.StartOf(time.Date(d.Date.Year(), d.Date.Month(), 1, 0, 0, 0, 0, time.UTC))
}

// StartOf is when one of the caller's calendar days began, as a UTC instant.
func (d CallerDay) StartOf(date time.Time) time.Time {
	day := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	return day.Add(d.StartUTC.Sub(d.Date))
}

// NewCallerDay reads the caller's date (YYYY-MM-DD) and UTC offset in
// minutes. Without a valid date, the server's UTC day is used.
func NewCallerDay(date string, utcOffsetMinutes int32, now time.Time) CallerDay {
	if utcOffsetMinutes < -14*60 || utcOffsetMinutes > 14*60 {
		utcOffsetMinutes = 0
	}
	offset := time.Duration(utcOffsetMinutes) * time.Minute
	day, err := time.Parse(time.DateOnly, date)
	if err != nil {
		local := now.UTC().Add(offset)
		day = time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)
	}
	return CallerDay{Date: day, StartUTC: day.Add(-offset)}
}
