package infrastructure_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/voxhire/internal/domain"
	"github.com/voxire/onion/internal/modules/voxhire/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// Permanently failing CVs are still retried, but a backlog of them must not
// push newer pending CVs out of every batch.
func TestVoxhire_ListPendingCVProcessingTargetsOrdersPendingBeforeErrored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := domain.WithOwnerID(context.Background(), testutil.OwnerID)

	career, err := repo.CreateCareer(ctx, domain.CareerInput{
		Title:   "CV Queue Regression Role",
		Details: "Integration test opening for the CV extraction queue",
	})
	if err != nil {
		t.Fatalf("CreateCareer: %v", err)
	}
	appIDs := make([]int64, 0, 14)
	t.Cleanup(func() {
		cleanupCtx := context.Background()
		for _, id := range appIDs {
			_, _ = db.ExecContext(cleanupCtx, "DELETE FROM voxire_career_applications WHERE id = ?", id)
		}
		_, _ = db.ExecContext(cleanupCtx, "DELETE FROM careers WHERE id = ?", career.ID)
	})

	newApplication := func(name, cvKey string) int64 {
		t.Helper()
		id, err := repo.CreateCareerApplication(ctx, domain.CareerApplicationInput{
			JobID: career.ID,
			Name:  name,
			Email: fmt.Sprintf("%s@example.com", name),
			CVKey: cvKey,
		})
		if err != nil {
			t.Fatalf("CreateCareerApplication %s: %v", name, err)
		}
		appIDs = append(appIDs, id)
		return id
	}

	// 13 permanently failing targets, older than the pending one: enough to
	// fill a full 25 target batch once they are retried first.
	erroredIDs := make(map[int64]struct{}, 13)
	older := time.Now().UTC().Add(-time.Hour)
	for i := 0; i < 13; i++ {
		id := newApplication(fmt.Sprintf("cv-queue-errored-%d", i), fmt.Sprintf("voxhire/applications/cv-queue/legacy-%d.doc", i))
		erroredIDs[id] = struct{}{}
		if _, err := db.ExecContext(ctx, `
			UPDATE voxire_career_applications
			   SET cv_processing_status = 'error',
			       cv_processing_error = 'legacy .doc CV parsing is not supported',
			       cv_processed_at = CURRENT_TIMESTAMP,
			       created_at = ?
			 WHERE id = ?`, older.Add(time.Duration(i)*time.Minute), id); err != nil {
			t.Fatalf("mark application %d errored: %v", id, err)
		}
	}
	pendingID := newApplication("cv-queue-pending", "voxhire/applications/cv-queue/resume-pending.pdf")

	targets, err := repo.ListPendingCVProcessingTargets(ctx, 100)
	if err != nil {
		t.Fatalf("ListPendingCVProcessingTargets: %v", err)
	}

	pendingPos := -1
	erroredPos := make(map[int64]int, len(erroredIDs))
	for i, target := range targets {
		if target.ID == pendingID {
			pendingPos = i
		}
		if _, ok := erroredIDs[target.ID]; ok {
			erroredPos[target.ID] = i
		}
	}
	if pendingPos < 0 {
		t.Fatalf("newer pending CV %d missing from the batch: %#v", pendingID, targets)
	}
	if len(erroredPos) == 0 {
		t.Fatalf("seeded errored CVs missing from the batch: %#v", targets)
	}
	for id, pos := range erroredPos {
		if pos < pendingPos {
			t.Fatalf("errored CV %d (position %d) was queued before pending CV %d (position %d): %#v",
				id, pos, pendingID, pendingPos, targets)
		}
	}
}
