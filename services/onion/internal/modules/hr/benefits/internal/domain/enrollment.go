package domain

import (
	"context"
	"errors"
	"time"
)

// EmployeeDirectory resolves which hr_employees row belongs to the calling
// platform user, so this module never has to depend on hr_core directly.
type EmployeeDirectory interface {
	GetEmployeeIDByUserID(ctx context.Context, businessID, userID int64) (int64, error)
}

// ErrAlreadyEnrolled is returned when an employee is already enrolled in the
// requested plan. Enrolling twice is treated as a conflict rather than a
// silent no-op so the caller can tell the two cases apart.
var ErrAlreadyEnrolled = errors.New("employee already enrolled in benefit plan")

// ErrEmployeeDirectoryUnavailable means the module was constructed without an
// EmployeeDirectory, so no caller can be resolved to an employee. Failing
// closed here keeps an unwired module from enrolling anyone.
var ErrEmployeeDirectoryUnavailable = errors.New("employee directory unavailable")

// BenefitEnrollment is one employee's enrollment in one benefit plan.
type BenefitEnrollment struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	EmployeeID  int64
	PlanID      int64
	Status      string
	EnrolledAt  time.Time
	CancelledAt *time.Time
}
