package api_test

import (
	"testing"

	flowlabelv2 "github.com/voxire/voxie/gen/go/flow/label/v2"

	"github.com/voxire/onion/internal/modules/workspaces/internal/api"
	"github.com/voxire/onion/internal/testutil"
)

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
