package domain

import (
	"testing"
	"time"
)

func day(year int, month time.Month, d int) time.Time {
	return time.Date(year, month, d, 0, 0, 0, 0, time.UTC)
}

func openPeriod(id int64, start, billingMonth time.Time, seats, price int) SeatPeriod {
	return SeatPeriod{ID: id, SeatCount: seats, PricePerSeatCents: price, PeriodStart: start, BillingMonth: billingMonth}
}

// A brand-new subscription has no period and must get one for the current month.
func TestPlanSeatLedger_NoOpenPeriodOpensOne(t *testing.T) {
	today := day(2026, time.April, 10)
	plan := PlanSeatLedger(nil, today, 12, 1000)

	if !plan.OpenNew {
		t.Fatal("expected a new period to be opened")
	}
	if !plan.NewStart.Equal(day(2026, time.April, 10)) {
		t.Errorf("new period starts %v, want 2026-04-10", plan.NewStart)
	}
	if !plan.NewBillingM.Equal(day(2026, time.April, 1)) {
		t.Errorf("billing month %v, want 2026-04-01", plan.NewBillingM)
	}
	if plan.CloseActive || len(plan.CloseNoDay) != 0 {
		t.Errorf("nothing to close, got %+v", plan)
	}
}

// The bug: a period from a previous month was left open forever, so the new
// billing month had no rows and billed zero.
func TestPlanSeatLedger_MonthRolloverOpensNewPeriod(t *testing.T) {
	today := day(2026, time.May, 1)
	open := []SeatPeriod{openPeriod(1, day(2026, time.April, 1), day(2026, time.April, 1), 12, 1000)}

	plan := PlanSeatLedger(open, today, 12, 1000)

	if !plan.OpenNew {
		t.Fatal("expected a period for the new billing month")
	}
	if !plan.NewBillingM.Equal(day(2026, time.May, 1)) {
		t.Errorf("billing month %v, want 2026-05-01", plan.NewBillingM)
	}
	if !plan.NewStart.Equal(day(2026, time.May, 1)) {
		t.Errorf("new period starts %v, want 2026-05-01", plan.NewStart)
	}
	if !plan.CloseActive || !plan.CloseRunning.Equal(day(2026, time.April, 30)) {
		t.Errorf("april must close on its own last day, got close=%v at %v", plan.CloseActive, plan.CloseRunning)
	}
}

// Seats added mid-month are prorated: the old count bills up to the day before
// the change and the new count bills from the change day, so no day is billed
// twice and none is skipped.
func TestPlanSeatLedger_SeatsAddedMidMonth(t *testing.T) {
	today := day(2026, time.April, 16)
	open := []SeatPeriod{openPeriod(1, day(2026, time.April, 1), day(2026, time.April, 1), 3, 1000)}

	plan := PlanSeatLedger(open, today, 20, 1000)

	if !plan.CloseActive || !plan.CloseRunning.Equal(day(2026, time.April, 15)) {
		t.Errorf("old period must close 2026-04-15, got close=%v at %v", plan.CloseActive, plan.CloseRunning)
	}
	if !plan.OpenNew || !plan.NewStart.Equal(day(2026, time.April, 16)) {
		t.Errorf("new period must start 2026-04-16, got open=%v at %v", plan.OpenNew, plan.NewStart)
	}
	if !plan.NewBillingM.Equal(day(2026, time.April, 1)) {
		t.Errorf("new period stays in the same billing month, got %v", plan.NewBillingM)
	}
}

// Seats removed mid-month are prorated the same way, in the merchant's favour:
// the reduced count applies from the change day, and no credit is issued for
// the days already billed.
func TestPlanSeatLedger_SeatsRemovedMidMonth(t *testing.T) {
	today := day(2026, time.April, 16)
	open := []SeatPeriod{openPeriod(1, day(2026, time.April, 1), day(2026, time.April, 1), 20, 1000)}

	plan := PlanSeatLedger(open, today, 3, 1000)

	if !plan.CloseActive || !plan.CloseRunning.Equal(day(2026, time.April, 15)) {
		t.Errorf("old period must close 2026-04-15, got close=%v at %v", plan.CloseActive, plan.CloseRunning)
	}
	if !plan.OpenNew || !plan.NewStart.Equal(day(2026, time.April, 16)) {
		t.Errorf("new period must start 2026-04-16, got open=%v at %v", plan.OpenNew, plan.NewStart)
	}
}

// A plan change moves the per-seat price, so the running period has to be split
// even though the seat count is identical.
func TestPlanSeatLedger_PriceChangeSplitsPeriod(t *testing.T) {
	today := day(2026, time.April, 20)
	open := []SeatPeriod{openPeriod(1, day(2026, time.April, 1), day(2026, time.April, 1), 12, 699)}

	plan := PlanSeatLedger(open, today, 12, 499)

	if !plan.OpenNew {
		t.Fatal("a price change must open a new period")
	}
	if !plan.CloseRunning.Equal(day(2026, time.April, 19)) {
		t.Errorf("old price bills through 2026-04-19, got %v", plan.CloseRunning)
	}
}

// Nothing changed, so nothing is written. A reconcile job that runs every
// fifteen minutes must not churn a period per run.
func TestPlanSeatLedger_UnchangedIsANoOp(t *testing.T) {
	today := day(2026, time.April, 20)
	open := []SeatPeriod{openPeriod(1, day(2026, time.April, 1), day(2026, time.April, 1), 12, 1000)}

	plan := PlanSeatLedger(open, today, 12, 1000)

	if plan.OpenNew || plan.CloseActive || len(plan.CloseNoDay) != 0 {
		t.Fatalf("expected no writes, got %+v", plan)
	}
	if plan.RunningID != 1 {
		t.Errorf("running period id = %d, want 1", plan.RunningID)
	}
}

// Two open periods is the damage the old unlocked read path could leave behind.
// The later duplicate is collapsed to zero days so the day is not billed twice.
func TestPlanSeatLedger_CollapsesDuplicateOpenPeriods(t *testing.T) {
	today := day(2026, time.April, 20)
	open := []SeatPeriod{
		openPeriod(1, day(2026, time.April, 1), day(2026, time.April, 1), 12, 1000),
		openPeriod(2, day(2026, time.April, 1), day(2026, time.April, 1), 12, 1000),
	}

	plan := PlanSeatLedger(open, today, 12, 1000)

	if plan.RunningID != 1 {
		t.Errorf("must keep the earliest period, got %d", plan.RunningID)
	}
	if len(plan.CloseNoDay) != 1 || plan.CloseNoDay[0].ID != 2 {
		t.Fatalf("expected only period 2 collapsed, got %+v", plan.CloseNoDay)
	}
	if !plan.CloseNoDay[0].End.Equal(day(2026, time.March, 31)) {
		t.Errorf("collapsed period must cover no day (closed the day before it started), got %v", plan.CloseNoDay[0].End)
	}
	if plan.OpenNew {
		t.Error("state is unchanged, so no replacement period is needed")
	}
}

// A reconciliation that runs every day must bill each day exactly once.
func TestPlanSeatLedger_RepeatedReconcileBillsEachDayOnce(t *testing.T) {
	periods := []SeatPeriod{}
	for dayOfMonth := 1; dayOfMonth <= 30; dayOfMonth++ {
		today := day(2026, time.April, dayOfMonth)
		open := make([]SeatPeriod, 0, len(periods))
		for _, p := range periods {
			if p.PeriodEnd == nil {
				open = append(open, p)
			}
		}
		plan := PlanSeatLedger(open, today, 12, 1000)
		for _, c := range plan.CloseNoDay {
			for i := range periods {
				if periods[i].ID == c.ID {
					end := c.End
					periods[i].PeriodEnd = &end
				}
			}
		}
		if plan.CloseActive {
			for i := range periods {
				if periods[i].ID == plan.RunningID {
					end := plan.CloseRunning
					periods[i].PeriodEnd = &end
				}
			}
		}
		if plan.OpenNew {
			periods = append(periods, SeatPeriod{
				ID: int64(len(periods) + 1), SeatCount: 12, PricePerSeatCents: 1000,
				PeriodStart: plan.NewStart, BillingMonth: plan.NewBillingM,
			})
		}
	}

	covered := map[time.Time]int{}
	for _, p := range periods {
		end := day(2026, time.April, 30)
		if p.PeriodEnd != nil {
			end = *p.PeriodEnd
		}
		for d := p.PeriodStart; !d.After(end); d = d.AddDate(0, 0, 1) {
			covered[d]++
		}
	}
	if len(periods) != 1 {
		t.Fatalf("30 reconciles of an unchanged subscription produced %d periods, want 1", len(periods))
	}
	for d, n := range covered {
		if n != 1 {
			t.Errorf("%v is covered by %d periods, want exactly 1", d, n)
		}
	}
	if len(covered) != 30 {
		t.Errorf("covered %d days of April, want 30", len(covered))
	}
}

func TestPlanSeatLedger_LeapFebruary(t *testing.T) {
	today := day(2028, time.March, 1)
	open := []SeatPeriod{openPeriod(1, day(2028, time.February, 1), day(2028, time.February, 1), 12, 1000)}

	plan := PlanSeatLedger(open, today, 12, 1000)

	if !plan.CloseRunning.Equal(day(2028, time.February, 29)) {
		t.Errorf("leap february must close on the 29th, got %v", plan.CloseRunning)
	}
}

// A full leap-February period must bill exactly one month of per-seat price;
// the reconciliation test above covers the boundary, this one the denominator.
func TestDaysInUTCMonth_LeapFebruary(t *testing.T) {
	if DaysInUTCMonth(day(2028, time.February, 1)) != 29 {
		t.Errorf("leap february has %d days, want 29", DaysInUTCMonth(day(2028, time.February, 1)))
	}
	if DaysInUTCMonth(day(2026, time.February, 1)) != 28 {
		t.Errorf("non-leap february has %d days, want 28", DaysInUTCMonth(day(2026, time.February, 1)))
	}
	if DaysInUTCMonth(day(2026, time.April, 1)) != 30 {
		t.Errorf("april has %d days, want 30", DaysInUTCMonth(day(2026, time.April, 1)))
	}
	if LastOfUTCMonth(day(2028, time.February, 10)) != day(2028, time.February, 29) {
		t.Errorf("last day of leap february = %v", LastOfUTCMonth(day(2028, time.February, 10)))
	}
}

// Boundaries are UTC days, so a period written in a non-UTC zone still lands on
// the same calendar day.
func TestPlanSeatLedger_NonUTCTimeNormalisesToUTCDay(t *testing.T) {
	tokyo := time.FixedZone("JST", 9*3600)
	// 2026-05-01 08:00 in Tokyo is 2026-04-30 23:00 UTC, so the UTC day is the
	// 30th and the new period belongs to April.
	today := time.Date(2026, time.May, 1, 8, 0, 0, 0, tokyo)
	open := []SeatPeriod{openPeriod(1, day(2026, time.April, 1), day(2026, time.April, 1), 12, 1000)}

	plan := PlanSeatLedger(open, today, 12, 1000)

	if !plan.NewStart.Equal(day(2026, time.April, 30)) {
		t.Errorf("new period start %v, want 2026-04-30 UTC", plan.NewStart)
	}
	if !plan.NewBillingM.Equal(day(2026, time.April, 1)) {
		t.Errorf("billing month %v, want 2026-04-01", plan.NewBillingM)
	}
}
