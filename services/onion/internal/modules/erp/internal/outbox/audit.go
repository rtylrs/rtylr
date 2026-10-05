package outbox

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/voxire/onion/internal/contracts/audit"
	"github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/eventbus"
)

func AppendAuditTx(ctx context.Context, tx *sql.Tx, event audit.Event) error {
	actorID, ip := event.ActorID, ""
	if user, ok := auth.GetAuthUser(ctx); ok && user != nil {
		if user.UserID > 0 {
			actorID = user.UserID
		}
		ip = user.IP
	}
	return Publish(ctx, tx, eventbus.Event{
		BusinessID: event.BusinessID, Type: audit.MutationRecorded,
		SubjectType: event.EntityType, SubjectID: event.EntityID,
		DedupeKey: "mutation:" + uuid.NewString(),
		Payload:   map[string]any{"app": "erp", "actor_id": actorID, "action": event.Action, "description": event.Description, "old_values": event.OldValues, "new_values": event.NewValues, "ip_address": ip},
	})
}
