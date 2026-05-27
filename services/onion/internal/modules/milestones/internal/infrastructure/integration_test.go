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

func TestCreateMilestone_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.MilestoneInput{
		OwnerID:   testutil.OwnerID,
		ProjectID: 1,
		Name:      "Positive ID Milestone",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestCreateMilestone_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const want = "NameStoredMilestone"
	id, err := repo.Create(ctx, domain.MilestoneInput{
		OwnerID:   testutil.OwnerID,
		ProjectID: 1,
		Name:      want,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	milestones, err := repo.List(ctx, testutil.OwnerID, nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, m := range milestones {
		if m.ID == id {
			if m.Name != want {
				t.Errorf("Name = %q, want %q", m.Name, want)
			}
			return
		}
	}
	t.Errorf("milestone id=%d not found", id)
}

func TestCreateMilestone_DescriptionIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const want = "Detailed description for integration test"
	id, err := repo.Create(ctx, domain.MilestoneInput{
		OwnerID:     testutil.OwnerID,
		ProjectID:   1,
		Name:        "DescriptionMilestone",
		Description: want,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	milestones, err := repo.List(ctx, testutil.OwnerID, nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, m := range milestones {
		if m.ID == id {
			if m.Description != want {
				t.Errorf("Description = %q, want %q", m.Description, want)
			}
			return
		}
	}
	t.Errorf("milestone id=%d not found", id)
}

func TestCreateMilestone_ProjectIDIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const projectID = int64(1)
	id, err := repo.Create(ctx, domain.MilestoneInput{
		OwnerID:   testutil.OwnerID,
		ProjectID: projectID,
		Name:      "ProjectIDMilestone",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	milestones, err := repo.List(ctx, testutil.OwnerID, nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, m := range milestones {
		if m.ID == id {
			if m.ProjectID != projectID {
				t.Errorf("ProjectID = %d, want %d", m.ProjectID, projectID)
			}
			return
		}
	}
	t.Errorf("milestone id=%d not found", id)
}

func TestListMilestones_AfterCreate_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.List(ctx, testutil.OwnerID, nil)
	if err != nil {
		t.Fatalf("List before: %v", err)
	}

	_, err = repo.Create(ctx, domain.MilestoneInput{
		OwnerID:   testutil.OwnerID,
		ProjectID: 1,
		Name:      "CountMilestone",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	after, err := repo.List(ctx, testutil.OwnerID, nil)
	if err != nil {
		t.Fatalf("List after: %v", err)
	}
	if len(after) != len(before)+1 {
		t.Errorf("expected count +1: before=%d after=%d", len(before), len(after))
	}
}

func TestUpdateMilestone_ChangesDescription(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.MilestoneInput{
		OwnerID:   testutil.OwnerID,
		ProjectID: 1,
		Name:      "DescUpdateMilestone",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	const newDesc = "Updated description"
	if err := repo.Update(ctx, domain.MilestoneUpdateInput{
		ID:          id,
		OwnerID:     testutil.OwnerID,
		Name:        "DescUpdateMilestone",
		Description: newDesc,
		Status:      "in_progress",
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	milestones, err := repo.List(ctx, testutil.OwnerID, nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, m := range milestones {
		if m.ID == id {
			if m.Description != newDesc {
				t.Errorf("Description = %q, want %q", m.Description, newDesc)
			}
			return
		}
	}
	t.Errorf("milestone id=%d not found after update", id)
}

func TestUpdateMilestone_ChangesStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.MilestoneInput{
		OwnerID:   testutil.OwnerID,
		ProjectID: 1,
		Name:      "StatusUpdateMilestone",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(ctx, domain.MilestoneUpdateInput{
		ID:      id,
		OwnerID: testutil.OwnerID,
		Name:    "StatusUpdateMilestone",
		Status:  "completed",
	}); err != nil {
		t.Fatalf("Update status: %v", err)
	}

	milestones, err := repo.List(ctx, testutil.OwnerID, nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, m := range milestones {
		if m.ID == id {
			if m.Status != "completed" {
				t.Errorf("Status = %q, want completed", m.Status)
			}
			return
		}
	}
	t.Errorf("milestone id=%d not found after status update", id)
}

func TestListMilestones_EachHasNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.Create(ctx, domain.MilestoneInput{
		OwnerID:   testutil.OwnerID,
		ProjectID: 1,
		Name:      "IDCheck",
	})

	milestones, err := repo.List(ctx, testutil.OwnerID, nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, m := range milestones {
		if m.ID == 0 {
			t.Errorf("milestones[%d].ID is zero", i)
		}
	}
}

func TestCreateMilestone_IDsAreUnique(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id1, err := repo.Create(ctx, domain.MilestoneInput{OwnerID: testutil.OwnerID, ProjectID: 1, Name: "UniqueA"})
	if err != nil {
		t.Fatalf("Create A: %v", err)
	}
	id2, err := repo.Create(ctx, domain.MilestoneInput{OwnerID: testutil.OwnerID, ProjectID: 1, Name: "UniqueB"})
	if err != nil {
		t.Fatalf("Create B: %v", err)
	}
	if id1 == id2 {
		t.Errorf("expected unique IDs, both got %d", id1)
	}
}

func TestListMilestones_WithProjectFilter_ReturnsAll(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	projectID := int64(2)
	_, err := repo.Create(ctx, domain.MilestoneInput{
		OwnerID:   testutil.OwnerID,
		ProjectID: projectID,
		Name:      "FilterProjectMilestone",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	milestones, err := repo.List(ctx, testutil.OwnerID, &projectID)
	if err != nil {
		t.Fatalf("List with project filter: %v", err)
	}
	_ = milestones
}

func TestUpdateMilestone_WithTargetDate_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.MilestoneInput{
		OwnerID:   testutil.OwnerID,
		ProjectID: 1,
		Name:      "TargetDateUpdate",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	target := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	if err := repo.Update(ctx, domain.MilestoneUpdateInput{
		ID:         id,
		OwnerID:    testutil.OwnerID,
		Name:       "TargetDateUpdate",
		Status:     "upcoming",
		TargetDate: &target,
	}); err != nil {
		t.Fatalf("Update with target date: %v", err)
	}
}

func TestListMilestones_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	milestones, err := repo.List(ctx, 0, nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(milestones) != 0 {
		t.Errorf("expected 0 for owner_id=0, got %d", len(milestones))
	}
}
