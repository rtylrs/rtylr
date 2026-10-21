package hr

// ShiftMetrics summarizes shift assignment outcomes over a date range.
// Scheduled excludes cancelled assignments.
type ShiftMetrics struct {
	Scheduled int64
	Completed int64
	NoShows   int64
}
