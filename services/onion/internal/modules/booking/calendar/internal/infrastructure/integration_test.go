package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/booking/calendar/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// Far enough in the future that no other fixture in the shared test database
// lands in the same window.
const (
	dateOnlyRangeStart = "2045-06-01"
	dateOnlyRangeEnd   = "2045-06-08"
)

// TestFetchAppointments_DateOnlyRange is the regression test for the date-only
// range: bookingCalendarToolRange and the REST route both send YYYY-MM-DD, and
// the repository used to reject every such value, so booking.get_calendar
// always failed.
func TestFetchAppointments_DateOnlyRange(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	insert := func(start, end string) int64 {
		t.Helper()
		res, err := db.ExecContext(ctx, `
INSERT INTO booking_appointments (owner_id, business_id, customer_name, start_time, end_time, status)
VALUES (?, ?, 'Calendar Range Test', ?, ?, 'confirmed')`,
			testutil.OwnerID, testutil.BusinessID, start, end)
		if err != nil {
			t.Fatalf("insert appointment %s: %v", start, err)
		}
		id, err := res.LastInsertId()
		if err != nil {
			t.Fatalf("appointment id: %v", err)
		}
		return id
	}

	onRangeStart := insert("2045-06-02 09:00:00", "2045-06-02 10:00:00")
	onRangeEnd := insert("2045-06-08 23:30:00", "2045-06-08 23:45:00")
	justOutside := insert("2045-06-09 00:00:00", "2045-06-09 01:00:00")
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(),
			"DELETE FROM booking_appointments WHERE id IN (?, ?, ?)",
			onRangeStart, onRangeEnd, justOutside)
	})

	appts, err := repo.FetchAppointments(ctx, testutil.BusinessID, dateOnlyRangeStart, dateOnlyRangeEnd)
	if err != nil {
		t.Fatalf("FetchAppointments with a date-only range: %v", err)
	}

	got := make(map[int64]bool, len(appts))
	for _, a := range appts {
		got[a.ID] = true
	}
	if !got[onRangeStart] {
		t.Errorf("appointment %d on %s is missing from the range", onRangeStart, dateOnlyRangeStart)
	}
	if !got[onRangeEnd] {
		t.Errorf("appointment %d on the inclusive end day is missing from the range", onRangeEnd)
	}
	if got[justOutside] {
		t.Errorf("appointment %d after the end day leaked into the range", justOutside)
	}
}

// A malformed value must still be rejected rather than silently widening the
// window to everything.
func TestFetchAppointments_RejectsInvalidRange(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	if _, err := repo.FetchAppointments(context.Background(), testutil.BusinessID, "not-a-date", dateOnlyRangeEnd); err == nil {
		t.Fatal("expected an error for an invalid from value, got nil")
	}
}
