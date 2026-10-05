package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/webhooks/internal/application"
	"github.com/voxire/onion/internal/modules/webhooks/internal/domain"
	"github.com/voxire/onion/internal/modules/webhooks/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/eventbus"
	"github.com/voxire/onion/internal/testutil"
)

const (
	listTenant  int64 = 884001
	listOther   int64 = 884002
	listOwner   int64 = 884000
	listHookA   int64 = 884011
	listHookB   int64 = 884012
	listHookFor int64 = 884013
)

func TestListDeliveriesForBusinessIsScopedAndFilterable(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	repo := infrastructure.NewRepository(db)
	service := application.NewService(repo, eventbus.NewSQLReader(db))

	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.ExecContext(ctx, query, args...); err != nil {
			t.Fatalf("seed (%s): %v", query, err)
		}
	}
	clearSeed := func() {
		for _, stmt := range []string{
			"DELETE FROM webhook_deliveries WHERE business_id IN (?, ?)",
			"DELETE FROM platform_events WHERE business_id IN (?, ?)",
			"DELETE FROM hub_webhooks WHERE business_id IN (?, ?)",
		} {
			_, _ = db.ExecContext(ctx, stmt, listTenant, listOther)
		}
	}
	clearSeed()
	t.Cleanup(clearSeed)

	for _, spec := range []struct {
		id       int64
		business int64
	}{{listHookA, listTenant}, {listHookB, listTenant}, {listHookFor, listOther}} {
		exec(`INSERT INTO hub_webhooks (id, owner_id, business_id, url, events, secret, is_active)
		      VALUES (?, ?, ?, 'https://example.test/hook', JSON_ARRAY('order_created'), 's', TRUE)`,
			spec.id, listOwner, spec.business)
	}
	exec(`INSERT INTO platform_events (id, business_id, event_type, subject_type, subject_id, payload, dedupe_key)
	      VALUES (884021, ?, 'order_created', 'order', 501, JSON_OBJECT('order_id',501), 'k1')`, listTenant)
	exec(`INSERT INTO platform_events (id, business_id, event_type, subject_type, subject_id, payload, dedupe_key)
	      VALUES (884022, ?, 'order_created', 'order', 502, JSON_OBJECT('order_id',502), 'k2')`, listOther)

	// Two for our tenant, one delivered and one dead, plus one belonging to
	// somebody else.
	exec(`INSERT INTO webhook_deliveries (id, business_id, webhook_id, event_id, status, attempts, response_status)
	      VALUES (884031, ?, ?, 884021, 'delivered', 1, 200)`, listTenant, listHookA)
	exec(`INSERT INTO webhook_deliveries (id, business_id, webhook_id, event_id, status, attempts, response_status, last_error)
	      VALUES (884032, ?, ?, 884021, 'dead', 5, 500, 'endpoint answered 500')`, listTenant, listHookB)
	exec(`INSERT INTO webhook_deliveries (id, business_id, webhook_id, event_id, status, attempts)
	      VALUES (884033, ?, ?, 884022, 'dead', 5)`, listOther, listHookFor)

	// Unfiltered: our two, never the other tenant's.
	got, total, err := service.ListDeliveries(ctx, listTenant, domain.DeliveryFilter{Limit: 50})
	if err != nil {
		t.Fatalf("ListDeliveriesForBusiness: %v", err)
	}
	if total != 2 || len(got) != 2 {
		t.Fatalf("got %d rows (total %d), want 2: another tenant's deliveries are visible", len(got), total)
	}
	// Most recent first.
	if got[0].ID != 884032 {
		t.Errorf("first row id = %d, want 884032 (newest first)", got[0].ID)
	}
	// The event type is joined in, which is what makes a row readable alone.
	if got[0].EventType != "order_created" {
		t.Errorf("event_type = %q, want order_created", got[0].EventType)
	}
	if got[0].Status != "dead" || got[0].ResponseStatus != 500 || got[0].LastError != "endpoint answered 500" {
		t.Errorf("dead row = %+v, want the status, response and error preserved", got[0])
	}

	// Status filter, which is the point of the surface.
	dead, deadTotal, err := service.ListDeliveries(ctx, listTenant, domain.DeliveryFilter{Status: "dead", Limit: 50})
	if err != nil {
		t.Fatalf("filter by status: %v", err)
	}
	if deadTotal != 1 || len(dead) != 1 || dead[0].ID != 884032 {
		t.Errorf("dead filter returned %d rows (total %d), want just 884032", len(dead), deadTotal)
	}

	// Registration filter.
	byHook, hookTotal, err := service.ListDeliveries(ctx, listTenant, domain.DeliveryFilter{WebhookID: listHookA, Limit: 50})
	if err != nil {
		t.Fatalf("filter by webhook: %v", err)
	}
	if hookTotal != 1 || len(byHook) != 1 || byHook[0].WebhookID != listHookA {
		t.Errorf("webhook filter returned %d rows (total %d), want just hook A", len(byHook), hookTotal)
	}

	// The filter is caller-supplied. Naming another tenant's registration must
	// return nothing rather than reaching it, because the business scope sits
	// underneath the filter rather than beside it.
	foreign, foreignTotal, err := service.ListDeliveries(ctx, listTenant, domain.DeliveryFilter{WebhookID: listHookFor, Limit: 50})
	if err != nil {
		t.Fatalf("filter by a foreign webhook: %v", err)
	}
	if foreignTotal != 0 || len(foreign) != 0 {
		t.Errorf("naming another tenant's registration returned %d rows (total %d), want 0", len(foreign), foreignTotal)
	}

	// An unrecognised status matches nothing rather than erroring.
	unknown, unknownTotal, err := service.ListDeliveries(ctx, listTenant, domain.DeliveryFilter{Status: "nonsense", Limit: 50})
	if err != nil {
		t.Fatalf("filter by an unknown status: %v", err)
	}
	if unknownTotal != 0 || len(unknown) != 0 {
		t.Errorf("unknown status returned %d rows (total %d), want 0", len(unknown), unknownTotal)
	}

	// Pagination reports the full match count, not the page size.
	firstPage, pagedTotal, err := service.ListDeliveries(ctx, listTenant, domain.DeliveryFilter{Limit: 1})
	if err != nil {
		t.Fatalf("paged listing: %v", err)
	}
	if len(firstPage) != 1 || pagedTotal != 2 {
		t.Errorf("page of 1 returned %d rows with total %d, want 1 and 2", len(firstPage), pagedTotal)
	}
}
