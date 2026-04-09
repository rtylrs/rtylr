package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrStaffNotFound = errors.New("staff not found")

type Staff struct {
	ID                    int64     `json:"id"`
	OwnerID               int64     `json:"ownerId"`
	UserID                int64     `json:"userId"`
	DepartmentID          int64     `json:"departmentId"`
	FirstName             string    `json:"firstName"`
	LastName              string    `json:"lastName"`
	Email                 string    `json:"email"`
	Phone                 string    `json:"phone"`
	Role                  string    `json:"role"`
	EmploymentType        string    `json:"employmentType"`
	Status                string    `json:"status"`
	Salary                float64   `json:"salary"`
	HourlyRate            float64   `json:"hourlyRate"`
	OvertimeRate          float64   `json:"overtimeRate"`
	StartDate             time.Time `json:"startDate"`
	EndDate               time.Time `json:"endDate"`
	Address               string    `json:"address"`
	EmergencyContactName  string    `json:"emergencyContactName"`
	EmergencyContactPhone string    `json:"emergencyContactPhone"`
	Apps                  string    `json:"apps"`
	Notes                 string    `json:"notes"`
	CreatedAt             time.Time `json:"createdAt"`
	UpdatedAt             time.Time `json:"updatedAt"`
}

func (s *Staff) FromRow(row *sql.Rows) error {
	var (
		userID                sql.NullInt64
		departmentID          sql.NullInt64
		phone                 sql.NullString
		employmentType        sql.NullString
		status                sql.NullString
		salary                sql.NullFloat64
		hourlyRate            sql.NullFloat64
		overtimeRate          sql.NullFloat64
		endDate               sql.NullTime
		address               sql.NullString
		emergencyContactName  sql.NullString
		emergencyContactPhone sql.NullString
		apps                  sql.NullString
		notes                 sql.NullString
	)

	err := row.Scan(
		&s.ID,
		&s.OwnerID,
		&userID,
		&departmentID,
		&s.FirstName,
		&s.LastName,
		&s.Email,
		&phone,
		&s.Role,
		&employmentType,
		&status,
		&salary,
		&hourlyRate,
		&overtimeRate,
		&s.StartDate,
		&endDate,
		&address,
		&emergencyContactName,
		&emergencyContactPhone,
		&apps,
		&notes,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if userID.Valid {
		s.UserID = userID.Int64
	}

	if departmentID.Valid {
		s.DepartmentID = departmentID.Int64
	}

	if phone.Valid {
		s.Phone = phone.String
	}

	if employmentType.Valid {
		s.EmploymentType = employmentType.String
	}

	if status.Valid {
		s.Status = status.String
	}

	if salary.Valid {
		s.Salary = salary.Float64
	}

	if hourlyRate.Valid {
		s.HourlyRate = hourlyRate.Float64
	}

	if overtimeRate.Valid {
		s.OvertimeRate = overtimeRate.Float64
	}

	if endDate.Valid {
		s.EndDate = endDate.Time
	}

	if address.Valid {
		s.Address = address.String
	}

	if emergencyContactName.Valid {
		s.EmergencyContactName = emergencyContactName.String
	}

	if emergencyContactPhone.Valid {
		s.EmergencyContactPhone = emergencyContactPhone.String
	}

	if apps.Valid {
		s.Apps = apps.String
	}

	if notes.Valid {
		s.Notes = notes.String
	}

	return nil
}

type StaffInput struct {
	OwnerID               int64   `json:"ownerId"`
	UserID                int64   `json:"userId"`
	DepartmentID          int64   `json:"departmentId"`
	FirstName             string  `json:"firstName"`
	LastName              string  `json:"lastName"`
	Email                 string  `json:"email"`
	Phone                 string  `json:"phone"`
	Role                  string  `json:"role"`
	EmploymentType        string  `json:"employmentType"`
	Status                string  `json:"status"`
	Salary                float64 `json:"salary"`
	HourlyRate            float64 `json:"hourlyRate"`
	OvertimeRate          float64 `json:"overtimeRate"`
	StartDate             string  `json:"startDate"`
	EndDate               string  `json:"endDate"`
	Address               string  `json:"address"`
	EmergencyContactName  string  `json:"emergencyContactName"`
	EmergencyContactPhone string  `json:"emergencyContactPhone"`
	Apps                  string  `json:"apps"`
	Notes                 string  `json:"notes"`
}

type StaffUpdateInput struct {
	ID                    int64   `json:"id"`
	DepartmentID          int64   `json:"departmentId"`
	FirstName             string  `json:"firstName"`
	LastName              string  `json:"lastName"`
	Email                 string  `json:"email"`
	Phone                 string  `json:"phone"`
	Role                  string  `json:"role"`
	EmploymentType        string  `json:"employmentType"`
	Status                string  `json:"status"`
	Salary                float64 `json:"salary"`
	HourlyRate            float64 `json:"hourlyRate"`
	OvertimeRate          float64 `json:"overtimeRate"`
	EndDate               string  `json:"endDate"`
	Address               string  `json:"address"`
	EmergencyContactName  string  `json:"emergencyContactName"`
	EmergencyContactPhone string  `json:"emergencyContactPhone"`
	Apps                  string  `json:"apps"`
	Notes                 string  `json:"notes"`
}
