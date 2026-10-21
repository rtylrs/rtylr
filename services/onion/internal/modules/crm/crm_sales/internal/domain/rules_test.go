package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNextLeadStatusFollowsTheWork(t *testing.T) {
	cases := []struct {
		name     string
		current  LeadStatus
		activity ActivityType
		outcome  CallOutcome
		want     LeadStatus
	}{
		{"an unanswered first call is an attempt", LeadNew, ActivityCall, CallNoAnswer, LeadAttempting},
		{"voicemail is an attempt", LeadNew, ActivityCall, CallVoicemail, LeadAttempting},
		{"a wrong number is not an attempt at this person", LeadNew, ActivityCall, CallWrongNumber, LeadNew},
		{"reaching the person connects", LeadAttempting, ActivityCall, CallConnected, LeadConnected},
		{"a booked meeting connects", LeadNew, ActivityCall, CallMeetingBooked, LeadConnected},
		{"not interested still reached them", LeadAttempting, ActivityCall, CallNotInterested, LeadConnected},
		{"a later unanswered call keeps a connected lead connected", LeadConnected, ActivityCall, CallNoAnswer, LeadConnected},
		{"a meeting connects a nurtured lead", LeadNurture, ActivityMeeting, "", LeadConnected},
		{"a first email is an attempt", LeadNew, ActivityEmail, "", LeadAttempting},
		{"notes change nothing", LeadNew, ActivityNote, "", LeadNew},
		{"qualified leads are left alone", LeadQualified, ActivityCall, CallConnected, LeadQualified},
		{"disqualified leads are left alone", LeadDisqualified, ActivityCall, CallConnected, LeadDisqualified},
	}
	for _, tc := range cases {
		if got := NextLeadStatus(tc.current, tc.activity, tc.outcome); got != tc.want {
			t.Errorf("%s: NextLeadStatus(%s, %s, %s) = %s, want %s", tc.name, tc.current, tc.activity, tc.outcome, got, tc.want)
		}
	}
}

func TestLeadInputNormalize(t *testing.T) {
	in := LeadInput{FirstName: "  Jane ", Email: " Jane@Example.COM ", Phone: "+961 (3) 123-456"}
	if err := in.Normalize(); err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	if in.FirstName != "Jane" || in.Email != "jane@example.com" || in.Billing != BillingOneTime {
		t.Fatalf("normalized = %+v", in)
	}
	if got := PhoneDigits(in.Phone); got != "9613123456" {
		t.Fatalf("PhoneDigits = %q", got)
	}

	for name, bad := range map[string]LeadInput{
		"nobody":    {Email: "x@y.z"},
		"bad email": {CompanyName: "Acme", Email: "not-an-email"},
		"bad phone": {CompanyName: "Acme", Phone: "call me"},
		"bad bill":  {CompanyName: "Acme", Billing: "weekly"},
	} {
		if err := bad.Normalize(); !IsValidation(err) {
			t.Errorf("%s: Normalize err = %v, want a validation error", name, err)
		}
	}
}

func TestLeadChangesStatusRules(t *testing.T) {
	base := LeadInput{CompanyName: "Acme"}
	if err := (&LeadChanges{LeadInput: base, Status: LeadQualified}).Normalize(LeadConnected); !IsValidation(err) {
		t.Errorf("qualifying by edit: err = %v, want validation error", err)
	}
	if err := (&LeadChanges{LeadInput: base, Status: LeadNew}).Normalize(LeadQualified); !errors.Is(err, ErrLeadConverted) {
		t.Errorf("un-qualifying: err = %v, want ErrLeadConverted", err)
	}
	if err := (&LeadChanges{LeadInput: base, Status: LeadDisqualified}).Normalize(LeadNew); !IsValidation(err) {
		t.Errorf("disqualifying without a reason: err = %v, want validation error", err)
	}
	keep := LeadChanges{LeadInput: base, DisqualifyReason: "stale"}
	if err := keep.Normalize(LeadAttempting); err != nil || keep.Status != LeadAttempting || keep.DisqualifyReason != "" {
		t.Errorf("an edit without a status keeps it and drops the reason: %+v, %v", keep, err)
	}
}

func TestPipelineInputNeedsWaysToClose(t *testing.T) {
	valid := PipelineInput{Name: "Sales", Stages: DefaultStages()}
	if err := valid.Normalize(); err != nil {
		t.Fatalf("default stages: %v", err)
	}
	noLost := PipelineInput{Name: "Sales", Stages: []StageInput{
		{Name: "Qualified", Kind: StageOpen}, {Name: "Won", Kind: StageWon},
	}}
	if err := noLost.Normalize(); !IsValidation(err) {
		t.Errorf("no lost stage: err = %v, want validation error", err)
	}
	pinned := PipelineInput{Name: "Sales", Stages: []StageInput{
		{Name: "Open", Kind: StageOpen, Probability: 30}, {Name: "Won", Kind: StageWon, Probability: 10}, {Name: "Lost", Kind: StageLost, Probability: 90},
	}}
	if err := pinned.Normalize(); err != nil || pinned.Stages[1].Probability != 100 || pinned.Stages[2].Probability != 0 {
		t.Errorf("won and lost probabilities are fixed: %+v, %v", pinned.Stages, err)
	}
}

func TestFollowUpWindowRange(t *testing.T) {
	today := time.Date(2026, 10, 20, 15, 0, 0, 0, time.UTC)
	day := func(d int) time.Time { return time.Date(2026, 10, d, 0, 0, 0, 0, time.UTC) }
	cases := map[FollowUpWindow][2]*time.Time{}
	yesterday, todayStart, tomorrow := day(19), day(20), day(21)
	cases[FollowUpOverdue] = [2]*time.Time{nil, &yesterday}
	cases[FollowUpToday] = [2]*time.Time{&todayStart, &todayStart}
	cases[FollowUpDue] = [2]*time.Time{nil, &todayStart}
	cases[FollowUpUpcoming] = [2]*time.Time{&tomorrow, nil}
	for window, want := range cases {
		from, to := window.Range(today)
		if !sameDay(from, want[0]) || !sameDay(to, want[1]) {
			t.Errorf("%s: Range = (%v, %v), want (%v, %v)", window, from, to, want[0], want[1])
		}
	}
}

func sameDay(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Equal(*b)
}

func TestCallerDayUsesTheCallersCalendar(t *testing.T) {
	// 22:30 UTC on the 20th is already the 21st in Beirut (UTC+3).
	now := time.Date(2026, 10, 20, 22, 30, 0, 0, time.UTC)
	day := NewCallerDay("", 180, now)
	if got := day.Date.Format(time.DateOnly); got != "2026-10-21" {
		t.Fatalf("Date = %s, want 2026-10-21", got)
	}
	if want := time.Date(2026, 10, 20, 21, 0, 0, 0, time.UTC); !day.StartUTC.Equal(want) {
		t.Fatalf("StartUTC = %s, want %s", day.StartUTC, want)
	}
	if want := time.Date(2026, 9, 30, 21, 0, 0, 0, time.UTC); !day.MonthStartUTC().Equal(want) {
		t.Fatalf("MonthStartUTC = %s, want %s", day.MonthStartUTC(), want)
	}
	explicit := NewCallerDay("2026-01-05", 0, now)
	if got := explicit.Date.Format(time.DateOnly); got != "2026-01-05" {
		t.Fatalf("explicit Date = %s", got)
	}
}

func TestFirstYearAmount(t *testing.T) {
	if got := FirstYearAmount(10000, BillingMonthly); got != 120000 {
		t.Errorf("monthly = %d", got)
	}
	if got := FirstYearAmount(10000, BillingYearly); got != 10000 {
		t.Errorf("yearly = %d", got)
	}
	if got := FirstYearAmount(10000, BillingOneTime); got != 10000 {
		t.Errorf("one-time = %d", got)
	}
}
