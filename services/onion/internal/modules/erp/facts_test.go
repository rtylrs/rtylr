package erp_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/contracts/audit"
	"github.com/voxire/onion/internal/modules/activity"
	"github.com/voxire/onion/internal/modules/erp"
	"github.com/voxire/onion/internal/modules/erp/internal/outbox"
	"github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/eventbus"
	"github.com/voxire/onion/internal/testutil"
)

func TestMutationFactSurvivesRelayRetryWithoutRepeatingActivity(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := auth.WithAuthUser(context.Background(), &auth.AuthUser{UserID: 7654321, IP: "203.0.113.7"})
	const businessID = int64(881009)
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM activity_log WHERE business_id = ?", businessID)
		_, _ = db.Exec("DELETE FROM platform_events WHERE business_id = ?", businessID)
		_, _ = db.Exec("DELETE FROM erp_outbox WHERE business_id = ?", businessID)
	})
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = outbox.AppendAuditTx(ctx, tx, audit.Event{BusinessID: businessID, ActorID: 99999, EntityType: "product", EntityID: 7654321, Action: "updated", OldValues: map[string]any{"name": "old"}, NewValues: map[string]any{"name": "new"}})
	if err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	var before int
	if err := db.QueryRow("SELECT COUNT(*) FROM activity_log WHERE business_id = ?", businessID).Scan(&before); err != nil {
		t.Fatal(err)
	}
	if before != 0 {
		t.Fatal("producer wrote the Activity projection before relay")
	}
	source := &failingAcknowledgement{FactSource: erp.Facts(db), fail: true}
	dispatcher := eventbus.NewDispatcher(db, activity.NewModule(ctx, db)).WithSources(source)
	if _, err := dispatcher.DispatchBatch(ctx); err == nil {
		t.Fatal("expected failed acknowledgement")
	}
	if _, err := dispatcher.DispatchBatch(ctx); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM activity_log WHERE business_id = ?", businessID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("Activity rows=%d, want one after retry", count)
	}
	var actorID int64
	var oldValues, newValues, ip sql.NullString
	if err := db.QueryRow("SELECT user_id, CAST(old_values AS CHAR), CAST(new_values AS CHAR), ip_address FROM activity_log WHERE business_id = ?", businessID).Scan(&actorID, &oldValues, &newValues, &ip); err != nil {
		t.Fatal(err)
	}
	var oldSnapshot, newSnapshot map[string]string
	if err := json.Unmarshal([]byte(oldValues.String), &oldSnapshot); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(newValues.String), &newSnapshot); err != nil {
		t.Fatal(err)
	}
	if actorID != 7654321 || ip.String != "203.0.113.7" || oldSnapshot["name"] != "old" || newSnapshot["name"] != "new" {
		t.Fatalf("incorrect actor or snapshots: actor=%d ip=%q old=%v new=%v", actorID, ip.String, oldSnapshot, newSnapshot)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM platform_events WHERE business_id = ?", businessID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("archive rows=%d, want one", count)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM erp_outbox WHERE business_id = ? AND relayed_at IS NULL", businessID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("pending producer facts=%d after retry", count)
	}
}

type failingAcknowledgement struct {
	eventbus.FactSource
	fail bool
}

func (s *failingAcknowledgement) MarkRelayed(ctx context.Context, id uint64) error {
	if s.fail {
		s.fail = false
		return errors.New("synthetic acknowledgement failure")
	}
	return s.FactSource.MarkRelayed(ctx, id)
}
