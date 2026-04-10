package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrOneOnOneNotFound = errors.New("one-on-one not found")

type OneOnOne struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"ownerId"`
	ManagerID   int64     `json:"managerId"`
	StaffID     int64     `json:"staffId"`
	ScheduledAt time.Time `json:"scheduledAt"`
	Notes       string    `json:"notes"`
	ActionItems string    `json:"actionItems"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (o *OneOnOne) FromRow(row *sql.Rows) error {
	var (
		notes       sql.NullString
		actionItems sql.NullString
		status      sql.NullString
	)

	err := row.Scan(
		&o.ID, &o.OwnerID, &o.ManagerID, &o.StaffID,
		&o.ScheduledAt, &notes, &actionItems, &status, &o.CreatedAt,
	)
	if err != nil {
		return err
	}

	if notes.Valid {
		o.Notes = notes.String
	}
	if actionItems.Valid {
		o.ActionItems = actionItems.String
	}
	if status.Valid {
		o.Status = status.String
	}

	return nil
}

type OneOnOneInput struct {
	ManagerID   int64  `json:"managerId"`
	StaffID     int64  `json:"staffId"`
	ScheduledAt string `json:"scheduledAt"`
	Notes       string `json:"notes"`
	ActionItems string `json:"actionItems"`
	Status      string `json:"status"`
}
