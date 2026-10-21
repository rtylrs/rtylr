package api_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/api"
	"github.com/voxire/onion/internal/testutil"
	flowlabelv2 "github.com/voxire/voxie/gen/go/flow/label/v2"
)

func TestLabelV2_CreateAndManageBusinessLabel(t *testing.T) {
	f := newRoleFixture(t)
	labelAPI := api.NewLabelAPIV2(f.svc())
	ctx := ctxAs(roleOwnerID)
	created, err := labelAPI.CreateLabel(ctx, &flowlabelv2.CreateLabelRequest{Name: "Portal label", Color: "#3b82f6"})
	if err != nil {
		t.Fatalf("CreateLabel without a workspace: %v", err)
	}
	id := created.Label.Id
	db := testutil.OpenDB(t)
	t.Cleanup(func() { _, _ = db.ExecContext(context.Background(), "DELETE FROM flow_labels WHERE id = ?", id) })
	var workspaceID, businessID int64
	if err := db.QueryRowContext(ctx, "SELECT workspace_id, business_id FROM flow_labels WHERE id = ?", id).Scan(&workspaceID, &businessID); err != nil {
		t.Fatal(err)
	}
	if workspaceID != 0 || businessID != testutil.BusinessID {
		t.Fatalf("label scope = workspace %d, business %d", workspaceID, businessID)
	}
	if _, err := labelAPI.UpdateLabel(ctx, &flowlabelv2.UpdateLabelRequest{Id: id, Name: "Renamed", Color: "#22c55e"}); err != nil {
		t.Fatalf("UpdateLabel: %v", err)
	}
	taskID := f.newTask(t)
	if _, err := labelAPI.SetTaskLabels(ctx, &flowlabelv2.SetTaskLabelsRequest{TaskId: testutil.Int64Ptr(taskID), LabelIds: []int64{id}}); err != nil {
		t.Fatalf("SetTaskLabels: %v", err)
	}
	listed, err := labelAPI.ListLabels(ctx, &flowlabelv2.ListLabelsRequest{Limit: 200})
	if err != nil {
		t.Fatalf("ListLabels: %v", err)
	}
	found := false
	for _, label := range listed.Labels {
		if label.Id == id && label.Name == "Renamed" && label.Color == "#22c55e" {
			found = true
		}
	}
	if !found {
		t.Fatal("persisted label missing from business list")
	}
	if _, err := labelAPI.DeleteLabel(ctx, &flowlabelv2.DeleteLabelRequest{Id: id}); err != nil {
		t.Fatalf("DeleteLabel: %v", err)
	}
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM flow_labels WHERE id = ?", id).Scan(&count); err != nil || count != 0 {
		t.Fatalf("label remains after deletion: count %d, error %v", count, err)
	}
}

// SetTaskLabels mutates a task through the label service, so it takes the same
// role tier as the task handlers. The repository guard underneath only checks
// workspace membership, which admits a viewer.
//
// The v1 handler had a RequireTaskRole check and said exactly that in a comment.
// v2 shipped without it, so a viewer could relabel any task they could see, and
// v2 is the registered path. These pin both sides of the gate.

func TestLabelV2_ViewerCannotSetTaskLabels(t *testing.T) {
	f := newRoleFixture(t)
	labelV2 := api.NewLabelAPIV2(f.svc())
	taskID := f.newTask(t)

	_, err := labelV2.SetTaskLabels(ctxAs(roleViewerID),
		&flowlabelv2.SetTaskLabelsRequest{TaskId: testutil.Int64Ptr(taskID), LabelIds: []int64{101}})

	assertPermissionDenied(t, err, "viewer SetTaskLabels via v2")
}

func TestLabelV2_MemberCanSetTaskLabels(t *testing.T) {
	f := newRoleFixture(t)
	labelV2 := api.NewLabelAPIV2(f.svc())
	taskID := f.newTask(t)

	if _, err := labelV2.SetTaskLabels(ctxAs(roleMemberID),
		&flowlabelv2.SetTaskLabelsRequest{TaskId: testutil.Int64Ptr(taskID), LabelIds: []int64{101}}); err != nil {
		t.Fatalf("member SetTaskLabels via v2: expected success, got %v", err)
	}
}
