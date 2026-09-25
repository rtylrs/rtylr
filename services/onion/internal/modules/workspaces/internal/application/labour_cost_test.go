package application

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/platform/labourcost"
)

type labourCostStub struct {
	gotBusinessID int64
	gotProjectID  int64
	result        labourcost.ProjectCost
}

func (s *labourCostStub) GetProjectLabourCost(_ context.Context, businessID, projectID int64) (labourcost.ProjectCost, error) {
	s.gotBusinessID = businessID
	s.gotProjectID = projectID
	return s.result, nil
}

func TestGetProjectLabourCostUsesTenantScopedProvider(t *testing.T) {
	service := NewService(context.Background(), nil)
	stub := &labourCostStub{result: labourcost.ProjectCost{BusinessID: 9, ProjectID: 42, CostCents: 12500, CostedMinutes: 60, UncostedMinutes: 30}}
	service.SetLabourCostProvider(stub)

	got, err := service.GetProjectLabourCost(context.Background(), 9, 42)
	if err != nil {
		t.Fatalf("GetProjectLabourCost: %v", err)
	}
	if stub.gotBusinessID != 9 || stub.gotProjectID != 42 {
		t.Fatalf("provider scope = business %d, project %d; want 9, 42", stub.gotBusinessID, stub.gotProjectID)
	}
	if got.CostCents != 12500 || got.UncostedMinutes != 30 {
		t.Fatalf("unexpected aggregate: %+v", got)
	}
}

func TestGetProjectLabourCostDefaultsToEmptyAggregate(t *testing.T) {
	service := NewService(context.Background(), nil)
	got, err := service.GetProjectLabourCost(context.Background(), 9, 42)
	if err != nil {
		t.Fatalf("GetProjectLabourCost: %v", err)
	}
	if got.BusinessID != 9 || got.ProjectID != 42 || got.CostCents != 0 {
		t.Fatalf("unexpected no-op aggregate: %+v", got)
	}
}
