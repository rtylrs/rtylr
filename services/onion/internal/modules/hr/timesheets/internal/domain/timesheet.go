package domain

type TimesheetEntry struct {
	StaffID       int64
	StaffName     string
	WeekStart     string
	TotalHours    float64
	RegularHours  float64
	OvertimeHours float64
}
