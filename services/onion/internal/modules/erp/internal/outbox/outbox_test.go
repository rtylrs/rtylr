package outbox_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/erp/internal/outbox"
	"github.com/voxire/onion/internal/platform/eventbus"
	"github.com/voxire/onion/internal/testutil"
)

const publishTenant int64 = 881001

// The outbox exists to make one promise: an event is recorded if and only if
// the change it describes is. Publish takes its owning transaction so
// that promise is structural, and this is the test that it holds.
func TestPublishJoinsTheCallersTransaction(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM erp_outbox WHERE business_id = ?", publishTenant)
	})

	count := func() int {
		t.Helper()
		var n int
		if err := db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM erp_outbox WHERE business_id = ?", publishTenant).Scan(&n); err != nil {
			t.Fatalf("count events: %v", err)
		}
		return n
	}

	event := eventbus.Event{
		BusinessID:  publishTenant,
		Type:        eventbus.TypeOrderCreated,
		SubjectType: eventbus.SubjectOrder,
		SubjectID:   4001,
		Payload:     map[string]any{"order_id": 4001},
	}

	// A rolled-back change takes its event with it.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if err := outbox.Publish(ctx, tx, event); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if got := count(); got != 0 {
		t.Fatalf("%d events survived a rolled-back transaction, want 0: subscribers would be told about a change that never happened", got)
	}

	// A committed change keeps it.
	tx, err = db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if err := outbox.Publish(ctx, tx, event); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	if got := count(); got != 1 {
		t.Fatalf("%d events after commit, want 1", got)
	}

	// The dedupe key defaults to type:subject_type:subject_id, so the same
	// producer running twice records once.
	var dedupeKey string
	if err := db.QueryRowContext(ctx,
		"SELECT dedupe_key FROM erp_outbox WHERE business_id = ?", publishTenant).Scan(&dedupeKey); err != nil {
		t.Fatalf("read dedupe key: %v", err)
	}
	if want := "order_created:order:4001"; dedupeKey != want {
		t.Errorf("dedupe_key = %q, want %q", dedupeKey, want)
	}

	tx, err = db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if err := outbox.Publish(ctx, tx, event); err != nil {
		t.Fatalf("Publish on retry: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit retry: %v", err)
	}
	if got := count(); got != 1 {
		t.Errorf("%d events after a retried producer, want 1", got)
	}
}

// An event with no tenant cannot be routed to anyone safely, and this is a
// fan-out path, so it is rejected before it reaches the database rather than
// relying on the NOT NULL to catch it.
func TestPublishRejectsAnEventWithNoTenant(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	err = outbox.Publish(ctx, tx, eventbus.Event{
		Type:        eventbus.TypeOrderCreated,
		SubjectType: eventbus.SubjectOrder,
		SubjectID:   4002,
	})
	if err == nil {
		t.Fatal("published an event with no business id")
	}
}

// Everything else a malformed event can be missing.
func TestPublishRejectsAnIncompleteEvent(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	cases := map[string]eventbus.Event{
		"no type":         {BusinessID: publishTenant, SubjectType: eventbus.SubjectOrder, SubjectID: 1},
		"no subject type": {BusinessID: publishTenant, Type: eventbus.TypeOrderCreated, SubjectID: 1},
		"no subject id":   {BusinessID: publishTenant, Type: eventbus.TypeOrderCreated, SubjectType: eventbus.SubjectOrder},
	}
	for name, event := range cases {
		t.Run(name, func(t *testing.T) {
			if err := outbox.Publish(ctx, tx, event); err == nil {
				t.Errorf("published an event with %s", name)
			}
		})
	}
}

// A producer outside a transaction has no outbox guarantee at all, so this is
// refused rather than quietly accepted against the pool.
func TestPublishRequiresATransaction(t *testing.T) {
	if err := outbox.Publish(context.Background(), nil, eventbus.Event{
		BusinessID:  publishTenant,
		Type:        eventbus.TypeOrderCreated,
		SubjectType: eventbus.SubjectOrder,
		SubjectID:   1,
	}); err == nil {
		t.Fatal("published without a transaction")
	}
}
