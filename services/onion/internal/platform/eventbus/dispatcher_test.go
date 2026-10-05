package eventbus_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/automations"
	"github.com/voxire/onion/internal/modules/webhooks"
	"github.com/voxire/onion/internal/platform/eventbus"
	"github.com/voxire/onion/internal/testutil"
)

// Fan-out has three properties worth proving, and all three are the kind that
// look right by inspection and are wrong in practice: only subscribed targets
// are picked up, only targets in the event's own tenant are picked up, and
// running twice does not double-send. The last one matters more here than
// anywhere else in the product, because the actions behind these queues include
// awarding loyalty points and sending email.
func TestDispatchBatchFansOutToSubscribedTargetsOnly(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	const (
		tenant      int64 = 880001
		otherTenant int64 = 880002
		ownerID     int64 = 880000
	)

	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.ExecContext(ctx, query, args...); err != nil {
			t.Fatalf("seed (%s): %v", query, err)
		}
	}

	t.Cleanup(func() {
		for _, table := range []string{"webhook_deliveries", "automation_runs", "platform_events"} {
			_, _ = db.ExecContext(ctx, "DELETE FROM "+table+" WHERE business_id IN (?, ?)", tenant, otherTenant)
		}
		_, _ = db.ExecContext(ctx, "DELETE FROM hub_webhooks WHERE business_id IN (?, ?)", tenant, otherTenant)
		_, _ = db.ExecContext(ctx, "DELETE FROM crm_automations WHERE business_id IN (?, ?)", tenant, otherTenant)
	})

	// The one registration that should receive the event.
	exec(`INSERT INTO hub_webhooks (id, owner_id, business_id, url, events, secret, is_active)
	      VALUES (880011, ?, ?, 'https://example.test/hook', JSON_ARRAY('order_created'), 's', TRUE)`, ownerID, tenant)
	// Subscribed to something else.
	exec(`INSERT INTO hub_webhooks (id, owner_id, business_id, url, events, secret, is_active)
	      VALUES (880012, ?, ?, 'https://example.test/other', JSON_ARRAY('order_paid'), 's', TRUE)`, ownerID, tenant)
	// Right event, switched off.
	exec(`INSERT INTO hub_webhooks (id, owner_id, business_id, url, events, secret, is_active)
	      VALUES (880013, ?, ?, 'https://example.test/off', JSON_ARRAY('order_created'), 's', FALSE)`, ownerID, tenant)
	// Right event, wrong tenant.
	exec(`INSERT INTO hub_webhooks (id, owner_id, business_id, url, events, secret, is_active)
	      VALUES (880014, ?, ?, 'https://example.test/foreign', JSON_ARRAY('order_created'), 's', TRUE)`, ownerID, otherTenant)

	// Seeded with the constant rather than a literal, so a drift between
	// eventbus.TypeOrderCreated and the automations module's TriggerOrderCreated
	// shows up here. They cannot share a constant: the trigger lives under
	// internal/modules/automations/internal, which a platform package may not
	// import.
	exec(`INSERT INTO crm_automations (id, owner_id, business_id, name, trigger_type, is_active)
	      VALUES (880021, ?, ?, 'welcome', ?, TRUE)`, ownerID, tenant, eventbus.TypeOrderCreated)
	exec(`INSERT INTO crm_automations (id, owner_id, business_id, name, trigger_type, is_active)
	      VALUES (880022, ?, ?, 'birthday', 'date_birthday', TRUE)`, ownerID, tenant)
	exec(`INSERT INTO crm_automations (id, owner_id, business_id, name, trigger_type, is_active)
	      VALUES (880023, ?, ?, 'foreign', 'order_created', TRUE)`, ownerID, otherTenant)

	exec(`INSERT INTO platform_events (id, business_id, event_type, subject_type, subject_id, payload, dedupe_key)
	      VALUES (880031, ?, 'order_created', 'order', 501, JSON_OBJECT('order_id', 501), 'order_created:501')`, tenant)

	countRow := func(query string, args ...any) int {
		t.Helper()
		var n int
		if err := db.QueryRowContext(ctx, query, args...).Scan(&n); err != nil {
			t.Fatalf("count (%s): %v", query, err)
		}
		return n
	}

	// Before: the event exists and nothing has been expanded from it.
	if got := countRow(`SELECT COUNT(*) FROM webhook_deliveries WHERE event_id = 880031`); got != 0 {
		t.Fatalf("precondition: %d deliveries before dispatch, want 0", got)
	}
	if got := countRow(`SELECT COUNT(*) FROM automation_runs WHERE event_id = 880031`); got != 0 {
		t.Fatalf("precondition: %d runs before dispatch, want 0", got)
	}

	dispatcher := eventbus.NewDispatcher(
		db,
		automations.NewModule(ctx, db, func(context.Context, int64, int64) (automations.CustomerContact, bool, error) {
			return automations.CustomerContact{}, false, nil
		}),
		webhooks.NewModule(ctx, db),
	)

	dispatched, err := dispatcher.DispatchBatch(ctx)
	if err != nil {
		t.Fatalf("DispatchBatch: %v", err)
	}
	if dispatched < 1 {
		t.Fatalf("expanded %d events, want at least the one seeded", dispatched)
	}

	// Exactly the subscribed, active, same-tenant registration.
	var webhookID int64
	if err := db.QueryRowContext(ctx,
		`SELECT webhook_id FROM webhook_deliveries WHERE event_id = 880031`).Scan(&webhookID); err != nil {
		t.Fatalf("read delivery: %v", err)
	}
	if webhookID != 880011 {
		t.Errorf("delivery went to webhook %d, want 880011", webhookID)
	}
	if got := countRow(`SELECT COUNT(*) FROM webhook_deliveries WHERE event_id = 880031`); got != 1 {
		t.Errorf("%d deliveries, want exactly 1: an unsubscribed, inactive or foreign registration was included", got)
	}

	var automationID int64
	if err := db.QueryRowContext(ctx,
		`SELECT automation_id FROM automation_runs WHERE event_id = 880031`).Scan(&automationID); err != nil {
		t.Fatalf("read run: %v", err)
	}
	if automationID != 880021 {
		t.Errorf("run created for automation %d, want 880021", automationID)
	}
	if got := countRow(`SELECT COUNT(*) FROM automation_runs WHERE event_id = 880031`); got != 1 {
		t.Errorf("%d runs, want exactly 1: a wrong-trigger or foreign automation was included", got)
	}

	// Nothing belonging to the other tenant was touched.
	if got := countRow(`SELECT COUNT(*) FROM webhook_deliveries WHERE business_id = ?`, otherTenant); got != 0 {
		t.Errorf("%d deliveries for the other tenant, want 0", got)
	}
	if got := countRow(`SELECT COUNT(*) FROM automation_runs WHERE business_id = ?`, otherTenant); got != 0 {
		t.Errorf("%d runs for the other tenant, want 0", got)
	}

	// The event is marked, so a second pass does not re-read it.
	if got := countRow(`SELECT COUNT(*) FROM platform_events WHERE id = 880031 AND dispatched_at IS NOT NULL`); got != 1 {
		t.Error("event was not marked dispatched, so it would be re-read on every pass forever")
	}

	// Idempotency. Unmark the event and expand it again: a dispatcher that
	// crashed after writing queue rows but before committing its mark would do
	// exactly this on restart, and it must not double-send.
	if _, err := db.ExecContext(ctx,
		`UPDATE platform_events SET dispatched_at = NULL WHERE id = 880031`); err != nil {
		t.Fatalf("unmark event: %v", err)
	}
	if _, err := dispatcher.DispatchBatch(ctx); err != nil {
		t.Fatalf("second DispatchBatch: %v", err)
	}
	if got := countRow(`SELECT COUNT(*) FROM webhook_deliveries WHERE event_id = 880031`); got != 1 {
		t.Errorf("%d deliveries after re-expanding the same event, want 1", got)
	}
	if got := countRow(`SELECT COUNT(*) FROM automation_runs WHERE event_id = 880031`); got != 1 {
		t.Errorf("%d runs after re-expanding the same event, want 1", got)
	}
}

// An event nobody subscribes to must still be marked, or the dispatcher reads
// it on every pass for the life of the system.
func TestDispatchBatchMarksAnEventWithNoSubscribers(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	const tenant int64 = 880003

	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM platform_events WHERE business_id = ?", tenant)
	})

	if _, err := db.ExecContext(ctx,
		`INSERT INTO platform_events (id, business_id, event_type, subject_type, subject_id, payload, dedupe_key)
		 VALUES (880041, ?, 'order_created', 'order', 777, JSON_OBJECT('order_id', 777), 'order_created:777')`,
		tenant); err != nil {
		t.Fatalf("seed event: %v", err)
	}

	if _, err := eventbus.NewDispatcher(db).DispatchBatch(ctx); err != nil {
		t.Fatalf("DispatchBatch: %v", err)
	}

	var dispatched int
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM platform_events WHERE id = 880041 AND dispatched_at IS NOT NULL`).Scan(&dispatched); err != nil {
		t.Fatalf("read event: %v", err)
	}
	if dispatched != 1 {
		t.Error("an event with no subscribers was left undispatched and would be re-read forever")
	}
}
