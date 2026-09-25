package domain

// BillableTimeEntry is approved billable HR time that has not been linked to
// a finance invoice yet.
type BillableTimeEntry struct {
	ID              int64
	EmployeeID      int64
	ProjectID       *int64
	TaskID          *int64
	Date            string
	DurationMinutes int64
	Description     string
	IsBillable      bool
	Status          string
	Source          string
}
