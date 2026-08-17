package domain

import (
	"errors"
	"strings"
)

var (
	ErrEmptyEmployeeName = errors.New("employee first name and last name are required")
	ErrEmptyEmail        = errors.New("employee email is required")
	ErrInvalidSalary     = errors.New("salary must be non-negative")
	ErrEmptyDeptName     = errors.New("department name is required")
	ErrEmptyLocationName = errors.New("location name is required")
	ErrEmptyRoleName     = errors.New("role name is required")
	ErrEmptyScheduleName = errors.New("work schedule name is required")
)

// ValidateEmployee checks that required employee fields are present and valid.
func ValidateEmployee(emp EmployeeRecord) error {
	if strings.TrimSpace(emp.FirstName) == "" || strings.TrimSpace(emp.LastName) == "" {
		return ErrEmptyEmployeeName
	}
	if strings.TrimSpace(emp.Email) == "" {
		return ErrEmptyEmail
	}
	if emp.SalaryCents < 0 {
		return ErrInvalidSalary
	}
	if emp.HourlyRateCents < 0 {
		return ErrInvalidSalary
	}
	return nil
}

// ValidateDepartment checks that required department fields are present.
func ValidateDepartment(dept DepartmentRecord) error {
	if strings.TrimSpace(dept.Name) == "" {
		return ErrEmptyDeptName
	}
	return nil
}

// ValidateLocation checks that required location fields are present.
func ValidateLocation(loc LocationRecord) error {
	if strings.TrimSpace(loc.Name) == "" {
		return ErrEmptyLocationName
	}
	return nil
}

// ValidateRole checks that required role fields are present.
func ValidateRole(role RoleRecord) error {
	if strings.TrimSpace(role.Name) == "" {
		return ErrEmptyRoleName
	}
	return nil
}

// ValidateWorkSchedule checks that required work schedule fields are present.
func ValidateWorkSchedule(ws WorkScheduleRecord) error {
	if strings.TrimSpace(ws.Name) == "" {
		return ErrEmptyScheduleName
	}
	return nil
}

// IsValidEmploymentStatus returns true when the status is a recognised value.
func IsValidEmploymentStatus(status string) bool {
	switch status {
	case EmploymentStatusActive, EmploymentStatusInactive, EmploymentStatusInvited:
		return true
	}
	return false
}

// IsValidEmploymentType returns true when the type is a recognised value.
func IsValidEmploymentType(empType string) bool {
	switch empType {
	case EmploymentTypeFullTime, EmploymentTypePartTime, EmploymentTypeContract, EmploymentTypeFreelance:
		return true
	}
	return false
}
