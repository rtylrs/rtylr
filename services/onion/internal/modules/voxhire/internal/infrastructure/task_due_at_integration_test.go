package infrastructure_test

import (
	"context"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/voxhire/internal/domain"
	"github.com/voxire/onion/internal/modules/voxhire/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// A partial task update that omits due_at must keep the stored follow-up date,
// while an explicitly supplied empty due date still clears it.
func TestVoxhire_UpdateTaskKeepsDueAtWhenOmitted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := domain.WithOwnerID(context.Background(), testutil.OwnerID)

	career, err := repo.CreateCareer(ctx, domain.CareerInput{
		Title:   "Due Date Regression Role",
		Details: "Integration test opening for partial task updates",
	})
	if err != nil {
		t.Fatalf("CreateCareer: %v", err)
	}
	appID, err := repo.CreateCareerApplication(ctx, domain.CareerApplicationInput{
		JobID: career.ID,
		Name:  "Due Date Candidate",
		Email: "due-date-candidate@example.com",
	})
	if err != nil {
		t.Fatalf("CreateCareerApplication: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx := context.Background()
		_, _ = db.ExecContext(cleanupCtx, "DELETE FROM voxhire_follow_up_tasks WHERE application_id = ?", appID)
		_, _ = db.ExecContext(cleanupCtx, "DELETE FROM voxhire_application_events WHERE application_id = ?", appID)
		_, _ = db.ExecContext(cleanupCtx, "DELETE FROM voxire_career_applications WHERE id = ?", appID)
		_, _ = db.ExecContext(cleanupCtx, "DELETE FROM careers WHERE id = ?", career.ID)
	})

	dueAt := time.Now().Add(72 * time.Hour).UTC().Truncate(time.Second)
	task, err := repo.UpsertTask(ctx, domain.TaskInput{
		ApplicationID: appID,
		Title:         "Call the candidate",
		DueAt:         &dueAt,
	})
	if err != nil {
		t.Fatalf("UpsertTask: %v", err)
	}
	if task.DueAt == nil {
		t.Fatal("UpsertTask did not store the due date")
	}

	updated, err := repo.UpdateTask(ctx, domain.TaskInput{ID: task.ID, Status: "done"})
	if err != nil {
		t.Fatalf("status-only UpdateTask: %v", err)
	}
	if updated.Status != "done" {
		t.Fatalf("status = %q, want done", updated.Status)
	}
	if updated.DueAt == nil {
		t.Fatal("status-only update cleared the stored due date")
	}
	if !updated.DueAt.Equal(dueAt) {
		t.Fatalf("due date = %v, want %v", updated.DueAt, dueAt)
	}

	cleared, err := repo.UpdateTask(ctx, domain.TaskInput{ID: task.ID, DueAtSet: true})
	if err != nil {
		t.Fatalf("clearing UpdateTask: %v", err)
	}
	if cleared.DueAt != nil {
		t.Fatalf("due date = %v, want cleared", cleared.DueAt)
	}
}
