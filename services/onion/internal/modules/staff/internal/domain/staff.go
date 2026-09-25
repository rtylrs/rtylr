package domain

import (
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
