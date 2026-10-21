package domain

import (
	"strings"
	"time"

	"github.com/voxire/onion/internal/platform/money"
)

// FollowUpWindow narrows a list by its next follow-up date.
type FollowUpWindow string

const (
	FollowUpAny       FollowUpWindow = ""
	FollowUpOverdue   FollowUpWindow = "overdue"
	FollowUpToday     FollowUpWindow = "today"
	FollowUpDue       FollowUpWindow = "due"
	FollowUpUpcoming  FollowUpWindow = "upcoming"
	FollowUpScheduled FollowUpWindow = "scheduled"
	FollowUpNone      FollowUpWindow = "none"
)

func (w FollowUpWindow) IsValid() bool {
	switch w {
	case FollowUpAny, FollowUpOverdue, FollowUpToday, FollowUpDue, FollowUpUpcoming, FollowUpScheduled, FollowUpNone:
		return true
	}
	return false
}

// earliestDate is before any follow-up, so a range from it means "any date".
var earliestDate = time.Date(1000, 1, 1, 0, 0, 0, 0, time.UTC)

// Range is the window's inclusive date bounds relative to today; nil means
// unbounded. FollowUpNone has no range: it selects leads without a date.
func (w FollowUpWindow) Range(today time.Time) (from, to *time.Time) {
	day := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	yesterday := day.AddDate(0, 0, -1)
	tomorrow := day.AddDate(0, 0, 1)
	switch w {
	case FollowUpOverdue:
		return nil, &yesterday
	case FollowUpToday:
		return &day, &day
	case FollowUpDue:
		return nil, &day
	case FollowUpUpcoming:
		return &tomorrow, nil
	case FollowUpScheduled:
		return &earliestDate, nil
	}
	return nil, nil
}

// Presence filters on whether a field is filled in.
type Presence string

const (
	PresenceAny Presence = ""
	PresenceYes Presence = "yes"
	PresenceNo  Presence = "no"
)

func (p Presence) IsValid() bool {
	return p == PresenceAny || p == PresenceYes || p == PresenceNo
}

// maxFilterValues caps each set filter, which travels in a query string.
const maxFilterValues = 200

// LeadFilter selects leads; the filters combine with AND. Empty sets, nil
// bounds and PresenceAny don't filter, and empty Statuses means every status.
// Reps may hold 0 for unassigned leads. Date bounds are the caller's calendar
// days, inclusive.
type LeadFilter struct {
	Statuses      []LeadStatus
	Reps          []int64
	Sources       []string
	Industries    []string
	Sizes         []string
	Countries     []string
	Cities        []string
	Billing       []Billing
	ImportBatchID *int64
	FollowUp      FollowUpWindow
	HasPhone      Presence
	HasEmail      Presence
	HasWebsite    Presence
	HasLinkedIn   Presence
	HasCompany    Presence
	HasValue      Presence
	// Contacted is whether the lead was ever worked.
	Contacted   Presence
	AttemptsMin *int32
	AttemptsMax *int32
	ValueMin    *money.Amount
	ValueMax    *money.Amount
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	ActiveFrom  *time.Time
	ActiveTo    *time.Time
	Search      string
	Sort        LeadSort
	Cursor      string
	Limit       int
}

// Normalize checks the filter and tidies its sets: blanks and repeats go,
// and each set is capped.
func (f *LeadFilter) Normalize() error {
	if f.Sort == "" {
		f.Sort = LeadSortNewest
	}
	if !f.Sort.IsValid() {
		return invalid("unknown sort %q", f.Sort)
	}
	for _, status := range f.Statuses {
		if !status.IsValid() {
			return invalid("unknown lead status %q", status)
		}
	}
	for _, billing := range f.Billing {
		if !billing.IsValid() {
			return invalid("unknown billing %q", billing)
		}
	}
	if !f.FollowUp.IsValid() {
		return invalid("unknown follow-up window %q", f.FollowUp)
	}
	for _, p := range []Presence{f.HasPhone, f.HasEmail, f.HasWebsite, f.HasLinkedIn, f.HasCompany, f.HasValue, f.Contacted} {
		if !p.IsValid() {
			return invalid("presence filters are yes or no, not %q", p)
		}
	}
	if (f.AttemptsMin != nil && *f.AttemptsMin < 0) || (f.AttemptsMax != nil && *f.AttemptsMax < 0) {
		return invalid("call attempts can't be negative")
	}
	if (f.ValueMin != nil && *f.ValueMin < 0) || (f.ValueMax != nil && *f.ValueMax < 0) {
		return invalid("value can't be negative")
	}
	f.Reps = distinct(f.Reps)
	f.Billing = distinct(f.Billing)
	for _, set := range []*[]string{&f.Sources, &f.Industries, &f.Sizes, &f.Countries, &f.Cities} {
		*set = distinct(trimAll(*set))
	}
	if len(f.Reps) > maxFilterValues || len(f.Sources) > maxFilterValues || len(f.Industries) > maxFilterValues ||
		len(f.Sizes) > maxFilterValues || len(f.Countries) > maxFilterValues || len(f.Cities) > maxFilterValues {
		return invalid("pick at most %d values per filter", maxFilterValues)
	}
	if runes := []rune(f.Search); len(runes) > 100 {
		f.Search = string(runes[:100])
	}
	return nil
}

// trimAll trims each value and drops the blank ones.
func trimAll(values []string) []string {
	out := values[:0:0]
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			out = append(out, value)
		}
	}
	return out
}

// distinct drops repeats, keeping the first of each.
func distinct[T comparable](values []T) []T {
	seen := make(map[T]bool, len(values))
	out := values[:0:0]
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	return out
}
