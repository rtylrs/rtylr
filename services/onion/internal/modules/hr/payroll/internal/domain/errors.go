package domain

import "errors"

var (
	// ErrPayrollPeriodAlreadyPaid is returned when a period already holds a
	// paid record. Recomputing it would restate figures for money that has
	// already left the business, so the whole run is refused rather than
	// merged with the rows that were already settled.
	ErrPayrollPeriodAlreadyPaid = errors.New("payroll for this period has already been paid and cannot be recalculated")

	// ErrPayrollPeriodConflict is returned when a concurrent run or payment
	// changed the period while this run was writing it. The write is rolled
	// back whole, so retrying is safe.
	ErrPayrollPeriodConflict = errors.New("payroll for this period changed while it was being written; retry the run")

	// ErrPayrollRecordNotFound is returned for a payroll id that does not
	// exist in the caller's business.
	ErrPayrollRecordNotFound = errors.New("payroll record not found")

	// ErrPayrollRecordAlreadyPaid is returned when the record itself has
	// already been settled.
	ErrPayrollRecordAlreadyPaid = errors.New("payroll record is already paid")

	// ErrPayrollRecordHasPaidSibling is returned when the same staff member
	// already has a paid record for the same period. That is the state a
	// duplicated payroll run leaves behind, and settling the second row would
	// pay the employee twice for one period.
	ErrPayrollRecordHasPaidSibling = errors.New("another payroll record for this staff member and period is already paid")
)
