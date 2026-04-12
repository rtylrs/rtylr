package domain

type DashboardData struct {
	TodayCount       int     `json:"today_count"`
	UpcomingCount    int     `json:"upcoming_count"`
	CancellationRate float64 `json:"cancellation_rate"`
	Revenue          float64 `json:"revenue"`
}
