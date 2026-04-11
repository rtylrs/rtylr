package domain

type DashboardData struct {
	TodayCount       int     `json:"todayCount"`
	UpcomingCount    int     `json:"upcomingCount"`
	CancellationRate float64 `json:"cancellationRate"`
	Revenue          float64 `json:"revenue"`
}
