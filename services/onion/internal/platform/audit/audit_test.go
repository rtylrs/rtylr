package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/testutil"
)

func TestAppend_PersistsTenantActorAndSnapshots(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.WithValue(context.Background(), auth.AuthUserKey, &auth.AuthUser{
		UserID: 7654321,
		IP:     "203.0.113.7",
	})

	const (
		businessID = int64(1001)
		entityID   = int64(7654321)
	)
	result, err := db.ExecContext(ctx, `
		INSERT INTO activity_log (owner_id, business_id, app, entity_type, entity_id, action, description, user_id)
		VALUES (?, ?, 'erp', 'audit_test', ?, 'pending', 'test sentinel', ?)`,
		businessID, businessID, entityID, entityID)
	if err != nil {
		t.Fatalf("insert sentinel: %v", err)
	}
	sentinelID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("sentinel id: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM activity_log WHERE id = ?", sentinelID)
	})

	if err := Append(ctx, db, Event{
		BusinessID:  businessID,
		ActorID:     9999999,
		EntityType:  "product",
		EntityID:    entityID,
		Action:      "updated",
		Description: "product updated",
		OldValues:   map[string]any{"name": "old"},
		NewValues:   map[string]any{"name": "new"},
	}); err != nil {
		t.Fatalf("Append: %v", err)
	}

	var (
		gotBusinessID, gotUserID int64
		gotEntityType, gotAction string
		gotOld, gotNew, gotIP    sql.NullString
	)
	err = db.QueryRowContext(ctx, `
		SELECT business_id, user_id, entity_type, action, CAST(old_values AS CHAR), CAST(new_values AS CHAR), ip_address
		FROM activity_log
		WHERE business_id = ? AND entity_type = 'product' AND entity_id = ? AND action = 'updated'
		ORDER BY id DESC LIMIT 1`, businessID, entityID).Scan(
		&gotBusinessID, &gotUserID, &gotEntityType, &gotAction, &gotOld, &gotNew, &gotIP)
	if err != nil {
		t.Fatalf("read audit event: %v", err)
	}
	if gotBusinessID != businessID || gotUserID != 7654321 || gotEntityType != "product" || gotAction != "updated" {
		t.Fatalf("audit identity = business %d user %d entity %q action %q", gotBusinessID, gotUserID, gotEntityType, gotAction)
	}
	var oldValues, newValues map[string]string
	if err := json.Unmarshal([]byte(gotOld.String), &oldValues); err != nil {
		t.Fatalf("decode old audit snapshot: %v", err)
	}
	if err := json.Unmarshal([]byte(gotNew.String), &newValues); err != nil {
		t.Fatalf("decode new audit snapshot: %v", err)
	}
	if oldValues["name"] != "old" || newValues["name"] != "new" || gotIP.String != "203.0.113.7" {
		t.Fatalf("audit snapshots = old %q new %q ip %q", gotOld.String, gotNew.String, gotIP.String)
	}
}
