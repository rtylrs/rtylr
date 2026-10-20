package outbox_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/voxire/onion/internal/contracts/audit"
	"github.com/voxire/onion/internal/modules/organization/team/internal/outbox"
	"github.com/voxire/onion/internal/testutil"
)

const auditTenant int64 = 881101

func TestRecordAuditIsRelayedAsHubMutation(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM team_outbox WHERE business_id = ?", auditTenant)
	})

	err := outbox.NewRecorder(db).RecordAudit(ctx, audit.Event{
		BusinessID: auditTenant, ActorID: 12, EntityType: "member", EntityID: 34,
		Action: "member.invited", Description: "Invited Jane as member",
	})
	if err != nil {
		t.Fatalf("RecordAudit: %v", err)
	}

	facts, err := outbox.NewSource(db).Pending(ctx, 500)
	if err != nil {
		t.Fatalf("Pending: %v", err)
	}
	for _, fact := range facts {
		if fact.BusinessID != auditTenant {
			continue
		}
		if fact.Type != audit.MutationRecorded || fact.SubjectType != "member" || fact.SubjectID != 34 {
			t.Fatalf("fact = %+v, want an entity_mutated fact for member 34", fact)
		}
		var payload struct {
			App     string `json:"app"`
			ActorID int64  `json:"actor_id"`
			Action  string `json:"action"`
		}
		// The JSON column scans as json.RawMessage, a named byte type, so a
		// plain []byte assertion would come back empty.
		var raw []byte
		switch p := fact.Payload.(type) {
		case []byte:
			raw = p
		case json.RawMessage:
			raw = p
		}
		if err := json.Unmarshal(raw, &payload); err != nil {
			t.Fatalf("decode payload %T: %v", fact.Payload, err)
		}
		if payload.App != "hub" || payload.ActorID != 12 || payload.Action != "member.invited" {
			t.Fatalf("payload = %+v, want app hub, actor 12, action member.invited", payload)
		}
		return
	}
	t.Fatal("recorded audit fact was not pending for relay")
}
