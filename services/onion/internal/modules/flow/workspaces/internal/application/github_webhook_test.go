package application

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

// unreachableGitHubApp answers like fakeGitHubApp except that GitHub cannot
// be reached for installation health.
type unreachableGitHubApp struct{ *fakeGitHubApp }

func (unreachableGitHubApp) GetInstallation(context.Context, int64) (domain.GitHubInstallationInfo, error) {
	return domain.GitHubInstallationInfo{}, domain.ErrGitHubUnavailable
}

func ghWebhook(ctx context.Context, t *testing.T, svc *Service, delivery, event, body string) domain.GitHubWebhookOutcome {
	t.Helper()
	outcome, err := svc.HandleGitHubWebhook(ctx, delivery, event, []byte(body))
	if err != nil {
		t.Fatalf("%s %s: %v", event, delivery, err)
	}
	return outcome
}

func ghInstallationBody(action string, installationID int64) string {
	return fmt.Sprintf(`{"action":%q,"installation":{"id":%d,"account":{"login":"acme"}},"sender":{"id":1,"login":"admin"}}`, action, installationID)
}

func TestGitHubWebhookRecordsHealthAndIgnoresRepeatDeliveries(t *testing.T) {
	svc, ctx, _, _, _ := ghLinkSvc(t)

	if got := ghWebhook(ctx, t, svc, "test-push-1", "push", `{"ref":"refs/heads/main","installation":{"id":9001}}`); got != domain.GitHubWebhookProcessed {
		t.Fatalf("first delivery outcome = %v", got)
	}
	conn, err := svc.GitHubConnection(ctx, v2Business)
	if err != nil || conn.LastWebhookAt == nil || conn.LastWebhookEvent != "push" {
		t.Fatalf("connection after a webhook = %+v, %v", conn, err)
	}
	other, err := svc.GitHubConnection(ctx, ghOtherBusiness)
	if err != nil || other.LastWebhookAt != nil {
		t.Fatalf("another business's health changed: %+v, %v", other, err)
	}

	// GitHub redelivers with the same X-GitHub-Delivery.
	if got := ghWebhook(ctx, t, svc, "test-push-1", "installation", ghInstallationBody("deleted", 9001)); got != domain.GitHubWebhookDuplicate {
		t.Fatalf("repeat delivery outcome = %v, want duplicate", got)
	}
	if conn, err := svc.GitHubConnection(ctx, v2Business); err != nil || !conn.Connected {
		t.Fatalf("a repeated delivery id was applied: %+v, %v", conn, err)
	}
}

func TestGitHubWebhookIgnoresUnknownInstallations(t *testing.T) {
	svc, ctx, _, _, _ := ghLinkSvc(t)
	if got := ghWebhook(ctx, t, svc, "test-unknown-1", "installation", ghInstallationBody("deleted", 424242)); got != domain.GitHubWebhookIgnored {
		t.Fatalf("unknown installation outcome = %v, want ignored", got)
	}
	if got := ghWebhook(ctx, t, svc, "test-ping-1", "ping", `{"zen":"Keep it logically awesome.","hook_id":1}`); got != domain.GitHubWebhookIgnored {
		t.Fatalf("delivery without installation outcome = %v, want ignored", got)
	}
	if _, err := svc.HandleGitHubWebhook(ctx, "test-bad-1", "installation", []byte(`{"action":`)); !errors.Is(err, domain.ErrGitHubWebhookPayload) {
		t.Fatalf("malformed payload err = %v, want ErrGitHubWebhookPayload", err)
	}
}

func TestGitHubWebhookInstallationDeletedRemovesTheConnectionAndSelections(t *testing.T) {
	svc, ctx, _, oauth, _ := ghLinkSvc(t)
	p := createV2Project(ctx, t, svc, "gh-webhook-delete", "", v2UserA)
	if _, err := svc.SetProjectRepositories(ctx, p.ID, v2UserA, v2Business, []int64{1, 2}); err != nil {
		t.Fatalf("SetProjectRepositories: %v", err)
	}
	if _, err := linkGitHub(ctx, t, svc, oauth, v2Business, v2UserA, octocat()); err != nil {
		t.Fatalf("link: %v", err)
	}

	ghWebhook(ctx, t, svc, "test-delete-1", "installation", ghInstallationBody("deleted", 9001))

	conn, err := svc.GitHubConnection(ctx, v2Business)
	if err != nil || conn.Connected {
		t.Fatalf("connection after uninstall = %+v, %v", conn, err)
	}
	if repos, err := svc.ListProjectRepositories(ctx, p.ID, v2UserA, v2Business); err != nil || len(repos) != 0 {
		t.Fatalf("selections survived uninstall: %v, %v", repos, err)
	}
	// The member's identity mapping is not the installation's to remove.
	if _, err := svc.GitHubUserLink(ctx, v2Business, v2UserA); err != nil {
		t.Fatalf("uninstall removed a member link: %v", err)
	}
	if other, err := svc.GitHubConnection(ctx, ghOtherBusiness); err != nil || !other.Connected {
		t.Fatalf("another business lost its connection: %+v, %v", other, err)
	}
}

func TestGitHubWebhookStoresSuspension(t *testing.T) {
	svc, ctx, app, _, clock := ghLinkSvc(t)

	ghWebhook(ctx, t, svc, "test-suspend-1", "installation",
		`{"action":"suspend","installation":{"id":9001,"suspended_at":"2026-11-17T10:00:00Z"}}`)
	inst, err := svc.repo.GetGitHubInstallation(ctx, v2Business)
	if err != nil || inst.SuspendedAt == nil || inst.SuspendedAt.UTC().Format("15:04") != "10:00" || inst.LastWebhookEvent != "installation" {
		t.Fatalf("installation after suspend = %+v, %v", inst, err)
	}
	// While GitHub cannot be asked, health falls back to the stored state.
	svc.SetGitHubApp(unreachableGitHubApp{app}, []byte("test-state-key"), clock.now)
	if conn, err := svc.GitHubConnection(ctx, v2Business); err != nil || conn.Status != domain.GitHubStatusSuspended {
		t.Fatalf("health with GitHub down = %+v, %v; want suspended", conn, err)
	}

	ghWebhook(ctx, t, svc, "test-unsuspend-1", "installation", ghInstallationBody("unsuspend", 9001))
	if inst, err := svc.repo.GetGitHubInstallation(ctx, v2Business); err != nil || inst.SuspendedAt != nil {
		t.Fatalf("installation after unsuspend = %+v, %v", inst, err)
	}
	if conn, err := svc.GitHubConnection(ctx, v2Business); err != nil || conn.Status != domain.GitHubStatusUnknown {
		t.Fatalf("health after unsuspend with GitHub down = %+v, %v; want unknown", conn, err)
	}
}

func TestGitHubWebhookDropsLostRepositoriesFromSelections(t *testing.T) {
	svc, ctx, app, _, _ := ghLinkSvc(t)
	p := createV2Project(ctx, t, svc, "gh-webhook-repos", "", v2UserA)
	if _, err := svc.SetProjectRepositories(ctx, p.ID, v2UserA, v2Business, []int64{1, 2}); err != nil {
		t.Fatalf("SetProjectRepositories: %v", err)
	}

	ghWebhook(ctx, t, svc, "test-repos-1", "installation_repositories",
		`{"action":"removed","installation":{"id":9001},"repository_selection":"selected","repositories_added":[],"repositories_removed":[{"id":2,"full_name":"acme/web"}]}`)
	repos, err := svc.ListProjectRepositories(ctx, p.ID, v2UserA, v2Business)
	if err != nil || len(repos) != 1 || repos[0].RepositoryID != 1 {
		t.Fatalf("selection after removal = %+v, %v", repos, err)
	}

	// Moving from all repositories to a selection reports nothing removed,
	// so the selection is compared with what the installation still lists.
	app.repos[9001] = []domain.GitHubRepository{{ID: 3, FullName: "acme/docs"}}
	ghWebhook(ctx, t, svc, "test-repos-2", "installation_repositories",
		`{"action":"removed","installation":{"id":9001},"repository_selection":"selected","repositories_added":[],"repositories_removed":[]}`)
	if repos, err := svc.ListProjectRepositories(ctx, p.ID, v2UserA, v2Business); err != nil || len(repos) != 0 {
		t.Fatalf("selection after narrowing to a selection = %+v, %v", repos, err)
	}
}

func TestGitHubAppAuthorizationRevokedRemovesThatUsersLinks(t *testing.T) {
	svc, ctx, _, oauth, _ := ghLinkSvc(t)
	for _, l := range []struct {
		businessID, userID int64
		user               domain.GitHubUser
	}{
		{v2Business, v2UserA, octocat()},
		{ghOtherBusiness, v2UserB, octocat()},
		{v2Business, v2UserB, hubot()},
	} {
		if _, err := linkGitHub(ctx, t, svc, oauth, l.businessID, l.userID, l.user); err != nil {
			t.Fatalf("link %d/%d: %v", l.businessID, l.userID, err)
		}
	}

	got := ghWebhook(ctx, t, svc, "test-revoke-1", "github_app_authorization",
		fmt.Sprintf(`{"action":"revoked","sender":{"id":%d,"login":"octocat"}}`, ghOctocat))
	if got != domain.GitHubWebhookProcessed {
		t.Fatalf("revoke outcome = %v", got)
	}
	for _, l := range [][2]int64{{v2Business, v2UserA}, {ghOtherBusiness, v2UserB}} {
		if _, err := svc.GitHubUserLink(ctx, l[0], l[1]); !errors.Is(err, domain.ErrGitHubNotLinked) {
			t.Errorf("link %v survived revocation: %v", l, err)
		}
	}
	if link, err := svc.GitHubUserLink(ctx, v2Business, v2UserB); err != nil || link.GitHubUserID != ghHubot {
		t.Fatalf("another GitHub user's link was removed: %+v, %v", link, err)
	}
}
