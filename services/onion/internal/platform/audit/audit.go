// Package audit writes tenant-scoped, append-only mutation events.
package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/voxire/onion/internal/platform/auth"
)

type Event struct {
	BusinessID  int64
	ActorID     int64
	EntityType  string
	EntityID    int64
	Action      string
	Description string
	OldValues   any
	NewValues   any
}

func Append(ctx context.Context, db *sql.DB, event Event) error {
	if db == nil {
		return nil
	}

	actorID, ipAddress := actorDetails(ctx, event.ActorID)
	oldValues, err := marshalNullable(event.OldValues)
	if err != nil {
		return fmt.Errorf("marshal audit old values: %w", err)
	}
	newValues, err := marshalNullable(event.NewValues)
	if err != nil {
		return fmt.Errorf("marshal audit new values: %w", err)
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO activity_log
		(owner_id, business_id, app, entity_type, entity_id, action, description, user_id, old_values, new_values, ip_address)
		VALUES (?, ?, 'erp', ?, ?, ?, ?, ?, ?, ?, ?)`,
		event.BusinessID, event.BusinessID, event.EntityType, event.EntityID,
		event.Action, event.Description, nullableInt64(actorID), oldValues, newValues, nullableString(ipAddress))
	if err != nil {
		return fmt.Errorf("append audit event: %w", err)
	}
	return nil
}

func actorDetails(ctx context.Context, fallbackID int64) (int64, string) {
	actorID := fallbackID
	ipAddress := ""
	if user, ok := auth.GetAuthUser(ctx); ok && user != nil {
		if user.UserID > 0 {
			actorID = user.UserID
		}
		ipAddress = user.IP
	}
	return actorID, ipAddress
}

func marshalNullable(value any) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	return json.Marshal(value)
}

func nullableInt64(value int64) any {
	if value == 0 {
		return nil
	}
	return value
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
