package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/flow_insights/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestGetInsightsStats_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	stats, err := repo.GetInsightsStats(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetInsightsStats: %v", err)
	}
	if stats.TasksByStatus == nil {
		t.Error("expected non-nil TasksByStatus map")
	}
	if stats.TasksByPriority == nil {
		t.Error("expected non-nil TasksByPriority map")
	}
}

func TestGetInsightsStats_CompletionRateInRange(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	stats, err := repo.GetInsightsStats(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetInsightsStats: %v", err)
	}
	if stats.CompletionRate < 0 || stats.CompletionRate > 100 {
		t.Errorf("CompletionRate out of range [0,100]: %f", stats.CompletionRate)
	}
}

func TestGetInsightsStats_TotalTasksConsistent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	stats, err := repo.GetInsightsStats(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetInsightsStats: %v", err)
	}
	if stats.CompletedTasks > stats.TotalTasks {
		t.Errorf("CompletedTasks (%d) > TotalTasks (%d)", stats.CompletedTasks, stats.TotalTasks)
	}
}

func TestGetInsightsStats_OverdueCountNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	stats, err := repo.GetInsightsStats(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetInsightsStats: %v", err)
	}
	if stats.OverdueCount < 0 {
		t.Errorf("OverdueCount should be >= 0, got %d", stats.OverdueCount)
	}
}

func TestGetInsightsStats_UnknownOwnerNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	stats, err := repo.GetInsightsStats(ctx, 0)
	if err != nil {
		t.Fatalf("GetInsightsStats with unknown owner: %v", err)
	}
	if stats.TotalTasks != 0 {
		t.Errorf("expected 0 total tasks for unknown owner, got %d", stats.TotalTasks)
	}
}

func TestGetInsightsStats_VelocityNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	stats, err := repo.GetInsightsStats(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetInsightsStats: %v", err)
	}
	if stats.VelocityThisWeek < 0 {
		t.Errorf("VelocityThisWeek should be >= 0, got %d", stats.VelocityThisWeek)
	}
	if stats.VelocityLastWeek < 0 {
		t.Errorf("VelocityLastWeek should be >= 0, got %d", stats.VelocityLastWeek)
	}
}

func TestGetTeamWorkload_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	result, err := repo.GetTeamWorkload(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetTeamWorkload: %v", err)
	}
	if result.Members == nil {
		t.Error("expected non-nil Members slice")
	}
}

func TestGetTeamWorkload_MembersHaveValidFields(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	result, err := repo.GetTeamWorkload(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetTeamWorkload: %v", err)
	}
	for _, m := range result.Members {
		if m.UserID == 0 {
			t.Error("member has zero UserID")
		}
		if m.TotalTasks < 0 {
			t.Errorf("member %d has negative TotalTasks", m.UserID)
		}
		if m.CompletedTasks > m.TotalTasks {
			t.Errorf("member %d: CompletedTasks > TotalTasks", m.UserID)
		}
	}
}

func TestGetTeamWorkload_UnknownOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	result, err := repo.GetTeamWorkload(ctx, 0)
	if err != nil {
		t.Fatalf("GetTeamWorkload unknown owner: %v", err)
	}
	if len(result.Members) != 0 {
		t.Errorf("expected 0 members for unknown owner, got %d", len(result.Members))
	}
}

func TestGetRoadmap_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	result, err := repo.GetRoadmap(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetRoadmap: %v", err)
	}
	if result.Projects == nil {
		t.Error("expected non-nil Projects slice")
	}
}

func TestGetRoadmap_ProjectsHaveMilestoneSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	result, err := repo.GetRoadmap(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetRoadmap: %v", err)
	}
	for _, p := range result.Projects {
		if p.Milestones == nil {
			t.Errorf("project %d has nil Milestones slice", p.ID)
		}
		if p.Progress < 0 || p.Progress > 100 {
			t.Errorf("project %d progress out of range: %f", p.ID, p.Progress)
		}
	}
}

func TestGetRoadmap_UnknownOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	result, err := repo.GetRoadmap(ctx, 0)
	if err != nil {
		t.Fatalf("GetRoadmap unknown owner: %v", err)
	}
	if len(result.Projects) != 0 {
		t.Errorf("expected 0 projects for unknown owner, got %d", len(result.Projects))
	}
}
