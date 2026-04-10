package domain

type TimesheetEntry struct {
	StaffID      int64   `json:"staffId"`
	StaffName    string  `json:"staffName"`
	WeekStart    string  `json:"weekStart"`
	TotalHours   float64 `json:"totalHours"`
	RegularHours float64 `json:"regularHours"`
	OvertimeHours float64 `json:"overtimeHours"`
}
