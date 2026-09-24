package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	flowworkspacev1 "github.com/voxire/voxie/gen/go/flow/workspace/v1"
)

func TestWorkspaceRosterWithoutProjectIsRetired(t *testing.T) {
	ctx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{UserID: 1, BusinessID: 2})
	response, err := NewWorkspaceAPI(nil).ListWorkspaceMembers(ctx, &flowworkspacev1.ListWorkspaceMembersRequest{})
	var status interface{ HTTPStatusCode() int }
	if response != nil || !errors.As(err, &status) || status.HTTPStatusCode() != http.StatusGone {
		t.Fatalf("response=%v error=%v, want retired endpoint", response, err)
	}
}
