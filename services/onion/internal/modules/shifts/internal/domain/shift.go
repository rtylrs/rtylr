package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrShiftNotFound = errors.New("shift not found")

type Shift struct {
	ID                   int64     `json:"id"`
	OwnerID              int64     `json:"ownerId"`
	StaffID              int64     `json:"staffId"`
	DepartmentID         int64     `json:"departmentId"`
	Date                 time.Time `json:"date"`
	StartTime            string    `json:"startTime"`
	EndTime              string    `json:"endTime"`
	BreakDurationMinutes int       `json:"breakDurationMinutes"`
	Status               string    `json:"status"`
	ClockInAt            time.Time `json:"clockInAt"`
	ClockOutAt           time.Time `json:"clockOutAt"`
	HoursWorked          float64   `json:"hoursWorked"`
	OvertimeHours        float64   `json:"overtimeHours"`
	Notes                string    `json:"notes"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

func (s *Shift) FromRow(row *sql.Rows) error {
	var (
		departmentID         sql.NullInt64
		breakDurationMinutes sql.NullInt64
		status               sql.NullString
		clockInAt            sql.NullTime
		clockOutAt           sql.NullTime
		hoursWorked          sql.NullFloat64
		overtimeHours        sql.NullFloat64
		notes                sql.NullString
	)

	err := row.Scan(
		&s.ID,
		&s.OwnerID,
		&s.StaffID,
		&departmentID,
		&s.Date,
		&s.StartTime,
		&s.EndTime,
		&breakDurationMinutes,
		&status,
		&clockInAt,
		&clockOutAt,
		&hoursWorked,
		&overtimeHours,
		&notes,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if departmentID.Valid {
		s.DepartmentID = departmentID.Int64
	}

	if breakDurationMinutes.Valid {
		s.BreakDurationMinutes = int(breakDurationMinutes.Int64)
	}

	if status.Valid {
		s.Status = status.String
	}

	if clockInAt.Valid {
		s.ClockInAt = clockInAt.Time
	}

	if clockOutAt.Valid {
		s.ClockOutAt = clockOutAt.Time
	}

	if hoursWorked.Valid {
		s.HoursWorked = hoursWorked.Float64
	}

	if overtimeHours.Valid {
		s.OvertimeHours = overtimeHours.Float64
	}

	if notes.Valid {
		s.Notes = notes.String
	}

	return nil
}

type ShiftInput struct {
	OwnerID              int64  `json:"ownerId"`
	StaffID              int64  `json:"staffId"`
	DepartmentID         int64  `json:"departmentId"`
	Date                 string `json:"date"`
	StartTime            string `json:"startTime"`
	EndTime              string `json:"endTime"`
	BreakDurationMinutes int    `json:"breakDurationMinutes"`
	Notes                string `json:"notes"`
}

type ShiftUpdateInput struct {
	ID                   int64  `json:"id"`
	DepartmentID         int64  `json:"departmentId"`
	Date                 string `json:"date"`
	StartTime            string `json:"startTime"`
	EndTime              string `json:"endTime"`
	BreakDurationMinutes int    `json:"breakDurationMinutes"`
	Status               string `json:"status"`
	Notes                string `json:"notes"`
}

type ClockInInput struct {
	ShiftID int64 `json:"shiftId"`
}

type ClockOutInput struct {
	ShiftID int64 `json:"shiftId"`
}
