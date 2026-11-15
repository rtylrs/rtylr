package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/testsupport"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

// Both delete statements are self-scoped by user_id, so a caller deleting a row
// they did not write matches zero rows. Returning nil there made the API answer
// "deleted successfully" while the row stayed put; requireAffected must turn
// the zero-row delete into a denial instead.
func TestDeleteCommentRefusesWhenCallerIsNotTheAuthor(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
	ctx := context.Background()

	taskID, _ := seedArchivedTestTask(t, db, repo, "DeleteCommentGuard")

	commentID, err := repo.CreateComment(ctx, taskID, testutil.OwnerID, "author's comment")
	if err != nil {
		t.Fatalf("CreateComment: %v", err)
	}

	if err := repo.DeleteComment(ctx, commentID, testutil.OtherOwnerID); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("DeleteComment as a non-author = %v, want ErrUnauthorized", err)
	}

	comments, err := repo.GetCommentsByTask(ctx, taskID)
	if err != nil {
		t.Fatalf("GetCommentsByTask: %v", err)
	}
	found := false
	for _, comment := range comments {
		if comment.ID == commentID {
			found = true
		}
	}
	if !found {
		t.Fatal("the comment was deleted by a caller who did not write it")
	}

	// The author's own delete still has to work.
	if err := repo.DeleteComment(ctx, commentID, testutil.OwnerID); err != nil {
		t.Fatalf("DeleteComment as the author: %v", err)
	}
	comments, err = repo.GetCommentsByTask(ctx, taskID)
	if err != nil {
		t.Fatalf("GetCommentsByTask after delete: %v", err)
	}
	for _, comment := range comments {
		if comment.ID == commentID {
			t.Error("the author's comment survived their own delete")
		}
	}
}

func TestDeleteTimeLogRefusesWhenCallerIsNotTheAuthor(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
	ctx := context.Background()

	taskID, businessID := seedArchivedTestTask(t, db, repo, "DeleteTimeLogGuard")

	logID, err := repo.CreateTimeLog(ctx, domain.TimeLogInput{
		OwnerID: testutil.OwnerID, BusinessID: businessID, TaskID: taskID,
		UserID: testutil.OwnerID, DurationMinutes: 30, Description: "author's log",
	})
	if err != nil {
		t.Fatalf("CreateTimeLog: %v", err)
	}

	if err := repo.DeleteTimeLog(ctx, logID, testutil.OtherOwnerID); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("DeleteTimeLog as a non-author = %v, want ErrUnauthorized", err)
	}

	logs, err := repo.GetTimeLogsByTask(ctx, taskID)
	if err != nil {
		t.Fatalf("GetTimeLogsByTask: %v", err)
	}
	found := false
	for _, log := range logs {
		if log.ID == logID {
			found = true
		}
	}
	if !found {
		t.Fatal("the time log was deleted by a caller who did not write it")
	}

	// The author's own delete still has to work.
	if err := repo.DeleteTimeLog(ctx, logID, testutil.OwnerID); err != nil {
		t.Fatalf("DeleteTimeLog as the author: %v", err)
	}
	logs, err = repo.GetTimeLogsByTask(ctx, taskID)
	if err != nil {
		t.Fatalf("GetTimeLogsByTask after delete: %v", err)
	}
	for _, log := range logs {
		if log.ID == logID {
			t.Error("the author's time log survived their own delete")
		}
	}
}

// Editing is author-only, like the delete: someone else's comment must stay as
// written and report a denial, while the author's edit lands and is marked.
func TestUpdateCommentIsAuthorOnly(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
	ctx := context.Background()

	taskID, _ := seedArchivedTestTask(t, db, repo, "UpdateCommentGuard")

	commentID, err := repo.CreateComment(ctx, taskID, testutil.OwnerID, "first wording")
	if err != nil {
		t.Fatalf("CreateComment: %v", err)
	}

	if err := repo.UpdateComment(ctx, commentID, testutil.OtherOwnerID, "hijacked"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("UpdateComment as a non-author = %v, want ErrUnauthorized", err)
	}
	stored, err := repo.GetCommentByID(ctx, commentID)
	if err != nil {
		t.Fatalf("GetCommentByID: %v", err)
	}
	if stored.Content != "first wording" || stored.IsEdited {
		t.Fatalf("a non-author changed the comment: content=%q edited=%v", stored.Content, stored.IsEdited)
	}

	if err := repo.UpdateComment(ctx, commentID, testutil.OwnerID, "second wording"); err != nil {
		t.Fatalf("UpdateComment as the author: %v", err)
	}
	stored, err = repo.GetCommentByID(ctx, commentID)
	if err != nil {
		t.Fatalf("GetCommentByID after edit: %v", err)
	}
	if stored.Content != "second wording" || !stored.IsEdited {
		t.Fatalf("author edit = content %q edited %v, want %q and edited", stored.Content, stored.IsEdited, "second wording")
	}
}
