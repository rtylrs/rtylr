package domain

import (
	"errors"
	"time"
)

var ErrStaffNotFound = errors.New("staff not found")

type Staff struct {
	ID                    int64
	OwnerID               int64
	BusinessID            int64
	UserID                int64
	DepartmentID          int64
	ManagerStaffID        int64
	DepartmentName        string
	FirstName             string
	LastName              string
	Email                 string
	Phone                 string
	Role                  string
	EmploymentType        string
	Status                string
	Salary                float64
	HourlyRate            float64
	OvertimeRate          float64
	StartDate             time.Time
	EndDate               time.Time
	Address               string
	EmergencyContactName  string
	EmergencyContactPhone string
	Apps                  string
	Notes                 string
	// PhotoURL and Bio are profile fields, distinct from Notes. Notes is
	// HR-admin commentary about a person; Bio is what that person says about
	// themselves and is shown on their profile.
	PhotoURL  string
	Bio       string
	CreatedAt time.Time
	UpdatedAt time.Time
}
type StaffInput struct {
	OwnerID               int64
	BusinessID            int64
	UserID                int64
	DepartmentID          int64
	ManagerStaffID        int64
	FirstName             string
	LastName              string
	Email                 string
	Phone                 string
	Role                  string
	EmploymentType        string
	Status                string
	Salary                float64
	HourlyRate            float64
	OvertimeRate          float64
	StartDate             string
	EndDate               string
	Address               string
	EmergencyContactName  string
	EmergencyContactPhone string
	Apps                  string
	Notes                 string
	PhotoURL              string
	Bio                   string
}

type StaffUpdateInput struct {
	ID                    int64
	DepartmentID          int64
	ManagerStaffID        int64
	FirstName             string
	LastName              string
	Email                 string
	Phone                 string
	Role                  string
	EmploymentType        string
	Status                string
	Salary                float64
	HourlyRate            float64
	OvertimeRate          float64
	EndDate               string
	Address               string
	EmergencyContactName  string
	EmergencyContactPhone string
	Apps                  string
	Notes                 string
	PhotoURL              string
	Bio                   string
}
