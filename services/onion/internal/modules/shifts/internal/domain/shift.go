package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrShiftNotFound = errors.New("shift not found")

type Shift struct {
	ID                   int64     `json:"id"`
	OwnerID              int64     `json:"owner_id"`
	BusinessID           int64     `json:"business_id"`
	StaffID              int64     `json:"staff_id"`
	DepartmentID         int64     `json:"department_id"`
	Date                 time.Time `json:"date"`
	StartTime            string    `json:"start_time"`
	EndTime              string    `json:"end_time"`
	BreakDurationMinutes int       `json:"break_duration_minutes"`
	Status               string    `json:"status"`
	ClockInAt            time.Time `json:"clock_in_at"`
	ClockOutAt           time.Time `json:"clock_out_at"`
	HoursWorked          float64   `json:"hours_worked"`
	OvertimeHours        float64   `json:"overtime_hours"`
	Notes                string    `json:"notes"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
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
		&s.BusinessID,
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
	OwnerID              int64  `json:"owner_id"`
	BusinessID           int64  `json:"business_id"`
	StaffID              int64  `json:"staff_id"`
	DepartmentID         int64  `json:"department_id"`
	Date                 string `json:"date"`
	StartTime            string `json:"start_time"`
	EndTime              string `json:"end_time"`
	BreakDurationMinutes int    `json:"break_duration_minutes"`
	Notes                string `json:"notes"`
}

type ShiftUpdateInput struct {
	ID                   int64  `json:"id"`
	DepartmentID         int64  `json:"department_id"`
	Date                 string `json:"date"`
	StartTime            string `json:"start_time"`
	EndTime              string `json:"end_time"`
	BreakDurationMinutes int    `json:"break_duration_minutes"`
	Status               string `json:"status"`
	Notes                string `json:"notes"`
}

type ClockInInput struct {
	ShiftID int64 `json:"shift_id"`
}

type ClockOutInput struct {
	ShiftID int64 `json:"shift_id"`
}
