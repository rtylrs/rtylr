package infrastructure_test

import (
	"context"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/milestones/internal/domain"
	"github.com/voxire/onion/internal/modules/milestones/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListMilestones_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	milestones, err := repo.List(ctx, testutil.OwnerID, nil)
	if err != nil {
		t.Fatalf("List milestones: %v", err)
	}
	_ = milestones
}

func TestListMilestones_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	milestones, err := repo.List(ctx, testutil.OwnerID, nil)
	if err != nil {
		t.Fatalf("List milestones: %v", err)
	}
	for _, m := range milestones {
		if m.OwnerID != testutil.OwnerID {
			t.Errorf("milestone %d has owner_id %d, want %d", m.ID, m.OwnerID, testutil.OwnerID)
		}
	}
}

func TestCreateMilestone_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.MilestoneInput{
		OwnerID:     testutil.OwnerID,
		ProjectID:   1,
		Name:        "Test Milestone",
		Description: "Integration test milestone",
	})
	if err != nil {
		t.Fatalf("Create milestone: %v", err)
	}
	if id == 0 {
		t.Fatal("Create returned id=0")
	}
}

func TestCreateMilestone_WithTargetDate_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	target := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	id, err := repo.Create(ctx, domain.MilestoneInput{
		OwnerID:    testutil.OwnerID,
		ProjectID:  1,
		Name:       "Test Milestone With Date",
		TargetDate: &target,
	})
	if err != nil {
		t.Fatalf("Create milestone with target date: %v", err)
	}
	if id == 0 {
		t.Fatal("Create returned id=0")
	}
}

func TestCreateMilestone_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	name := "Appears In List Test"
	id, err := repo.Create(ctx, domain.MilestoneInput{
		OwnerID:   testutil.OwnerID,
		ProjectID: 1,
		Name:      name,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	milestones, err := repo.List(ctx, testutil.OwnerID, nil)
	if err != nil {
		t.Fatalf("List after create: %v", err)
	}
	found := false
	for _, m := range milestones {
		if m.ID == id {
			found = true
			if m.Name != name {
				t.Errorf("milestone name = %q, want %q", m.Name, name)
			}
			break
		}
	}
	if !found {
		t.Errorf("created milestone id %d not found in list", id)
	}
}

func TestListMilestonesByProject_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	projectID := int64(1)
	milestones, err := repo.List(ctx, testutil.OwnerID, &projectID)
	if err != nil {
		t.Fatalf("List milestones by project: %v", err)
	}
	_ = milestones
}

func TestListMilestonesByProject_AllMatchProjectID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	projectID := int64(1)
	milestones, err := repo.List(ctx, testutil.OwnerID, &projectID)
	if err != nil {
		t.Fatalf("List milestones by project: %v", err)
	}
	for _, m := range milestones {
		if m.ProjectID != projectID {
			t.Errorf("milestone %d has project_id %d, want %d", m.ID, m.ProjectID, projectID)
		}
	}
}

func TestUpdateMilestone_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.MilestoneInput{
		OwnerID:   testutil.OwnerID,
		ProjectID: 1,
		Name:      "Update Test Milestone",
	})
	if err != nil {
		t.Fatalf("Create for update test: %v", err)
	}

	err = repo.Update(ctx, domain.MilestoneUpdateInput{
		ID:      id,
		OwnerID: testutil.OwnerID,
		Name:    "Updated Milestone Name",
		Status:  "in_progress",
	})
	if err != nil {
		t.Fatalf("Update milestone: %v", err)
	}
}

func TestUpdateMilestone_ReflectedInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.MilestoneInput{
		OwnerID:   testutil.OwnerID,
		ProjectID: 1,
		Name:      "Reflect Update Test",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	newName := "Reflected Updated Name"
	if err := repo.Update(ctx, domain.MilestoneUpdateInput{
		ID:      id,
		OwnerID: testutil.OwnerID,
		Name:    newName,
		Status:  "upcoming",
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	milestones, err := repo.List(ctx, testutil.OwnerID, nil)
	if err != nil {
		t.Fatalf("List after update: %v", err)
	}
	for _, m := range milestones {
		if m.ID == id {
			if m.Name != newName {
				t.Errorf("updated name = %q, want %q", m.Name, newName)
			}
			return
		}
	}
	t.Errorf("updated milestone %d not found in list", id)
}

func TestUpdateMilestone_NonExistent_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.Update(ctx, domain.MilestoneUpdateInput{
		ID:      9999999,
		OwnerID: testutil.OwnerID,
		Name:    "Ghost Milestone",
		Status:  "upcoming",
	})
	if err == nil {
		t.Fatal("expected error updating non-existent milestone, got nil")
	}
}

func TestListMilestones_UnknownOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	milestones, err := repo.List(ctx, 9999999, nil)
	if err != nil {
		t.Fatalf("List milestones unknown owner: %v", err)
	}
	if len(milestones) != 0 {
		t.Errorf("expected 0 milestones for unknown owner, got %d", len(milestones))
	}
}
