package infrastructure_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_onboarding/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_onboarding/internal/infrastructure"
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
	if !errors.Is(err, domain.ErrTemplateNotFound) {
		t.Errorf("expected ErrTemplateNotFound, got %v", err)
	}
}

func TestCreateAndGetTemplate_RoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	db.ExecContext(ctx, "DELETE FROM hr_onboarding_checklist_templates WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Integration Test Onboarding")

	created, err := repo.CreateTemplate(ctx, testutil.OwnerID, domain.CreateTemplateParams{
		Name:        "Integration Test Onboarding",
		TriggerType: "hire",
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM hr_onboarding_checklist_templates WHERE id = ?", created.ID)
	})
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

	db.ExecContext(ctx, "DELETE FROM hr_onboarding_checklist_templates WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Template With Item")
	tmpl, err := repo.CreateTemplate(ctx, testutil.OwnerID, domain.CreateTemplateParams{
		Name:        "Template With Item",
		TriggerType: "hire",
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_onboarding_checklist_templates WHERE id = ?", tmpl.ID)
	})

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

	db.ExecContext(ctx, "DELETE FROM hr_onboarding_checklist_templates WHERE owner_id = ? AND name = ?", testutil.OwnerID, "List Items Test")
	tmpl, err := repo.CreateTemplate(ctx, testutil.OwnerID, domain.CreateTemplateParams{
		Name:        "List Items Test",
		TriggerType: "hire",
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_onboarding_checklist_templates WHERE id = ?", tmpl.ID)
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

	db.ExecContext(ctx, "DELETE FROM hr_onboarding_checklist_templates WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Trigger Test Template")
	tmpl, err := repo.CreateTemplate(ctx, testutil.OwnerID, domain.CreateTemplateParams{
		Name:        "Trigger Test Template",
		TriggerType: "hire",
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_onboarding_checklist_templates WHERE id = ?", tmpl.ID)
	})
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

func TestUpdateTemplate_NameChanges(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	db.ExecContext(ctx, "DELETE FROM hr_onboarding_checklist_templates WHERE owner_id = ? AND name IN (?,?)", testutil.OwnerID, "Before Update Template", "After Update Template")
	tmpl, err := repo.CreateTemplate(ctx, testutil.OwnerID, domain.CreateTemplateParams{
		Name:        "Before Update Template",
		TriggerType: "hire",
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_onboarding_checklist_templates WHERE id = ?", tmpl.ID)
	})

	err = repo.UpdateTemplate(ctx, testutil.OwnerID, tmpl.ID, domain.UpdateTemplateParams{
		Name: "After Update Template",
	})
	if err != nil {
		t.Fatalf("UpdateTemplate: %v", err)
	}

	got, err := repo.GetTemplate(ctx, testutil.OwnerID, tmpl.ID)
	if err != nil {
		t.Fatalf("GetTemplate after update: %v", err)
	}
	if got.Name != "After Update Template" {
		t.Errorf("Name = %q, want 'After Update Template'", got.Name)
	}
}

func TestGetTemplate_ReturnsOwnerID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	db.ExecContext(ctx, "DELETE FROM hr_onboarding_checklist_templates WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Owner ID Check Template")
	tmpl, err := repo.CreateTemplate(ctx, testutil.OwnerID, domain.CreateTemplateParams{
		Name:        "Owner ID Check Template",
		TriggerType: "hire",
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_onboarding_checklist_templates WHERE id = ?", tmpl.ID)
	})

	got, err := repo.GetTemplate(ctx, testutil.OwnerID, tmpl.ID)
	if err != nil {
		t.Fatalf("GetTemplate: %v", err)
	}
	if got.OwnerID != testutil.OwnerID {
		t.Errorf("OwnerID = %d, want %d", got.OwnerID, testutil.OwnerID)
	}
}

func TestCreateTemplateItem_DueOffsetDaysStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	db.ExecContext(ctx, "DELETE FROM hr_onboarding_checklist_templates WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Due Offset Template")
	tmpl, err := repo.CreateTemplate(ctx, testutil.OwnerID, domain.CreateTemplateParams{
		Name:        "Due Offset Template",
		TriggerType: "hire",
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_onboarding_checklist_templates WHERE id = ?", tmpl.ID)
	})

	item, err := repo.CreateTemplateItem(ctx, testutil.OwnerID, tmpl.ID, domain.CreateTemplateItemParams{
		Title:         "Offset Task",
		AssigneeRole:  "hr",
		DueOffsetDays: 7,
	})
	if err != nil {
		t.Fatalf("CreateTemplateItem: %v", err)
	}
	if item.DueOffsetDays != 7 {
		t.Errorf("DueOffsetDays = %d, want 7", item.DueOffsetDays)
	}
}

func TestGetEmployeeChecklist_NotFoundReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetEmployeeChecklist(ctx, testutil.OwnerID, 4101, 9999999)
	if !errors.Is(err, domain.ErrChecklistNotFound) {
		t.Errorf("expected ErrChecklistNotFound, got %v", err)
	}
}

func TestListEmployeeChecklists_CountMatchesTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	checklists, total, err := repo.ListEmployeeChecklists(ctx, testutil.OwnerID, 4101, 1000, 0)
	if err != nil {
		t.Fatalf("ListEmployeeChecklists: %v", err)
	}
	if int64(len(checklists)) > total {
		t.Errorf("returned %d checklists but total=%d", len(checklists), total)
	}
}

func TestUpdateEmployeeChecklistStatus_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	db.ExecContext(ctx, "DELETE FROM hr_onboarding_checklist_templates WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Status Update Template")
	tmpl, err := repo.CreateTemplate(ctx, testutil.OwnerID, domain.CreateTemplateParams{
		Name:        "Status Update Template",
		TriggerType: "hire",
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_onboarding_checklist_templates WHERE id = ?", tmpl.ID)
	})
	_, _ = repo.CreateTemplateItem(ctx, testutil.OwnerID, tmpl.ID, domain.CreateTemplateItemParams{
		Title:         "Status task",
		AssigneeRole:  "hr",
		DueOffsetDays: 1,
	})

	checklist, err := repo.TriggerChecklist(ctx, testutil.OwnerID, 4101, tmpl.ID, time.Now())
	if err != nil {
		t.Fatalf("TriggerChecklist: %v", err)
	}

	err = repo.UpdateEmployeeChecklistStatus(ctx, testutil.OwnerID, checklist.ID, "completed")
	if err != nil {
		t.Fatalf("UpdateEmployeeChecklistStatus: %v", err)
	}
}

func TestGetChecklistCompletionStats_TotalEqualsItemCount(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	db.ExecContext(ctx, "DELETE FROM hr_onboarding_checklist_templates WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Stats Template")
	tmpl, err := repo.CreateTemplate(ctx, testutil.OwnerID, domain.CreateTemplateParams{
		Name:        "Stats Template",
		TriggerType: "hire",
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_onboarding_checklist_templates WHERE id = ?", tmpl.ID)
	})
	for i := 0; i < 3; i++ {
		_, _ = repo.CreateTemplateItem(ctx, testutil.OwnerID, tmpl.ID, domain.CreateTemplateItemParams{
			Title:         "Stats task",
			AssigneeRole:  "hr",
			DueOffsetDays: int32(i + 1),
		})
	}

	checklist, err := repo.TriggerChecklist(ctx, testutil.OwnerID, 4104, tmpl.ID, time.Now())
	if err != nil {
		t.Fatalf("TriggerChecklist: %v", err)
	}

	total, completed, err := repo.GetChecklistCompletionStats(ctx, testutil.OwnerID, checklist.ID)
	if err != nil {
		t.Fatalf("GetChecklistCompletionStats: %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if completed != 0 {
		t.Errorf("completed = %d, want 0", completed)
	}
}

func TestListEmployeeChecklistItems_ReturnsItems(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	db.ExecContext(ctx, "DELETE FROM hr_onboarding_checklist_templates WHERE owner_id = ? AND name = ?", testutil.OwnerID, "List Items Checklist Template")
	tmpl, err := repo.CreateTemplate(ctx, testutil.OwnerID, domain.CreateTemplateParams{
		Name:        "List Items Checklist Template",
		TriggerType: "hire",
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_onboarding_checklist_templates WHERE id = ?", tmpl.ID)
	})
	_, _ = repo.CreateTemplateItem(ctx, testutil.OwnerID, tmpl.ID, domain.CreateTemplateItemParams{
		Title:         "List check task",
		AssigneeRole:  "hr",
		DueOffsetDays: 2,
	})

	checklist, err := repo.TriggerChecklist(ctx, testutil.OwnerID, 4105, tmpl.ID, time.Now())
	if err != nil {
		t.Fatalf("TriggerChecklist: %v", err)
	}

	items, err := repo.ListEmployeeChecklistItems(ctx, testutil.OwnerID, checklist.ID)
	if err != nil {
		t.Fatalf("ListEmployeeChecklistItems: %v", err)
	}
	if len(items) == 0 {
		t.Error("expected at least 1 checklist item")
	}
}

func TestCreateTemplate_TriggerTypeStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	db.ExecContext(ctx, "DELETE FROM hr_onboarding_checklist_templates WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Trigger Type Test")
	tmpl, err := repo.CreateTemplate(ctx, testutil.OwnerID, domain.CreateTemplateParams{
		Name:        "Trigger Type Test",
		TriggerType: "termination",
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_onboarding_checklist_templates WHERE id = ?", tmpl.ID)
	})
	if tmpl.TriggerType != "termination" {
		t.Errorf("TriggerType = %q, want 'termination'", tmpl.TriggerType)
	}
}

func TestGetLatestEmployeeChecklist_ResolvesChecklistByEmployee(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	tmpl, err := repo.CreateTemplate(ctx, testutil.OwnerID, domain.CreateTemplateParams{
		Name:        "Latest Checklist Template",
		TriggerType: "hire",
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(), "DELETE FROM hr_onboarding_checklist_templates WHERE id = ?", tmpl.ID); err != nil {
			t.Errorf("cleanup template: %v", err)
		}
	})
	if _, err := repo.CreateTemplateItem(ctx, testutil.OwnerID, tmpl.ID, domain.CreateTemplateItemParams{
		Title:         "Latest task",
		AssigneeRole:  "employee",
		DueOffsetDays: 1,
	}); err != nil {
		t.Fatalf("CreateTemplateItem: %v", err)
	}

	const employeeID int64 = 4110
	if _, err := repo.TriggerChecklist(ctx, testutil.OwnerID, employeeID, tmpl.ID, time.Now().Add(-time.Hour)); err != nil {
		t.Fatalf("TriggerChecklist (older): %v", err)
	}
	latest, err := repo.TriggerChecklist(ctx, testutil.OwnerID, employeeID, tmpl.ID, time.Now())
	if err != nil {
		t.Fatalf("TriggerChecklist (latest): %v", err)
	}

	got, err := repo.GetLatestEmployeeChecklist(ctx, testutil.OwnerID, employeeID)
	if err != nil {
		t.Fatalf("GetLatestEmployeeChecklist: %v", err)
	}
	if got.ID != latest.ID {
		t.Errorf("checklist id = %d, want the latest %d", got.ID, latest.ID)
	}
	if got.ID == employeeID {
		t.Errorf("employee id %d was used as the checklist id", employeeID)
	}
	if got.EmployeeID != employeeID {
		t.Errorf("employee id = %d, want %d", got.EmployeeID, employeeID)
	}
	if len(got.Items) != 1 {
		t.Errorf("items = %d, want the 1 item copied from the template", len(got.Items))
	}
}

func TestGetLatestEmployeeChecklist_NoChecklistReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetLatestEmployeeChecklist(ctx, testutil.OwnerID, 419999999)
	if !errors.Is(err, domain.ErrChecklistNotFound) {
		t.Errorf("expected ErrChecklistNotFound, got %v", err)
	}
}

func TestTriggerChecklist_StatusIsInProgress(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	db.ExecContext(ctx, "DELETE FROM hr_onboarding_checklist_templates WHERE owner_id = ? AND name = ?", testutil.OwnerID, "In Progress Status Template")
	tmpl, err := repo.CreateTemplate(ctx, testutil.OwnerID, domain.CreateTemplateParams{
		Name:        "In Progress Status Template",
		TriggerType: "hire",
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_onboarding_checklist_templates WHERE id = ?", tmpl.ID)
	})
	_, _ = repo.CreateTemplateItem(ctx, testutil.OwnerID, tmpl.ID, domain.CreateTemplateItemParams{
		Title:         "In progress task",
		AssigneeRole:  "manager",
		DueOffsetDays: 5,
	})

	checklist, err := repo.TriggerChecklist(ctx, testutil.OwnerID, 4106, tmpl.ID, time.Now())
	if err != nil {
		t.Fatalf("TriggerChecklist: %v", err)
	}
	if checklist.Status != "in_progress" {
		t.Errorf("Status = %q, want 'in_progress'", checklist.Status)
	}
}
