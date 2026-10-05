package infrastructure_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_onboarding/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_onboarding/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCompleteChecklistItemConcurrentRequestsHaveOneWinner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	stamp := time.Now().UnixNano()

	tmpl, err := repo.CreateTemplate(ctx, testutil.BusinessID, domain.CreateTemplateParams{
		Name: fmt.Sprintf("concurrent completion %d", stamp), TriggerType: "hire",
	})
	if err != nil {
		t.Fatalf("create template: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_onboarding_checklist_templates WHERE id = ?", tmpl.ID)
	})
	if _, err := repo.CreateTemplateItem(ctx, testutil.BusinessID, tmpl.ID, domain.CreateTemplateItemParams{
		Title: "Concurrent task", AssigneeRole: "employee",
	}); err != nil {
		t.Fatalf("create template item: %v", err)
	}

	checklist, err := repo.TriggerChecklist(ctx, testutil.BusinessID, 7_700_000_000+stamp%100_000_000, tmpl.ID, time.Now())
	if err != nil {
		t.Fatalf("trigger checklist: %v", err)
	}
	if len(checklist.Items) != 1 {
		t.Fatalf("checklist items = %d, want 1", len(checklist.Items))
	}
	itemID := checklist.Items[0].ID
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_onboarding_employee_checklist_items WHERE employee_checklist_id = ?", checklist.ID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_onboarding_employee_checklists WHERE id = ?", checklist.ID)
	})

	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for i := int64(0); i < 2; i++ {
		wg.Add(1)
		go func(completedBy int64) {
			defer wg.Done()
			<-start
			results <- repo.CompleteChecklistItemWithAttachment(ctx, testutil.BusinessID, itemID, completedBy, nil)
		}(8_600_000_000 + i)
	}
	close(start)
	wg.Wait()
	close(results)

	var successes, conflicts int
	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, domain.ErrChecklistAlreadyCompleted):
			conflicts++
		default:
			t.Fatalf("unexpected concurrent completion error: %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("concurrent completion results: successes=%d conflicts=%d, want one each", successes, conflicts)
	}
}
