package domain

import (
	"encoding/json"
	"testing"

	"github.com/voxire/onion/internal/platform/money"
)

func TestDashboardDataRevenueUsesExactMoneyJSON(t *testing.T) {
	data := DashboardData{Revenue: money.Amount(123456)}

	payload, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal dashboard data: %v", err)
	}

	const want = `{"today_count":0,"upcoming_count":0,"cancellation_rate":0,"revenue":1234.56}`
	if string(payload) != want {
		t.Fatalf("dashboard JSON = %s, want %s", payload, want)
	}
}
