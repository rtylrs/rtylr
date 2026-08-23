package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrStaffNotFound = errors.New("staff not found")

type Staff struct {
	ID                    int64     `json:"id"`
	OwnerID               int64     `json:"owner_id"`
	BusinessID            int64     `json:"business_id"`
	UserID                int64     `json:"user_id"`
	DepartmentID          int64     `json:"department_id"`
	ManagerStaffID        int64     `json:"manager_staff_id"`
	DepartmentName        string    `json:"department_name"`
	FirstName             string    `json:"first_name"`
	LastName              string    `json:"last_name"`
	Email                 string    `json:"email"`
	Phone                 string    `json:"phone"`
	Role                  string    `json:"role"`
	EmploymentType        string    `json:"employment_type"`
	Status                string    `json:"status"`
	Salary                float64   `json:"salary"`
	HourlyRate            float64   `json:"hourly_rate"`
	OvertimeRate          float64   `json:"overtime_rate"`
	StartDate             time.Time `json:"start_date"`
	EndDate               time.Time `json:"end_date"`
	Address               string    `json:"address"`
	EmergencyContactName  string    `json:"emergency_contact_name"`
	EmergencyContactPhone string    `json:"emergency_contact_phone"`
	Apps                  string    `json:"apps"`
	Notes                 string    `json:"notes"`
	// PhotoURL and Bio are profile fields, distinct from Notes. Notes is
	// HR-admin commentary about a person; Bio is what that person says about
	// themselves and is shown on their profile.
	PhotoURL  string    `json:"photo_url"`
	Bio       string    `json:"bio"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
	OwnerID               int64   `json:"owner_id"`
	BusinessID            int64   `json:"business_id"`
	UserID                int64   `json:"user_id"`
	DepartmentID          int64   `json:"department_id"`
	ManagerStaffID        int64   `json:"manager_staff_id"`
	FirstName             string  `json:"first_name"`
	LastName              string  `json:"last_name"`
	Email                 string  `json:"email"`
	Phone                 string  `json:"phone"`
	Role                  string  `json:"role"`
	EmploymentType        string  `json:"employment_type"`
	Status                string  `json:"status"`
	Salary                float64 `json:"salary"`
	HourlyRate            float64 `json:"hourly_rate"`
	OvertimeRate          float64 `json:"overtime_rate"`
	StartDate             string  `json:"start_date"`
	EndDate               string  `json:"end_date"`
	Address               string  `json:"address"`
	EmergencyContactName  string  `json:"emergency_contact_name"`
	EmergencyContactPhone string  `json:"emergency_contact_phone"`
	Apps                  string  `json:"apps"`
	Notes                 string  `json:"notes"`
	PhotoURL              string  `json:"photo_url"`
	Bio                   string  `json:"bio"`
}

type StaffUpdateInput struct {
	ID                    int64   `json:"id"`
	DepartmentID          int64   `json:"department_id"`
	ManagerStaffID        int64   `json:"manager_staff_id"`
	FirstName             string  `json:"first_name"`
	LastName              string  `json:"last_name"`
	Email                 string  `json:"email"`
	Phone                 string  `json:"phone"`
	Role                  string  `json:"role"`
	EmploymentType        string  `json:"employment_type"`
	Status                string  `json:"status"`
	Salary                float64 `json:"salary"`
	HourlyRate            float64 `json:"hourly_rate"`
	OvertimeRate          float64 `json:"overtime_rate"`
	EndDate               string  `json:"end_date"`
	Address               string  `json:"address"`
	EmergencyContactName  string  `json:"emergency_contact_name"`
	EmergencyContactPhone string  `json:"emergency_contact_phone"`
	Apps                  string  `json:"apps"`
	Notes                 string  `json:"notes"`
	PhotoURL              string  `json:"photo_url"`
	Bio                   string  `json:"bio"`
}
