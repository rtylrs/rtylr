package domain

type DashboardStats struct {
	TotalStaff     int `json:"total_staff"`
	OnShiftToday   int `json:"on_shift_today"`
	PendingTimeOff int `json:"pending_time_off"`
	PayrollDraft   int `json:"payroll_draft"`
}
