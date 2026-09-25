package application

import (
	"context"
	"sync"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_time/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	"github.com/voxire/onion/internal/testutil"
)

// initHRTimeCache wires the process caches the way an API process does, minus
// Redis: the local cache is real and the no-op remote tier still clears it on
// scope invalidation, so the tests exercise GetOrComputeSingle with a real
// backing store.
var initHRTimeCacheOnce sync.Once

func initHRTimeCache(t *testing.T) context.Context {
	t.Helper()

	ctx := context.Background()
	initHRTimeCacheOnce.Do(func() {
		if err := cacher.InitializeLocalCache(ctx); err != nil {
			t.Fatalf("InitializeLocalCache: %v", err)
		}
		cacher.DisableRedisForProcess(ctx)
	})
	return ctx
}

func hrTimeTestPtr[T any](v T) *T { return &v }

// TestListTimeEntries_DifferentDateFiltersDoNotShareACachedPage is the
// B08-R02 regression: the cache key used to render only limit, offset and
// employee_id with %d, which prints the pointer address (or a %!d formatting
// marker for nil). Two requests with the same limit/offset and no employee_id
// therefore shared one cache entry, so the second filter combination was
// answered from the first page for the whole 2-minute TTL.
func TestListTimeEntries_DifferentDateFiltersDoNotShareACachedPage(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := initHRTimeCache(t)

	const businessID int64 = 9100791
	const employeeID int64 = 4301

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		for _, stmt := range []string{
			"DELETE FROM hr_time_entries WHERE business_id = ?",
			"DELETE FROM hr_timesheets WHERE business_id = ?",
		} {
			if _, err := db.ExecContext(cleanupCtx, stmt, businessID); err != nil {
				t.Logf("cleanup %q for business %d: %v", stmt, businessID, err)
			}
		}
	})

	permissions := newW2PermissionProvider()
	permissions.employeeByUser[employeeID] = employeeID
	service := NewService(ctx, db, nil, permissions)

	// Two October entries and one November entry make the two filter
	// combinations distinguishable by total, not just by item identity.
	for _, date := range []string{"2026-10-05", "2026-10-06", "2026-11-04"} {
		if _, err := service.CreateTimeEntry(ctx, businessID, employeeID, domain.TimeEntryRecord{
			EmployeeID:      employeeID,
			Date:            date,
			DurationMinutes: 60,
			Status:          domain.EntryStatusDraft,
			Source:          domain.EntrySourceManual,
		}); err != nil {
			t.Fatalf("CreateTimeEntry(%s): %v", date, err)
		}
	}

	october, octoberTotal, err := service.ListTimeEntries(ctx, businessID, domain.ListTimeEntriesParams{
		FromDate: hrTimeTestPtr("2026-10-01"),
		ToDate:   hrTimeTestPtr("2026-10-31"),
		Limit:    20,
	})
	if err != nil {
		t.Fatalf("ListTimeEntries(October): %v", err)
	}
	if octoberTotal != 2 || len(october) != 2 {
		t.Fatalf("October page = %d entries / total %d, want 2/2", len(october), octoberTotal)
	}

	november, novemberTotal, err := service.ListTimeEntries(ctx, businessID, domain.ListTimeEntriesParams{
		FromDate: hrTimeTestPtr("2026-11-01"),
		ToDate:   hrTimeTestPtr("2026-11-30"),
		Limit:    20,
	})
	if err != nil {
		t.Fatalf("ListTimeEntries(November): %v", err)
	}
	if novemberTotal != 1 || len(november) != 1 {
		t.Fatalf("November page = %d entries / total %d, want 1/1; filters shared one cached page (B08-R02)", len(november), novemberTotal)
	}
	if november[0].Date != "2026-11-04" {
		t.Errorf("November page returned entry dated %q, want 2026-11-04", november[0].Date)
	}
}

// TestListTimeEntriesCacheKey_DistinguishesEveryFilter pins the key itself: a
// nil filter must not render the same as another combination, which is the
// shape that let unrelated pages collide.
func TestListTimeEntriesCacheKey_DistinguishesEveryFilter(t *testing.T) {
	base := listTimeEntriesCacheKey(1, domain.ListTimeEntriesParams{Limit: 20})

	variants := map[string]domain.ListTimeEntriesParams{
		"from_date":   {Limit: 20, FromDate: hrTimeTestPtr("2026-10-01")},
		"to_date":     {Limit: 20, ToDate: hrTimeTestPtr("2026-10-31")},
		"employee":    {Limit: 20, EmployeeID: hrTimeTestPtr[int64](4301)},
		"project":     {Limit: 20, ProjectID: hrTimeTestPtr[int64](17)},
		"is_billable": {Limit: 20, IsBillable: hrTimeTestPtr(false)},
		"status":      {Limit: 20, Status: hrTimeTestPtr("approved")},
		"limit":       {Limit: 50},
		"offset":      {Limit: 20, Offset: 20},
	}

	for name, params := range variants {
		if got := listTimeEntriesCacheKey(1, params); got == base {
			t.Errorf("key for %s filter collides with the unfiltered key: %q", name, got)
		}
	}

	// Identical filters must still share one key, and two different
	// businesses must never share one.
	same := domain.ListTimeEntriesParams{Limit: 20, FromDate: hrTimeTestPtr("2026-10-01")}
	if listTimeEntriesCacheKey(1, same) != listTimeEntriesCacheKey(1, same) {
		t.Error("identical filters produced different keys")
	}
	if listTimeEntriesCacheKey(1, same) == listTimeEntriesCacheKey(2, same) {
		t.Error("different businesses share one key")
	}
}

func TestListTimesheetsCacheKey_DistinguishesBothFilters(t *testing.T) {
	base := listTimesheetsCacheKey(1, domain.ListTimesheetsParams{Limit: 20})

	if got := listTimesheetsCacheKey(1, domain.ListTimesheetsParams{Limit: 20, Status: hrTimeTestPtr(domain.TimesheetStatusSubmitted)}); got == base {
		t.Errorf("status filter collides with the unfiltered key: %q", got)
	}
	if got := listTimesheetsCacheKey(1, domain.ListTimesheetsParams{Limit: 20, EmployeeID: hrTimeTestPtr[int64](4401)}); got == base {
		t.Errorf("employee filter collides with the unfiltered key: %q", got)
	}
}

// TestListTimesheets_DifferentStatusFiltersDoNotShareACachedPage pins the
// service-level consequence for B44-R01: now that status is a filter it must
// be part of the cache key, or a submitted-only page would answer a draft-only
// query for the whole TTL.
func TestListTimesheets_DifferentStatusFiltersDoNotShareACachedPage(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := initHRTimeCache(t)

	const businessID int64 = 9100796
	const employeeID int64 = 4501

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		for _, stmt := range []string{
			"DELETE FROM hr_timesheet_approval_actions WHERE business_id = ?",
			"DELETE FROM hr_timesheets WHERE business_id = ?",
		} {
			if _, err := db.ExecContext(cleanupCtx, stmt, businessID); err != nil {
				t.Logf("cleanup %q for business %d: %v", stmt, businessID, err)
			}
		}
	})

	service := NewService(ctx, db, nil, nil)

	october, err := service.GetOrCreateTimesheet(ctx, businessID, employeeID, "2026-10-01", "2026-10-31")
	if err != nil {
		t.Fatalf("GetOrCreateTimesheet(October): %v", err)
	}
	submitted, err := service.SubmitTimesheet(ctx, businessID, october.ID)
	if err != nil {
		t.Fatalf("SubmitTimesheet: %v", err)
	}
	november, err := service.GetOrCreateTimesheet(ctx, businessID, employeeID, "2026-11-01", "2026-11-30")
	if err != nil {
		t.Fatalf("GetOrCreateTimesheet(November): %v", err)
	}

	submittedPage, submittedTotal, err := service.ListTimesheets(ctx, businessID, domain.ListTimesheetsParams{
		Status: hrTimeTestPtr(domain.TimesheetStatusSubmitted),
		Limit:  20,
	})
	if err != nil {
		t.Fatalf("ListTimesheets(submitted): %v", err)
	}
	if submittedTotal != 1 || len(submittedPage) != 1 || submittedPage[0].ID != submitted.ID {
		t.Fatalf("submitted page = %d rows / total %d, want just timesheet %d", len(submittedPage), submittedTotal, submitted.ID)
	}

	draftPage, draftTotal, err := service.ListTimesheets(ctx, businessID, domain.ListTimesheetsParams{
		Status: hrTimeTestPtr(domain.TimesheetStatusDraft),
		Limit:  20,
	})
	if err != nil {
		t.Fatalf("ListTimesheets(draft): %v", err)
	}
	if draftTotal != 1 || len(draftPage) != 1 || draftPage[0].ID != november.ID {
		t.Fatalf("draft page = %d rows / total %d, want just timesheet %d; status filters shared one cached page", len(draftPage), draftTotal, november.ID)
	}

	all, allTotal, err := service.ListTimesheets(ctx, businessID, domain.ListTimesheetsParams{Limit: 20})
	if err != nil {
		t.Fatalf("ListTimesheets(all): %v", err)
	}
	if allTotal != 2 || len(all) != 2 {
		t.Errorf("unfiltered page = %d rows / total %d, want 2/2", len(all), allTotal)
	}
}
