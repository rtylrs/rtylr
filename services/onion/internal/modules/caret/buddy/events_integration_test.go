package buddy

import (
	"context"
	"sync"
	"testing"

	supportcontract "github.com/voxire/onion/internal/modules/support/contract"
	"github.com/voxire/onion/internal/platform/eventbus"
	"github.com/voxire/onion/internal/testutil"
)

func TestEventDispatchCoalescesRetriesAndRespectsCap(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	const business = int64(920599)
	clean := func() {
		for _, q := range []string{"DELETE FROM caret_buddy_events WHERE buddy_id IN (SELECT id FROM caret_buddies WHERE business_id = ?)", "DELETE FROM caret_buddy_runs WHERE business_id = ?", "DELETE FROM caret_buddies WHERE business_id = ?"} {
			if _, err := db.Exec(q, business); err != nil {
				t.Fatal(err)
			}
		}
	}
	clean()
	t.Cleanup(clean)
	m := NewModule(ctx, db)
	if err := m.Save(ctx, Settings{BusinessID: business, RunAsUserID: 77, Enabled: true, IntervalMinutes: 15, DailyRunCap: 2}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("UPDATE caret_buddies SET next_run_at = NULL WHERE business_id = ?", business); err != nil {
		t.Fatal(err)
	}
	count := func() int {
		var n int
		if err := db.QueryRow("SELECT COUNT(*) FROM caret_buddy_runs WHERE business_id = ?", business).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	event := eventbus.Event{ID: 990001, BusinessID: business, Type: supportcontract.TicketCreated}
	t.Setenv(EnvEnabled, "false")
	if err := m.DispatchEvent(ctx, event); err != nil {
		t.Fatal(err)
	}
	if count() != 0 {
		t.Fatal("disabled feature queued a run")
	}
	t.Setenv(EnvEnabled, "true")
	var wg sync.WaitGroup
	errs := make(chan error, 20)
	for i := range 20 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs <- m.DispatchEvent(ctx, eventbus.Event{ID: uint64(990001 + i), BusinessID: business, Type: supportcontract.CustomerMessageCreated})
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if n := count(); n != 1 {
		t.Fatalf("burst queued %d runs", n)
	}
	if _, err := db.Exec("UPDATE caret_buddy_runs SET status = 'done' WHERE business_id = ?", business); err != nil {
		t.Fatal(err)
	}
	if err := m.DispatchEvent(ctx, event); err != nil {
		t.Fatal(err)
	}
	if n := count(); n != 1 {
		t.Fatalf("retry queued %d runs", n)
	}
	if err := m.DispatchEvent(ctx, eventbus.Event{ID: 990100, BusinessID: business, Type: supportcontract.TicketCreated}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("UPDATE caret_buddy_runs SET status = 'done' WHERE business_id = ?", business); err != nil {
		t.Fatal(err)
	}
	if err := m.DispatchEvent(ctx, eventbus.Event{ID: 990101, BusinessID: business, Type: supportcontract.CustomerMessageCreated}); err != nil {
		t.Fatal(err)
	}
	if n := count(); n != 2 {
		t.Fatalf("cap allowed %d runs", n)
	}
	if err := m.Save(ctx, Settings{BusinessID: business, RunAsUserID: 77, Enabled: false, IntervalMinutes: 15, DailyRunCap: 10}); err != nil {
		t.Fatal(err)
	}
	if err := m.DispatchEvent(ctx, eventbus.Event{ID: 990102, BusinessID: business, Type: supportcontract.TicketCreated}); err != nil {
		t.Fatal(err)
	}
	if n := count(); n != 2 {
		t.Fatalf("disabled buddy queued %d runs", n)
	}
}
