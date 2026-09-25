package api_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/workspaces/internal/api"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/testutil"
	flowtaskv2 "github.com/voxire/voxie/gen/go/flow/task/v2"
)

// The comment and time-log deletes are scoped to the row's author, so an
// admin-tier caller deleting a teammate's row gets ErrUnauthorized from the
// repository. The handler has to answer with a permission denial: it used to
// answer 500, and before requireAffected it answered success without deleting
// anything (the reported bug).
func TestTaskAPIV2_DeleteCommentByNonAuthorIsDenied(t *testing.T) {
	f := newRoleFixture(t)
	taskAPI := api.NewTaskAPIV2(f.svc())
	ctx := context.Background()
	taskID := f.newTask(t)

	commentID, err := f.repo.CreateComment(ctx, taskID, roleOwnerID, "owner's comment")
	if err != nil {
		t.Fatalf("CreateComment: %v", err)
	}

	_, err = taskAPI.DeleteComment(ctxAs(roleManagerID), &flowtaskv2.DeleteCommentRequest{Id: commentID})
	assertPermissionDenied(t, err, "admin DeleteComment for another author")

	comments, err := f.repo.GetCommentsByTask(ctx, taskID)
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
	if _, err := taskAPI.DeleteComment(ctxAs(roleOwnerID), &flowtaskv2.DeleteCommentRequest{Id: commentID}); err != nil {
		t.Fatalf("author DeleteComment: %v", err)
	}
}

func TestTaskAPIV2_DeleteTimeLogByNonAuthorIsDenied(t *testing.T) {
	f := newRoleFixture(t)
	taskAPI := api.NewTaskAPIV2(f.svc())
	ctx := context.Background()
	taskID := f.newTask(t)

	logID, err := f.repo.CreateTimeLog(ctx, domain.TimeLogInput{
		OwnerID: roleOwnerID, BusinessID: testutil.BusinessID, TaskID: taskID,
		UserID: roleOwnerID, DurationMinutes: 30, Description: "owner's log",
	})
	if err != nil {
		t.Fatalf("CreateTimeLog: %v", err)
	}

	_, err = taskAPI.DeleteTimeLog(ctxAs(roleManagerID), &flowtaskv2.DeleteTimeLogRequest{Id: logID})
	assertPermissionDenied(t, err, "admin DeleteTimeLog for another author")

	logs, err := f.repo.GetTimeLogsByTask(ctx, taskID)
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
	if _, err := taskAPI.DeleteTimeLog(ctxAs(roleOwnerID), &flowtaskv2.DeleteTimeLogRequest{Id: logID}); err != nil {
		t.Fatalf("author DeleteTimeLog: %v", err)
	}
}
