package domain

import "time"

// Seat periods are stored as DATE columns, so every boundary is a UTC calendar
// day. Keeping the normalisation in one place stops the service and repository
// layers from disagreeing about which day a boundary falls on.
func UTCDay(t time.Time) time.Time {
	u := t.UTC()
	return time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
}

// FirstOfUTCMonth is the billing_month key: midnight UTC on the 1st.
func FirstOfUTCMonth(t time.Time) time.Time {
	u := t.UTC()
	return time.Date(u.Year(), u.Month(), 1, 0, 0, 0, 0, time.UTC)
}

// LastOfUTCMonth is the final billable day of a billing month, so 28, 29 in a
// leap February and 30 or 31 otherwise.
func LastOfUTCMonth(t time.Time) time.Time {
	return FirstOfUTCMonth(t).AddDate(0, 1, -1)
}

// DaysInUTCMonth is the proration denominator. A leap February returns 29, so
// a full-month period always bills exactly pricePerSeat per seat.
func DaysInUTCMonth(t time.Time) int {
	return int(FirstOfUTCMonth(t).AddDate(0, 1, 0).Sub(FirstOfUTCMonth(t)).Hours() / 24)
}
