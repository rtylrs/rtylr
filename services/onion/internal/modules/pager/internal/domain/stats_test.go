package domain

import (
	"testing"
	"time"
)

func TestBuildStatsCountsInTheVenueTimeZone(t *testing.T) {
	beirut, err := time.LoadLocation("Asia/Beirut")
	if err != nil {
		t.Skip("no tzdata:", err)
	}
	// 2026-11-22 10:00 in Beirut (UTC+2).
	now := time.Date(2026, 11, 22, 8, 0, 0, 0, time.UTC)
	at := func(day, hour, minute int) time.Time {
		return time.Date(2026, 11, day, hour, minute, 0, 0, beirut)
	}
	tickets := []TicketTimes{
		// 23:30 UTC on the 20th is 01:30 on the 21st in Beirut.
		{CreatedAt: time.Date(2026, 11, 20, 23, 30, 0, 0, time.UTC), ReadyAt: time.Date(2026, 11, 20, 23, 34, 0, 0, time.UTC)},
		{CreatedAt: at(22, 9, 0), ReadyAt: at(22, 9, 2)},
		{CreatedAt: at(22, 9, 10), ReadyAt: at(22, 9, 16)},
		{CreatedAt: at(22, 9, 20), ReadyAt: at(22, 9, 30)},
		// Never marked ready: counted as an order, not as a wait.
		{CreatedAt: at(22, 9, 40)},
		// Before the period.
		{CreatedAt: at(19, 12, 0), ReadyAt: at(19, 12, 5)},
	}

	stats := BuildStats(tickets, now, beirut, 3)

	if stats.Orders != 5 || stats.ReadyOrders != 4 {
		t.Fatalf("orders = %d ready = %d, want 5 and 4", stats.Orders, stats.ReadyOrders)
	}
	if got := []string{stats.ByDay[0].Date, stats.ByDay[2].Date}; got[0] != "2026-11-20" || got[1] != "2026-11-22" {
		t.Fatalf("days = %v, want 2026-11-20 to 2026-11-22", got)
	}
	if stats.ByDay[0].Orders != 0 || stats.ByDay[1].Orders != 1 || stats.ByDay[2].Orders != 4 {
		t.Fatalf("by day = %+v", stats.ByDay)
	}
	if stats.ByHour[1] != 1 || stats.ByHour[9] != 4 {
		t.Fatalf("by hour 01 = %d, 09 = %d; want 1 and 4", stats.ByHour[1], stats.ByHour[9])
	}
	// Waits 2, 4, 6 and 10 minutes.
	if stats.AvgWait != 5*time.Minute+30*time.Second || stats.MedianWait != 5*time.Minute || stats.LongestWait != 10*time.Minute {
		t.Fatalf("avg %v median %v longest %v", stats.AvgWait, stats.MedianWait, stats.LongestWait)
	}
}

func TestStatsDaysStaysInRange(t *testing.T) {
	for requested, want := range map[int]int{0: DefaultStatsDays, -3: DefaultStatsDays, 7: 7, 500: MaxStatsDays} {
		if got := StatsDays(requested); got != want {
			t.Errorf("StatsDays(%d) = %d, want %d", requested, got, want)
		}
	}
}

func TestReviewLinksMustBeHTTPS(t *testing.T) {
	if got, err := NormalizeReviewURL("  https://g.page/r/abc/review "); err != nil || got != "https://g.page/r/abc/review" {
		t.Fatalf("valid link = %q, %v", got, err)
	}
	if got, err := NormalizeReviewURL(""); err != nil || got != "" {
		t.Fatalf("empty link = %q, %v; want it cleared", got, err)
	}
	for _, bad := range []string{"http://g.page/r/abc", "javascript:alert(1)", "g.page/r/abc", "https://user:pw@g.page/r"} {
		if _, err := NormalizeReviewURL(bad); err != ErrInvalidReviewURL {
			t.Errorf("NormalizeReviewURL(%q) err = %v, want ErrInvalidReviewURL", bad, err)
		}
	}
}
