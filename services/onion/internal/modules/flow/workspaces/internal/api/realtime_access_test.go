package api_test

import (
	"fmt"
	"testing"

	flowprojectv2 "github.com/voxire/voxie/gen/go/flow/project/v2"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/api"
	"github.com/voxire/onion/internal/testutil"
)

// Live updates kept reaching someone after they lost access to a project,
// until the hub's cached decision expired. The handlers that take access away
// now ask every hub to recheck at once.
func recordRechecks(t *testing.T) *[]string {
	t.Helper()
	var scopes []string
	original := *api.RecheckRealtimeAccess
	*api.RecheckRealtimeAccess = func(scope string) { scopes = append(scopes, scope) }
	t.Cleanup(func() { *api.RecheckRealtimeAccess = original })
	return &scopes
}

func TestRemovingAMemberRechecksTheirLiveAccess(t *testing.T) {
	f := newRoleFixture(t)
	scopes := recordRechecks(t)

	if _, err := f.projAPI.RemoveProjectMember(ctxAs(roleOwnerID), &flowprojectv2.RemoveProjectMemberRequest{
		ProjectId: testutil.Int64Ptr(f.projID), UserId: testutil.Int64Ptr(roleMemberID),
	}); err != nil {
		t.Fatalf("RemoveProjectMember: %v", err)
	}
	if want := fmt.Sprintf("user:%d", roleMemberID); len(*scopes) != 1 || (*scopes)[0] != want {
		t.Fatalf("rechecked %v, want [%s]", *scopes, want)
	}
}

func TestMakingAProjectPrivateRechecksEveryoneWatchingIt(t *testing.T) {
	f := newRoleFixture(t)
	scopes := recordRechecks(t)

	if _, err := f.projAPI.SetProjectVisibility(ctxAs(roleOwnerID), &flowprojectv2.SetProjectVisibilityRequest{
		Id: f.projID, Visibility: "public",
	}); err != nil {
		t.Fatalf("SetProjectVisibility public: %v", err)
	}
	if len(*scopes) != 0 {
		t.Fatalf("making a project public rechecked %v; it takes no access away", *scopes)
	}
	if _, err := f.projAPI.SetProjectVisibility(ctxAs(roleOwnerID), &flowprojectv2.SetProjectVisibilityRequest{
		Id: f.projID, Visibility: "private",
	}); err != nil {
		t.Fatalf("SetProjectVisibility private: %v", err)
	}
	if want := fmt.Sprintf("project:%d", f.projID); len(*scopes) != 1 || (*scopes)[0] != want {
		t.Fatalf("rechecked %v, want [%s]", *scopes, want)
	}
}
