package domain

type TimesheetEntry struct {
	StaffID       int64   `json:"staff_id"`
	StaffName     string  `json:"staff_name"`
	WeekStart     string  `json:"week_start"`
	TotalHours    float64 `json:"total_hours"`
	RegularHours  float64 `json:"regular_hours"`
	OvertimeHours float64 `json:"overtime_hours"`
}
