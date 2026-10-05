package domain

import "time"

// SeatLedgerPlan describes the writes one reconciliation must make to the seat
// ledger. Keeping the decision pure (no clock, no database) is what makes the
// billing policy reviewable and testable on its own.
type SeatLedgerPlan struct {
	// CloseNoDay is a duplicate open period and the day before it started, so
	// it covers no day and bills nothing. An earlier version of this module
	// could leave two open periods for one subscription, which double-bills
	// every day they overlap.
	CloseNoDay []SeatPeriodClose
	// CloseActive/CloseRunning end the period currently being billed. CloseAt
	// is yesterday, or the last day of that period's own billing month,
	// whichever is earlier, so a rollover never leaves a day unattributed.
	CloseActive  bool
	CloseRunning time.Time
	// OpenNew opens a period for the current billing month at the given start.
	OpenNew     bool
	NewStart    time.Time
	NewBillingM time.Time
	RunningID   int64
}

// SeatPeriodClose is a period id and the day it stops billing.
type SeatPeriodClose struct {
	ID  int64
	End time.Time
}

// PlanSeatLedger decides how to advance a subscription's seat periods.
//
// A period covers inclusive UTC calendar days [PeriodStart, PeriodEnd] and is
// attributed to BillingMonth, so a new period is opened when any of these
// changes: the billing month rolled over, the seat count moved, the plan's
// per-seat price moved, or there is no open period at all. The bill for the
// month is then the sum of the periods, each prorated by day.
func PlanSeatLedger(open []SeatPeriod, today time.Time, seatCount, pricePerSeatCents int) SeatLedgerPlan {
	today = UTCDay(today)
	month := FirstOfUTCMonth(today)

	plan := SeatLedgerPlan{NewStart: today, NewBillingM: month}
	if len(open) == 0 {
		plan.OpenNew = true
		return plan
	}

	// Keep the earliest open period: it covers the most days, so collapsing the
	// rest into zero-length periods loses the least.
	keep := open[0]
	for _, p := range open[1:] {
		if UTCDay(p.PeriodStart).Before(UTCDay(keep.PeriodStart)) {
			keep = p
		}
	}
	plan.RunningID = keep.ID

	for _, p := range open {
		if p.ID == keep.ID {
			continue
		}
		plan.CloseNoDay = append(plan.CloseNoDay, SeatPeriodClose{
			ID:  p.ID,
			End: UTCDay(p.PeriodStart).AddDate(0, 0, -1),
		})
	}

	stale := !UTCDay(keep.BillingMonth).Equal(month) ||
		keep.SeatCount != seatCount ||
		keep.PricePerSeatCents != pricePerSeatCents ||
		UTCDay(keep.PeriodStart).After(today)
	if !stale {
		return plan
	}

	closeAt := today.AddDate(0, 0, -1)
	if monthEnd := LastOfUTCMonth(keep.BillingMonth); monthEnd.Before(closeAt) {
		closeAt = monthEnd
	}
	plan.CloseActive = true
	plan.CloseRunning = closeAt
	plan.OpenNew = true
	return plan
}
