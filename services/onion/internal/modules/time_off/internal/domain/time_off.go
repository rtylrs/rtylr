package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrTimeOffNotFound = errors.New("time off request not found")

type TimeOffRequest struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"owner_id"`
	BusinessID int64     `json:"business_id"`
	StaffID    int64     `json:"staff_id"`
	Type       string    `json:"type"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	DaysCount  float64   `json:"days_count"`
	Status     string    `json:"status"`
	Reason     string    `json:"reason"`
	ApprovedBy int64     `json:"approved_by"`
	ApprovedAt time.Time `json:"approved_at"`
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (t *TimeOffRequest) FromRow(row *sql.Rows) error {
	var (
		status     sql.NullString
		reason     sql.NullString
		approvedBy sql.NullInt64
		approvedAt sql.NullTime
		notes      sql.NullString
	)

	err := row.Scan(
		&t.ID,
		&t.OwnerID,
		&t.BusinessID,
		&t.StaffID,
		&t.Type,
		&t.StartDate,
		&t.EndDate,
		&t.DaysCount,
		&status,
		&reason,
		&approvedBy,
		&approvedAt,
		&notes,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if status.Valid {
		t.Status = status.String
	}

	if reason.Valid {
		t.Reason = reason.String
	}

	if approvedBy.Valid {
		t.ApprovedBy = approvedBy.Int64
	}

	if approvedAt.Valid {
		t.ApprovedAt = approvedAt.Time
	}

	if notes.Valid {
		t.Notes = notes.String
	}

	return nil
}

type TimeOffInput struct {
	OwnerID    int64   `json:"owner_id"`
	BusinessID int64   `json:"business_id"`
	StaffID    int64   `json:"staff_id"`
	Type      string  `json:"type"`
	StartDate string  `json:"start_date"`
	EndDate   string  `json:"end_date"`
	DaysCount float64 `json:"days_count"`
	Reason    string  `json:"reason"`
	Notes     string  `json:"notes"`
}

type TimeOffApprovalInput struct {
	ID         int64  `json:"id"`
	Status     string `json:"status"`
	ApprovedBy int64  `json:"approved_by"`
	Notes      string `json:"notes"`
}
