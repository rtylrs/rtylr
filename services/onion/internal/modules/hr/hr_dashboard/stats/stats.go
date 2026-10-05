package stats

import "context"

type DashboardStats struct {
	TotalStaff     int
	OnShiftToday   int
	PendingTimeOff int
	PayrollDraft   int
}

type Loader func(ctx context.Context, businessID int64) (DashboardStats, error)
