package domain

import (
	"math"
	"time"
)

const (
	// InsightsThroughputWeeks is how many weeks, the current one included, the
	// throughput series covers.
	InsightsThroughputWeeks = 6
	// InsightsVelocityCycles is how many completed cycles the velocity series
	// covers.
	InsightsVelocityCycles = 6
	// InsightsCycleTimeDays is the completion window the cycle time averages.
	InsightsCycleTimeDays = 90

	maxBurndownDays = 366
)

func utcDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// WeekStart is the Monday that starts t's week, at midnight UTC.
func WeekStart(t time.Time) time.Time {
	day := utcDay(t)
	return day.AddDate(0, 0, -((int(day.Weekday()) + 6) % 7))
}

// WeeklyThroughput folds completions per day (yyyy-mm-dd) into the last weeks
// weeks ending with today's, oldest first, with empty weeks kept.
func WeeklyThroughput(byDay map[string]int, today time.Time, weeks int) []InsightsThroughput {
	first := WeekStart(today).AddDate(0, 0, -7*(weeks-1))
	out := make([]InsightsThroughput, weeks)
	for i := range out {
		out[i].Week = first.AddDate(0, 0, 7*i).Format("2006-01-02")
	}
	for day, count := range byDay {
		at, err := time.Parse("2006-01-02", day)
		if err != nil || at.Before(first) {
			continue
		}
		if i := int(at.Sub(first).Hours()/24) / 7; i < weeks {
			out[i].Completed += count
		}
	}
	return out
}

// CycleBurndown is one point per cycle day, from the start date through today,
// of the scoped tasks still open at the end of that day. completions holds one
// entry per task in scope: when it was finished, or nil while it is open.
// Ideal burns the full scope down to zero by the end date.
//
// It is empty for a cycle with nothing in scope or not yet begun. A cycle on
// its first day still gets its second day, because the chart draws a line.
func CycleBurndown(start, end, today time.Time, completions []*time.Time) []InsightsBurndownPoint {
	start, end, today = utcDay(start), utcDay(end), utcDay(today)
	total := len(completions)
	if total == 0 || end.Before(start) || today.Before(start) {
		return []InsightsBurndownPoint{}
	}
	span := int(end.Sub(start).Hours() / 24)
	if span > maxBurndownDays {
		span = maxBurndownDays
	}
	last := int(today.Sub(start).Hours() / 24)
	if last > span {
		last = span
	}
	if last == 0 && span > 0 {
		last = 1
	}
	points := make([]InsightsBurndownPoint, 0, last+1)
	for d := 0; d <= last; d++ {
		dayEnd := start.AddDate(0, 0, d+1)
		remaining := total
		for _, at := range completions {
			if at != nil && at.Before(dayEnd) {
				remaining--
			}
		}
		ideal := 0.0
		if span > 0 {
			ideal = math.Round(float64(total)*(1-float64(d)/float64(span))*10) / 10
		}
		points = append(points, InsightsBurndownPoint{Day: d, Ideal: ideal, Actual: float64(remaining)})
	}
	return points
}
