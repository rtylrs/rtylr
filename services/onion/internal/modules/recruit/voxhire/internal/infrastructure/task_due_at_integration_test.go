package infrastructure_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/domain"
	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/infrastructure"
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

func TestVoxhire_ShortlistTimelineAndAnalyticsStayOwnerScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := domain.WithOwnerID(context.Background(), testutil.OwnerID)
	foreignCtx := domain.WithOwnerID(context.Background(), 1002)
	career, err := repo.CreateCareer(ctx, domain.CareerInput{Title: "Scoped Candidate Role", Details: "Synthetic test opening"})
	if err != nil {
		t.Fatal(err)
	}
	applicationID, err := repo.CreateCareerApplication(ctx, domain.CareerApplicationInput{
		JobID: career.ID, Name: "Scoped Candidate", Email: "scoped-candidate@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE i FROM voxhire_shortlist_items i JOIN voxhire_shortlists s ON s.id = i.shortlist_id WHERE s.job_id = ?", career.ID)
		_, _ = db.Exec("DELETE FROM voxhire_shortlists WHERE job_id = ?", career.ID)
		_, _ = db.Exec("DELETE FROM voxhire_application_events WHERE application_id = ?", applicationID)
		_, _ = db.Exec("DELETE FROM voxhire_follow_up_tasks WHERE application_id = ?", applicationID)
		_, _ = db.Exec("DELETE FROM voxire_career_applications WHERE id = ?", applicationID)
		_, _ = db.Exec("DELETE FROM careers WHERE id = ?", career.ID)
	})
	baseline, err := repo.GetAnalytics(ctx)
	if err != nil {
		t.Fatal(err)
	}
	list, err := repo.CreateShortlist(ctx, domain.ShortlistInput{JobID: career.ID, Name: "Interview", Status: "draft"})
	if err != nil {
		t.Fatal(err)
	}
	item, err := repo.AddShortlistItem(ctx, domain.ShortlistItemInput{ShortlistID: list.ID, ApplicationID: applicationID, Rank: 2, Notes: "Schedule interview"})
	if err != nil {
		t.Fatal(err)
	}
	if item.CandidateName != "Scoped Candidate" || item.Rank != 2 || item.Notes != "Schedule interview" {
		t.Fatalf("shortlist item did not round trip: %+v", item)
	}
	lists, err := repo.ListShortlists(ctx, career.ID)
	if err != nil || len(lists) != 1 || len(lists[0].Items) != 1 || lists[0].Items[0].ID != item.ID {
		t.Fatalf("list shortlist: %+v, %v", lists, err)
	}
	timeline, err := repo.ListApplicationTimeline(ctx, applicationID)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, event := range timeline {
		if event.Kind == "shortlist" {
			found = true
		}
	}
	if !found {
		t.Fatalf("timeline lacks shortlist event: %+v", timeline)
	}
	updated, err := repo.GetAnalytics(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Totals.ShortlistsTotal != baseline.Totals.ShortlistsTotal+1 || updated.Totals.ShortlistedCandidates != baseline.Totals.ShortlistedCandidates+1 {
		t.Fatalf("shortlist totals did not update: before=%+v after=%+v", baseline.Totals, updated.Totals)
	}
	foreignLists, err := repo.ListShortlists(foreignCtx, career.ID)
	if err != nil || len(foreignLists) != 0 {
		t.Fatalf("foreign shortlist leaked: %+v, %v", foreignLists, err)
	}
	if _, err := repo.ListApplicationTimeline(foreignCtx, applicationID); !errors.Is(err, domain.ErrApplicationNotFound) {
		t.Fatalf("foreign timeline access: %v", err)
	}
	if _, err := repo.AddShortlistItem(foreignCtx, domain.ShortlistItemInput{ShortlistID: list.ID, ApplicationID: applicationID}); !errors.Is(err, domain.ErrApplicationNotFound) {
		t.Fatalf("foreign shortlist mutation: %v", err)
	}
}
