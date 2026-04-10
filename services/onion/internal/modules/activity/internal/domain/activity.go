package domain

import (
	"database/sql"
	"time"
)

type ActivityLog struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"ownerId"`
	App         string    `json:"app"`
	EntityType  string    `json:"entityType"`
	EntityID    int64     `json:"entityId"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	UserID      int64     `json:"userId"`
	UserName    string    `json:"userName"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (a *ActivityLog) FromRow(row *sql.Rows) error {
	var (
		app         sql.NullString
		entityID    sql.NullInt64
		description sql.NullString
		userID      sql.NullInt64
		userName    sql.NullString
	)

	err := row.Scan(
		&a.ID,
		&a.OwnerID,
		&app,
		&a.EntityType,
		&entityID,
		&a.Action,
		&description,
		&userID,
		&userName,
		&a.CreatedAt,
	)
	if err != nil {
		return err
	}

	if app.Valid {
		a.App = app.String
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

	if userName.Valid {
		a.UserName = userName.String
	}

	return nil
}
