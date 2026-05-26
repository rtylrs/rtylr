package infrastructure_test

import (
	"context"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListTemplates_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	templates, total, err := repo.ListTemplates(ctx, testutil.OwnerID, 50, 0)
	if err != nil {
		t.Fatalf("ListTemplates: %v", err)
	}
	if total < 0 {
		t.Errorf("expected non-negative total, got %d", total)
	}
	_ = templates
}

func TestListTemplates_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	templates, _, err := repo.ListTemplates(ctx, testutil.OwnerID, 50, 0)
	if err != nil {
		t.Fatalf("ListTemplates: %v", err)
	}
	for _, tmpl := range templates {
		if tmpl.OwnerID != testutil.OwnerID {
			t.Errorf("template %d belongs to wrong owner: %d", tmpl.ID, tmpl.OwnerID)
		}
	}
}

func TestListTemplates_UnknownOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	templates, total, err := repo.ListTemplates(ctx, 0, 50, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 || len(templates) != 0 {
		t.Errorf("expected empty result for unknown owner, got total=%d templates=%d", total, len(templates))
	}
}

func TestListTemplates_CountMatchesRecordBound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	templates, total, err := repo.ListTemplates(ctx, testutil.OwnerID, 50, 0)
	if err != nil {
		t.Fatalf("ListTemplates: %v", err)
	}
	if int64(len(templates)) > total {
		t.Errorf("returned %d templates but total reported as %d", len(templates), total)
	}
}

func TestGetTemplate_NotFoundReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetTemplate(ctx, testutil.OwnerID, 999999999)
	if err != domain.ErrTemplateNotFound {
		t.Errorf("expected ErrTemplateNotFound, got %v", err)
	}
}

func TestCreateAndGetTemplate_RoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateTemplate(ctx, testutil.OwnerID, domain.CreateTemplateParams{
		Name:        "Integration Test Onboarding",
		TriggerType: "hire",
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero ID after create")
	}
	if created.Name != "Integration Test Onboarding" {
		t.Errorf("name mismatch: got %q", created.Name)
	}
}

func TestCreateTemplateItem_AddsItemToTemplate(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	tmpl, err := repo.CreateTemplate(ctx, testutil.OwnerID, domain.CreateTemplateParams{
		Name:        "Template With Item",
		TriggerType: "hire",
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}

	item, err := repo.CreateTemplateItem(ctx, testutil.OwnerID, tmpl.ID, domain.CreateTemplateItemParams{
		Title:         "Complete paperwork",
		AssigneeRole:  "hr",
		DueOffsetDays: 3,
	})
	if err != nil {
		t.Fatalf("CreateTemplateItem: %v", err)
	}
	if item.ID == 0 {
		t.Fatal("expected non-zero ID on checklist item")
	}
	if item.Title != "Complete paperwork" {
		t.Errorf("title mismatch: got %q", item.Title)
	}
}

func TestListTemplateItems_ReturnsItemsForTemplate(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	tmpl, _ := repo.CreateTemplate(ctx, testutil.OwnerID, domain.CreateTemplateParams{
		Name:        "List Items Test",
		TriggerType: "hire",
	})
	_, _ = repo.CreateTemplateItem(ctx, testutil.OwnerID, tmpl.ID, domain.CreateTemplateItemParams{
		Title:         "Task A",
		AssigneeRole:  "hr",
		DueOffsetDays: 1,
	})

	items, err := repo.ListTemplateItems(ctx, testutil.OwnerID, tmpl.ID)
	if err != nil {
		t.Fatalf("ListTemplateItems: %v", err)
	}
	if len(items) == 0 {
		t.Error("expected at least 1 item")
	}
}

func TestTriggerChecklist_CreatesEmployeeChecklist(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	tmpl, err := repo.CreateTemplate(ctx, testutil.OwnerID, domain.CreateTemplateParams{
		Name:        "Trigger Test Template",
		TriggerType: "hire",
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	_, _ = repo.CreateTemplateItem(ctx, testutil.OwnerID, tmpl.ID, domain.CreateTemplateItemParams{
		Title:         "Welcome task",
		AssigneeRole:  "employee",
		DueOffsetDays: 7,
	})

	checklist, err := repo.TriggerChecklist(ctx, testutil.OwnerID, 4101, tmpl.ID, time.Now())
	if err != nil {
		t.Fatalf("TriggerChecklist: %v", err)
	}
	if checklist.ID == 0 {
		t.Fatal("expected non-zero checklist ID")
	}
	if checklist.EmployeeID != 4101 {
		t.Errorf("expected employee 4101, got %d", checklist.EmployeeID)
	}
	if len(checklist.Items) == 0 {
		t.Error("expected at least 1 checklist item after trigger")
	}
}

func TestListEmployeeChecklists_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	checklists, total, err := repo.ListEmployeeChecklists(ctx, testutil.OwnerID, 4101, 50, 0)
	if err != nil {
		t.Fatalf("ListEmployeeChecklists: %v", err)
	}
	if total < 0 {
		t.Errorf("expected non-negative total, got %d", total)
	}
	_ = checklists
}

func TestListTemplates_PaginationOffsetWorks(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	first, total, err := repo.ListTemplates(ctx, testutil.OwnerID, 2, 0)
	if err != nil {
		t.Fatalf("ListTemplates page 1: %v", err)
	}
	second, _, err := repo.ListTemplates(ctx, testutil.OwnerID, 2, 2)
	if err != nil {
		t.Fatalf("ListTemplates page 2: %v", err)
	}
	if total > 2 && len(first) > 0 && len(second) > 0 {
		if first[0].ID == second[0].ID {
			t.Error("pagination offset had no effect")
		}
	}
}
