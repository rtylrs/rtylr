package infrastructure_test

import (
	"context"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/application"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/platform/money"
	"github.com/voxire/onion/internal/testutil"
)

// TestSalesActivityCountsCallerDaysAndReps works in the second fixture tenant
// and compares the activity before and after its own writes, so other rows
// in the shared database can't move the numbers it checks.
func TestSalesActivityCountsCallerDaysAndReps(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	// Lead creation takes the database's clock, so the service runs on the
	// real one too and every timestamp lands on today's side of the window.
	now := time.Now().UTC().Truncate(time.Second)
	svc := application.NewService(ctx, configuredRepository(db)).WithClock(func() time.Time { return now })
	const biz, rep = testutil.OtherBusinessID, testutil.OtherOwnerID
	const offset = 180 // the caller is three hours ahead of UTC
	day := domain.NewCallerDay("", offset, now)
	local := func(at time.Time) string { return at.Add(offset * time.Minute).Format(time.DateOnly) }

	activity := func(rep *int64) domain.SalesActivity {
		t.Helper()
		got, err := svc.SalesActivity(ctx, biz, rep, 7, day)
		if err != nil {
			t.Fatalf("SalesActivity: %v", err)
		}
		if len(got.Days) != 7 || got.Days[6].Date.Format(time.DateOnly) != day.Date.Format(time.DateOnly) {
			t.Fatalf("days = %d ending %v, want 7 ending today", len(got.Days), got.Days[len(got.Days)-1].Date)
		}
		return got
	}
	byDate := func(a domain.SalesActivity) map[string]domain.ActivityDay {
		out := map[string]domain.ActivityDay{}
		for _, d := range a.Days {
			out[d.Date.Format(time.DateOnly)] = d
		}
		return out
	}
	share := func(a domain.SalesActivity, id int64) domain.RepActivity {
		for _, r := range a.Reps {
			if r.RepID == id {
				return r
			}
		}
		return domain.RepActivity{RepID: id}
	}
	before := activity(nil)

	newLead := func(name string, value money.Amount) domain.Lead {
		t.Helper()
		lead, err := svc.CreateLead(ctx, biz, rep, domain.LeadInput{FirstName: name, CompanyName: unique(t, name), Value: &value})
		if err != nil {
			t.Fatalf("CreateLead: %v", err)
		}
		return lead
	}
	lead := newLead("Activity", 1000)
	call := func(outcome domain.CallOutcome, at time.Time) {
		t.Helper()
		if _, _, err := svc.LogActivity(ctx, biz, rep, domain.ActivityInput{
			LeadID: &lead.ID, Type: domain.ActivityCall, Outcome: outcome, OccurredAt: &at,
		}); err != nil {
			t.Fatalf("LogActivity: %v", err)
		}
	}
	// 00:30 yesterday for the caller is still the day before in UTC.
	shifted := day.StartOf(day.Date.AddDate(0, 0, -1)).Add(30 * time.Minute)
	threeDaysAgo := now.AddDate(0, 0, -3)
	call(domain.CallNoAnswer, shifted)
	call(domain.CallConnected, shifted.Add(time.Hour))
	call(domain.CallMeetingBooked, threeDaysAgo)
	call(domain.CallNoAnswer, now.AddDate(0, 0, -10)) // before the window

	won := newLead("Won", 2500)
	_, wonDeal, err := svc.ConvertLead(ctx, biz, rep, won.ID, domain.ConvertLeadInput{CreateDeal: true})
	if err != nil {
		t.Fatalf("ConvertLead: %v", err)
	}
	if _, _, err := svc.WinDeal(ctx, biz, rep, wonDeal.ID, domain.WinDealInput{}); err != nil {
		t.Fatalf("WinDeal: %v", err)
	}
	lost := newLead("Lost", 900)
	_, lostDeal, err := svc.ConvertLead(ctx, biz, rep, lost.ID, domain.ConvertLeadInput{CreateDeal: true})
	if err != nil {
		t.Fatalf("ConvertLead: %v", err)
	}
	if _, err := svc.LoseDeal(ctx, biz, rep, lostDeal.ID, "Went with a competitor"); err != nil {
		t.Fatalf("LoseDeal: %v", err)
	}

	after := activity(nil)
	b, a := byDate(before), byDate(after)
	yesterday := local(shifted)
	if yesterday == shifted.Format(time.DateOnly) {
		t.Fatalf("fixture: %v should fall on different days in UTC and for the caller", shifted)
	}
	if got := a[yesterday].Calls - b[yesterday].Calls; got != 2 {
		t.Errorf("calls on the caller's yesterday (%s) grew by %d, want 2", yesterday, got)
	}
	if got := a[yesterday].Reached - b[yesterday].Reached; got != 1 {
		t.Errorf("reached calls on %s grew by %d, want 1", yesterday, got)
	}
	if got := a[local(threeDaysAgo)].Reached - b[local(threeDaysAgo)].Reached; got != 1 {
		t.Errorf("a booked meeting counts as reached: grew by %d, want 1", got)
	}
	if got := after.Calls - before.Calls; got != 3 {
		t.Errorf("window calls grew by %d, want 3 (the 10-day-old call is outside)", got)
	}
	created := local(lead.CreatedAt)
	if got := a[created].LeadsAdded - b[created].LeadsAdded; got != 3 {
		t.Errorf("leads added on %s grew by %d, want 3", created, got)
	}
	if got := a[local(now)].DealsWon - b[local(now)].DealsWon; got != 1 {
		t.Errorf("deals won today grew by %d, want 1", got)
	}
	if after.DealsWon-before.DealsWon != 1 || after.DealsLost-before.DealsLost != 1 ||
		after.LeadsQualified-before.LeadsQualified != 2 || after.WonAmount-before.WonAmount != 2500 {
		t.Errorf("totals moved by won %d, lost %d, qualified %d, amount %d; want 1, 1, 2, 2500",
			after.DealsWon-before.DealsWon, after.DealsLost-before.DealsLost,
			after.LeadsQualified-before.LeadsQualified, after.WonAmount-before.WonAmount)
	}

	// Calls belong to whoever logged them; the deals and leads are unassigned.
	if got := share(after, rep).Calls - share(before, rep).Calls; got != 3 {
		t.Errorf("rep %d calls grew by %d, want 3", rep, got)
	}
	unassignedBefore, unassignedAfter := share(before, 0), share(after, 0)
	if unassignedAfter.DealsWon-unassignedBefore.DealsWon != 1 || unassignedAfter.LeadsQualified-unassignedBefore.LeadsQualified != 2 {
		t.Errorf("unassigned share moved by won %d, qualified %d; want 1, 2",
			unassignedAfter.DealsWon-unassignedBefore.DealsWon, unassignedAfter.LeadsQualified-unassignedBefore.LeadsQualified)
	}

	// One rep's view leaves out the unassigned work.
	mine := activity(ptr(rep))
	if mine.DealsWon != 0 && share(mine, 0).DealsWon != 0 {
		t.Errorf("rep filter kept unassigned deals: %+v", mine)
	}
	if got := mine.Calls; got < 3 {
		t.Errorf("rep filter calls = %d, want at least this test's 3", got)
	}
	for _, r := range mine.Reps {
		if r.RepID != rep {
			t.Errorf("rep filter returned rep %d", r.RepID)
		}
	}

	// The window is 7, 30 or 90 days; anything else is 30.
	if got, err := svc.SalesActivity(ctx, biz, nil, 12, day); err != nil || len(got.Days) != 30 {
		t.Errorf("12-day window: %d days, err %v; want 30", len(got.Days), err)
	}
}
