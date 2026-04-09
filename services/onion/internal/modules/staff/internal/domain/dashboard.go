package domain

type DashboardStats struct {
	TotalStaff     int `json:"totalStaff"`
	OnShiftToday   int `json:"onShiftToday"`
	PendingTimeOff int `json:"pendingTimeOff"`
	PayrollDraft   int `json:"payrollDraft"`
}
