package application

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

const ghOtherBusiness = int64(5202)

type fakeGitHubApp struct {
	installations map[int64]domain.GitHubInstallationInfo
	repos         map[int64][]domain.GitHubRepository
	deleted       []int64
}

func (f *fakeGitHubApp) InstallURL(state string) string {
	return "https://github.test/install?state=" + state
}

func (f *fakeGitHubApp) GetInstallation(_ context.Context, id int64) (domain.GitHubInstallationInfo, error) {
	info, ok := f.installations[id]
	if !ok {
		return domain.GitHubInstallationInfo{}, domain.ErrGitHubInstallationNotFound
	}
	return info, nil
}

func (f *fakeGitHubApp) ListInstallationRepositories(_ context.Context, id int64) ([]domain.GitHubRepository, error) {
	return f.repos[id], nil
}

func (f *fakeGitHubApp) DeleteInstallation(_ context.Context, id int64) error {
	f.deleted = append(f.deleted, id)
	return nil
}

type ghClock struct{ t time.Time }

func (c *ghClock) now() time.Time { return c.t }

func ghSvc(t *testing.T) (*Service, context.Context, *fakeGitHubApp, *ghClock) {
	t.Helper()
	svc, ctx, db := v2Svc(t)
	cleanGitHub(t, db)
	t.Cleanup(func() { cleanGitHub(t, db) })
	clock := &ghClock{t: time.Now().UTC().Truncate(time.Second)}
	app := &fakeGitHubApp{
		installations: map[int64]domain.GitHubInstallationInfo{
			9001: {ID: 9001, AccountLogin: "acme", AccountType: "Organization", CreatedAt: clock.t.Add(30 * time.Second)},
			9002: {ID: 9002, AccountLogin: "old", AccountType: "User", CreatedAt: clock.t.Add(-24 * time.Hour)},
		},
		repos: map[int64][]domain.GitHubRepository{
			9001: {{ID: 1, FullName: "acme/api"}, {ID: 2, FullName: "acme/web", Private: true}},
		},
	}
	svc.SetGitHubApp(app, []byte("test-state-key"), clock.now)
	return svc, ctx, app, clock
}

func cleanGitHub(t *testing.T, db *sql.DB) {
	t.Helper()
	for _, stmt := range []string{
		"DELETE FROM flow_project_repositories WHERE business_id IN (?, ?)",
		"DELETE FROM flow_github_installations WHERE business_id IN (?, ?) OR installation_id IN (9001, 9002)",
	} {
		if _, err := db.Exec(stmt, v2Business, ghOtherBusiness); err != nil {
			t.Fatalf("github cleanup: %v", err)
		}
	}
}

func ghState(t *testing.T, svc *Service, businessID, userID int64) string {
	t.Helper()
	installURL, err := svc.StartGitHubInstall(businessID, userID)
	if err != nil {
		t.Fatalf("StartGitHubInstall: %v", err)
	}
	return strings.TrimPrefix(installURL, "https://github.test/install?state=")
}

func connectGitHub(ctx context.Context, t *testing.T, svc *Service) {
	t.Helper()
	if _, err := svc.CompleteGitHubInstall(ctx, v2Business, v2UserA, 9001, ghState(t, svc, v2Business, v2UserA)); err != nil {
		t.Fatalf("CompleteGitHubInstall: %v", err)
	}
}

func TestCompleteGitHubInstallRejectsBadStates(t *testing.T) {
	svc, ctx, _, clock := ghSvc(t)
	valid := ghState(t, svc, v2Business, v2UserA)

	tampered := strings.Replace(valid, "5201.", "5202.", 1)
	cases := map[string]struct {
		state      string
		businessID int64
		userID     int64
		advance    time.Duration
	}{
		"tampered business":     {tampered, ghOtherBusiness, v2UserA, 0},
		"other business":        {valid, ghOtherBusiness, v2UserA, 0},
		"other user":            {valid, v2Business, v2UserB, 0},
		"truncated signature":   {valid[:len(valid)-2], v2Business, v2UserA, 0},
		"no signature":          {"5201.880001.1.ab", v2Business, v2UserA, 0},
		"expired":               {valid, v2Business, v2UserA, githubStateTTL + time.Second},
		"signed by another key": {"", v2Business, v2UserA, 0},
	}
	other := &Service{}
	other.SetGitHubApp(&fakeGitHubApp{}, []byte("another-key"), clock.now)
	foreign, _ := other.signGitHubState(v2Business, v2UserA)
	c := cases["signed by another key"]
	c.state = foreign
	cases["signed by another key"] = c

	start := clock.t
	for name, tc := range cases {
		clock.t = start.Add(tc.advance)
		_, err := svc.CompleteGitHubInstall(ctx, tc.businessID, tc.userID, 9001, tc.state)
		if !errors.Is(err, domain.ErrGitHubInvalidState) {
			t.Errorf("%s: err = %v, want ErrGitHubInvalidState", name, err)
		}
	}
	clock.t = start
	conn, err := svc.GitHubConnection(ctx, v2Business)
	if err != nil || conn.Connected {
		t.Fatalf("a rejected state stored a connection: %+v, %v", conn, err)
	}

	conn, err = svc.CompleteGitHubInstall(ctx, v2Business, v2UserA, 9001, valid)
	if err != nil {
		t.Fatalf("valid state: %v", err)
	}
	if !conn.Connected || conn.AccountLogin != "acme" || conn.Status != domain.GitHubStatusActive {
		t.Fatalf("connection = %+v", conn)
	}
}

func TestCompleteGitHubInstallRejectsUnknownAndOlderInstallations(t *testing.T) {
	svc, ctx, app, clock := ghSvc(t)
	if _, err := svc.CompleteGitHubInstall(ctx, v2Business, v2UserA, 9999, ghState(t, svc, v2Business, v2UserA)); !errors.Is(err, domain.ErrGitHubInstallationNotFound) {
		t.Errorf("unknown installation err = %v", err)
	}
	if _, err := svc.CompleteGitHubInstall(ctx, v2Business, v2UserA, 9002, ghState(t, svc, v2Business, v2UserA)); !errors.Is(err, domain.ErrGitHubInstallationStale) {
		t.Errorf("pre-existing installation err = %v, want ErrGitHubInstallationStale", err)
	}

	// The business's own installation stays completable after it ages, e.g.
	// when GitHub redirects back after the account is renamed.
	connectGitHub(ctx, t, svc)
	clock.t = clock.t.Add(time.Hour)
	info := app.installations[9001]
	info.AccountLogin = "acme-renamed"
	app.installations[9001] = info
	conn, err := svc.CompleteGitHubInstall(ctx, v2Business, v2UserA, 9001, ghState(t, svc, v2Business, v2UserA))
	if err != nil || conn.AccountLogin != "acme-renamed" {
		t.Errorf("re-completing the business's own installation = %+v, %v", conn, err)
	}
}

func TestGitHubInstallationIsScopedToOneBusiness(t *testing.T) {
	svc, ctx, app, _ := ghSvc(t)
	connectGitHub(ctx, t, svc)

	other, err := svc.GitHubConnection(ctx, ghOtherBusiness)
	if err != nil || other.Connected {
		t.Fatalf("another business sees the connection: %+v, %v", other, err)
	}
	_, err = svc.CompleteGitHubInstall(ctx, ghOtherBusiness, v2UserB, 9001, ghState(t, svc, ghOtherBusiness, v2UserB))
	if !errors.Is(err, domain.ErrGitHubInstallationClaimed) {
		t.Fatalf("claiming another business's installation err = %v, want ErrGitHubInstallationClaimed", err)
	}

	p := createV2Project(ctx, t, svc, "gh-scope", "", v2UserA)
	if _, err := svc.SetProjectRepositories(ctx, p.ID, v2UserA, v2Business, []int64{1}); err != nil {
		t.Fatalf("SetProjectRepositories: %v", err)
	}
	// v2UserA is an admin of p, but p is not in the other business.
	if _, err := svc.ListProjectRepositories(ctx, p.ID, v2UserA, ghOtherBusiness); !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("cross-business list err = %v, want ErrUnauthorized", err)
	}
	if _, err := svc.SetProjectRepositories(ctx, p.ID, v2UserA, ghOtherBusiness, nil); !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("cross-business set err = %v, want ErrUnauthorized", err)
	}

	uninstalled, err := svc.DisconnectGitHub(ctx, v2Business)
	if err != nil || !uninstalled || len(app.deleted) != 1 || app.deleted[0] != 9001 {
		t.Fatalf("DisconnectGitHub = %v, %v, deleted %v", uninstalled, err, app.deleted)
	}
	repos, err := svc.ListProjectRepositories(ctx, p.ID, v2UserA, v2Business)
	if err != nil || len(repos) != 0 {
		t.Fatalf("selections survived disconnect: %v, %v", repos, err)
	}
}

func TestSetProjectRepositoriesNeedsProjectAdmin(t *testing.T) {
	svc, ctx, _, _ := ghSvc(t)
	connectGitHub(ctx, t, svc)
	p := createV2Project(ctx, t, svc, "gh-gate", "", v2UserA)

	if _, err := svc.SetProjectRepositories(ctx, p.ID, v2UserB, v2Business, []int64{1}); !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("non-member err = %v, want ErrUnauthorized", err)
	}
	if err := svc.AddProjectMember(ctx, domain.AddProjectMemberInput{ProjectID: p.ID, UserID: v2UserB, Role: domain.FlowRoleMember, AddedBy: v2UserA}); err != nil {
		t.Fatalf("AddProjectMember: %v", err)
	}
	if _, err := svc.SetProjectRepositories(ctx, p.ID, v2UserB, v2Business, []int64{1}); !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("member err = %v, want ErrUnauthorized", err)
	}
	if _, err := svc.ListGitHubRepositories(ctx, p.ID, v2UserB, v2Business); !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("member listing installation repos err = %v, want ErrUnauthorized", err)
	}
	if repos, err := svc.ListProjectRepositories(ctx, p.ID, v2UserB, v2Business); err != nil || len(repos) != 0 {
		t.Errorf("member reading the selection = %v, %v", repos, err)
	}

	if _, err := svc.SetProjectRepositories(ctx, p.ID, v2UserA, v2Business, []int64{1, 404}); !errors.Is(err, domain.ErrGitHubRepositoryNotAllowed) {
		t.Errorf("unknown repository err = %v, want ErrGitHubRepositoryNotAllowed", err)
	}
	got, err := svc.SetProjectRepositories(ctx, p.ID, v2UserA, v2Business, []int64{2, 1, 2})
	if err != nil {
		t.Fatalf("admin SetProjectRepositories: %v", err)
	}
	if len(got) != 2 || got[0].FullName != "acme/api" || got[1].FullName != "acme/web" || !got[1].Private {
		t.Fatalf("selection = %+v", got)
	}
	if got, err = svc.SetProjectRepositories(ctx, p.ID, v2UserA, v2Business, nil); err != nil || len(got) != 0 {
		t.Fatalf("clearing the selection = %v, %v", got, err)
	}
}

func TestGitHubConnectionReportsRemovedInstallations(t *testing.T) {
	svc, ctx, app, _ := ghSvc(t)
	connectGitHub(ctx, t, svc)
	delete(app.installations, 9001)
	conn, err := svc.GitHubConnection(ctx, v2Business)
	if err != nil || conn.Status != domain.GitHubStatusRemoved {
		t.Fatalf("connection = %+v, %v; want status removed", conn, err)
	}
}

func TestGitHubWithoutConfigurationIsOff(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	if _, err := svc.StartGitHubInstall(v2Business, v2UserA); !errors.Is(err, domain.ErrGitHubNotConfigured) {
		t.Fatalf("StartGitHubInstall err = %v, want ErrGitHubNotConfigured", err)
	}
	conn, err := svc.GitHubConnection(ctx, v2Business)
	if err != nil || conn.Configured {
		t.Fatalf("connection = %+v, %v", conn, err)
	}
}
