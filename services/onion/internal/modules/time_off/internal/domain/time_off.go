package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrTimeOffNotFound = errors.New("time off request not found")

type TimeOffRequest struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"ownerId"`
	StaffID    int64     `json:"staffId"`
	Type       string    `json:"type"`
	StartDate  time.Time `json:"startDate"`
	EndDate    time.Time `json:"endDate"`
	DaysCount  float64   `json:"daysCount"`
	Status     string    `json:"status"`
	Reason     string    `json:"reason"`
	ApprovedBy int64     `json:"approvedBy"`
	ApprovedAt time.Time `json:"approvedAt"`
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
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
	OwnerID   int64   `json:"ownerId"`
	StaffID   int64   `json:"staffId"`
	Type      string  `json:"type"`
	StartDate string  `json:"startDate"`
	EndDate   string  `json:"endDate"`
	DaysCount float64 `json:"daysCount"`
	Reason    string  `json:"reason"`
	Notes     string  `json:"notes"`
}

type TimeOffApprovalInput struct {
	ID         int64  `json:"id"`
	Status     string `json:"status"`
	ApprovedBy int64  `json:"approvedBy"`
	Notes      string `json:"notes"`
}
