package api

import (
	"testing"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
)

func TestToHoursSummaryResponse(t *testing.T) {
	got := toHoursSummaryResponse(domain.HoursSummary{
		Logged:       domain.CountCard{Value: 2_400, DeltaPct: 50},
		PriorMinutes: 1_600,
		Projects:     []domain.ProjectHours{{ProjectID: 7, Name: "Website", Minutes: 2_070}},
	})
	if got.Logged == nil || got.Logged.Value != 2_400 || got.Logged.DeltaPct != 50 || got.PriorMinutes != 1_600 {
		t.Errorf("response = %+v, want 2400 minutes up 50%% on 1600", got)
	}
	if len(got.Projects) != 1 || got.Projects[0].ProjectId != 7 || got.Projects[0].ProjectName != "Website" || got.Projects[0].Minutes != 2_070 {
		t.Errorf("Projects = %+v, want Website with 2070 minutes", got.Projects)
	}
}

// A business without time logs still gets a logged card and an empty list, so
// a client reading logged.value or iterating projects never meets null.
func TestToHoursSummaryResponseWithoutTime(t *testing.T) {
	got := toHoursSummaryResponse(domain.HoursSummary{})
	if got.Logged == nil {
		t.Fatal("Logged = nil, want an empty card")
	}
	if got.Projects == nil {
		t.Fatal("Projects = nil, want an empty slice")
	}
	if got.Logged.Value != 0 || got.Logged.DeltaPct != 0 || got.PriorMinutes != 0 || len(got.Projects) != 0 {
		t.Errorf("response = %+v, want zeros", got)
	}
}
