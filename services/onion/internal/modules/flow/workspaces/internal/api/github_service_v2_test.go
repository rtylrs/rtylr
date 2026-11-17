package api

import (
	"context"
	"net/http"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	flowgithubv2 "github.com/voxire/voxie/gen/go/flow/github/v2"
)

// A nil service panics if reached, so a 403 here proves the gate runs first.
func TestGitHubConnectionChangesNeedABusinessAdmin(t *testing.T) {
	api := &GitHubAPIV2{}
	ctx := authPkg.WithAuthUser(context.Background(), flowUser("flow:projects:view", "flow:projects:edit"))
	installationID := int64(1)

	calls := map[string]func() error{
		"start": func() error {
			_, err := api.StartGitHubInstall(ctx, &flowgithubv2.StartGitHubInstallRequest{})
			return err
		},
		"complete": func() error {
			_, err := api.CompleteGitHubInstall(ctx, &flowgithubv2.CompleteGitHubInstallRequest{InstallationId: &installationID, State: "s"})
			return err
		},
		"disconnect": func() error { _, err := api.DisconnectGitHub(ctx, &flowgithubv2.DisconnectGitHubRequest{}); return err },
	}
	for name, call := range calls {
		if got := flowPermissionStatus(call()); got != http.StatusForbidden {
			t.Errorf("%s: status = %d, want %d", name, got, http.StatusForbidden)
		}
	}
}

func TestRepositoryPickingNeedsFlowProjectEditPermission(t *testing.T) {
	api := &GitHubAPIV2{}
	ctx := authPkg.WithAuthUser(context.Background(), flowUser("flow:projects:view"))
	projectID := int64(1)
	if _, err := api.SetProjectRepositories(ctx, &flowgithubv2.SetProjectRepositoriesRequest{ProjectId: &projectID}); flowPermissionStatus(err) != http.StatusForbidden {
		t.Errorf("set: status = %d, want %d", flowPermissionStatus(err), http.StatusForbidden)
	}
	if _, err := api.ListGitHubRepositories(ctx, &flowgithubv2.ListGitHubRepositoriesRequest{ProjectId: &projectID}); flowPermissionStatus(err) != http.StatusForbidden {
		t.Errorf("list: status = %d, want %d", flowPermissionStatus(err), http.StatusForbidden)
	}
}

func TestGitHubAccountLinkingNeedsFlowAccess(t *testing.T) {
	api := &GitHubAPIV2{}
	ctx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{BusinessID: 9, UserID: 7})
	calls := map[string]func() error{
		"get":   func() error { _, err := api.GetMyGitHubLink(ctx, &flowgithubv2.GetMyGitHubLinkRequest{}); return err },
		"start": func() error { _, err := api.StartGitHubLink(ctx, &flowgithubv2.StartGitHubLinkRequest{}); return err },
		"complete": func() error {
			_, err := api.CompleteGitHubLink(ctx, &flowgithubv2.CompleteGitHubLinkRequest{Code: "c", State: "s"})
			return err
		},
		"unlink": func() error { _, err := api.UnlinkGitHub(ctx, &flowgithubv2.UnlinkGitHubRequest{}); return err },
	}
	for name, call := range calls {
		if got := flowPermissionStatus(call()); got != http.StatusForbidden {
			t.Errorf("%s: status = %d, want %d", name, got, http.StatusForbidden)
		}
	}
}
