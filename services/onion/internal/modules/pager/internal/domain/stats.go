package domain

import (
	"slices"
	"time"
)

const (
	DefaultStatsDays = 30
	MaxStatsDays     = 90
)

// TicketTimes is what the stats read of one labelled ticket.
type TicketTimes struct {
	CreatedAt time.Time
	// ReadyAt is zero for an order nobody marked ready.
	ReadyAt time.Time
}

type DayCount struct {
	// Date is the local date, YYYY-MM-DD.
	Date   string
	Orders int
}

// Stats is how a venue's orders went over its last Days days, in its own
// time zone. Waits run from the scan or join to Ready.
type Stats struct {
	Days        int
	Timezone    string
	Orders      int
	ReadyOrders int
	AvgWait     time.Duration
	MedianWait  time.Duration
	LongestWait time.Duration
	ByHour      [24]int
	ByDay       []DayCount
}

// StatsDays is the period asked for, held to 1 to MaxStatsDays.
func StatsDays(requested int) int {
	switch {
	case requested <= 0:
		return DefaultStatsDays
	case requested > MaxStatsDays:
		return MaxStatsDays
	}
	return requested
}

// StatsSince is the start of the oldest local day in a period ending today.
func StatsSince(now time.Time, loc *time.Location, days int) time.Time {
	local := now.In(loc)
	today := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
	return today.AddDate(0, 0, -(days - 1))
}

// BuildStats counts the tickets of a period. Tickets outside it are ignored.
func BuildStats(tickets []TicketTimes, now time.Time, loc *time.Location, days int) Stats {
	since := StatsSince(now, loc, days)
	stats := Stats{Days: days, Timezone: loc.String(), ByDay: make([]DayCount, days)}
	index := make(map[string]int, days)
	for i := range days {
		date := since.AddDate(0, 0, i).Format(time.DateOnly)
		stats.ByDay[i] = DayCount{Date: date}
		index[date] = i
	}

	var waits []time.Duration
	for _, ticket := range tickets {
		local := ticket.CreatedAt.In(loc)
		day, ok := index[local.Format(time.DateOnly)]
		if !ok {
			continue
		}
		stats.Orders++
		stats.ByDay[day].Orders++
		stats.ByHour[local.Hour()]++
		if wait := ticket.ReadyAt.Sub(ticket.CreatedAt); !ticket.ReadyAt.IsZero() && wait >= 0 {
			waits = append(waits, wait)
		}
	}

	stats.ReadyOrders = len(waits)
	if len(waits) == 0 {
		return stats
	}
	slices.Sort(waits)
	var total time.Duration
	for _, wait := range waits {
		total += wait
	}
	stats.AvgWait = total / time.Duration(len(waits))
	stats.LongestWait = waits[len(waits)-1]
	middle := len(waits) / 2
	stats.MedianWait = waits[middle]
	if len(waits)%2 == 0 {
		stats.MedianWait = (waits[middle-1] + waits[middle]) / 2
	}
	return stats
}
