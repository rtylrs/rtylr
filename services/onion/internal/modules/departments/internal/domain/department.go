package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrDepartmentNotFound = errors.New("department not found")

type Department struct {
	ID             int64     `json:"id"`
	OwnerID        int64     `json:"ownerId"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	ManagerStaffID int64     `json:"managerStaffId"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func (d *Department) FromRow(row *sql.Rows) error {
	var (
		description    sql.NullString
		managerStaffID sql.NullInt64
	)

	err := row.Scan(
		&d.ID,
		&d.OwnerID,
		&d.Name,
		&description,
		&managerStaffID,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if description.Valid {
		d.Description = description.String
	}

	if managerStaffID.Valid {
		d.ManagerStaffID = managerStaffID.Int64
	}

	return nil
}

type DepartmentInput struct {
	ID             int64  `json:"id"`
	OwnerID        int64  `json:"ownerId"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	ManagerStaffID int64  `json:"managerStaffId"`
}
