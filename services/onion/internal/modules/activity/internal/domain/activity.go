package domain

import (
	"database/sql"
	"time"
)

type ActivityLog struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"ownerId"`
	EntityType  string    `json:"entityType"`
	EntityID    int64     `json:"entityId"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	UserID      int64     `json:"userId"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (a *ActivityLog) FromRow(row *sql.Rows) error {
	var (
		entityID    sql.NullInt64
		description sql.NullString
		userID      sql.NullInt64
	)

	err := row.Scan(
		&a.ID,
		&a.OwnerID,
		&a.EntityType,
		&entityID,
		&a.Action,
		&description,
		&userID,
		&a.CreatedAt,
	)
	if err != nil {
		return err
	}

	if entityID.Valid {
		a.EntityID = entityID.Int64
	}

	if description.Valid {
		a.Description = description.String
	}

	if userID.Valid {
		a.UserID = userID.Int64
	}

	return nil
}
