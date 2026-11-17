package application

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

const (
	ghOctocat = int64(770001)
	ghHubot   = int64(770002)
)

// fakeGitHubOAuth stands in for GitHub's web application flow, including
// PKCE: a code only exchanges with the verifier whose challenge it was
// issued for, and each code works once.
type fakeGitHubOAuth struct {
	challenges map[string]string
	grants     map[string]fakeGitHubGrant
	tokens     map[string]domain.GitHubUser
	issued     int
}

type fakeGitHubGrant struct {
	challenge string
	user      domain.GitHubUser
}

func newFakeGitHubOAuth() *fakeGitHubOAuth {
	return &fakeGitHubOAuth{
		challenges: map[string]string{},
		grants:     map[string]fakeGitHubGrant{},
		tokens:     map[string]domain.GitHubUser{},
	}
}

func (f *fakeGitHubOAuth) AuthorizeURL(state, challenge string) string {
	f.challenges[state] = challenge
	return "https://github.test/login/oauth/authorize?state=" + url.QueryEscape(state)
}

// authorize is the member approving the app on GitHub for this state.
func (f *fakeGitHubOAuth) authorize(state string, user domain.GitHubUser) string {
	f.issued++
	code := fmt.Sprintf("code-%d", f.issued)
	f.grants[code] = fakeGitHubGrant{challenge: f.challenges[state], user: user}
	return code
}

func (f *fakeGitHubOAuth) ExchangeCode(_ context.Context, code, verifier string) (string, error) {
	grant, ok := f.grants[code]
	delete(f.grants, code)
	sum := sha256.Sum256([]byte(verifier))
	if !ok || base64.RawURLEncoding.EncodeToString(sum[:]) != grant.challenge {
		return "", domain.ErrGitHubAuthorizationFailed
	}
	token := "ghu_" + code
	f.tokens[token] = grant.user
	return token, nil
}

func (f *fakeGitHubOAuth) GetUser(_ context.Context, token string) (domain.GitHubUser, error) {
	user, ok := f.tokens[token]
	if !ok {
		return domain.GitHubUser{}, domain.ErrGitHubAuthorizationFailed
	}
	return user, nil
}

func octocat() domain.GitHubUser {
	return domain.GitHubUser{ID: ghOctocat, Login: "octocat", AvatarURL: "https://avatars.test/u/770001"}
}

func hubot() domain.GitHubUser {
	return domain.GitHubUser{ID: ghHubot, Login: "hubot"}
}

// ghLinkSvc is ghSvc with linking on and both test businesses connected:
// v2Business to installation 9001, ghOtherBusiness to 9003.
func ghLinkSvc(t *testing.T) (*Service, context.Context, *fakeGitHubApp, *fakeGitHubOAuth, *ghClock) {
	t.Helper()
	svc, ctx, app, clock := ghSvc(t)
	db := testutil.OpenDB(t)
	cleanGitHubPartB(t, db)
	t.Cleanup(func() { cleanGitHubPartB(t, db) })
	oauth := newFakeGitHubOAuth()
	svc.SetGitHubOAuth(oauth)
	app.installations[9003] = domain.GitHubInstallationInfo{ID: 9003, AccountLogin: "other", AccountType: "Organization", CreatedAt: clock.t.Add(30 * time.Second)}
	connectGitHub(ctx, t, svc)
	if _, err := svc.CompleteGitHubInstall(ctx, ghOtherBusiness, v2UserB, 9003, ghState(t, svc, ghOtherBusiness, v2UserB)); err != nil {
		t.Fatalf("connect other business: %v", err)
	}
	return svc, ctx, app, oauth, clock
}

func cleanGitHubPartB(t *testing.T, db *sql.DB) {
	t.Helper()
	stmts := map[string][]any{
		"DELETE FROM flow_github_user_links WHERE business_id IN (?, ?) OR github_user_id IN (770001, 770002, 770003)": {v2Business, ghOtherBusiness},
		"DELETE FROM flow_github_installations WHERE business_id IN (?, ?) OR installation_id = 9003":                  {v2Business, ghOtherBusiness},
		"DELETE FROM flow_github_webhook_deliveries WHERE delivery_id LIKE 'test-%'":                                   nil,
	}
	for stmt, args := range stmts {
		if _, err := db.Exec(stmt, args...); err != nil {
			t.Fatalf("github part B cleanup: %v", err)
		}
	}
}

func ghLinkState(t *testing.T, svc *Service, businessID, userID int64) string {
	t.Helper()
	authorizeURL, err := svc.StartGitHubLink(context.Background(), businessID, userID)
	if err != nil {
		t.Fatalf("StartGitHubLink: %v", err)
	}
	u, err := url.Parse(authorizeURL)
	if err != nil {
		t.Fatal(err)
	}
	return u.Query().Get("state")
}

func linkGitHub(ctx context.Context, t *testing.T, svc *Service, oauth *fakeGitHubOAuth, businessID, userID int64, user domain.GitHubUser) (domain.GitHubUserLink, error) {
	t.Helper()
	state := ghLinkState(t, svc, businessID, userID)
	return svc.CompleteGitHubLink(ctx, businessID, userID, oauth.authorize(state, user), state)
}

func TestGitHubLinkStoresTheAccountAndCanBeReplacedAndRemoved(t *testing.T) {
	svc, ctx, _, oauth, _ := ghLinkSvc(t)

	link, err := linkGitHub(ctx, t, svc, oauth, v2Business, v2UserA, octocat())
	if err != nil {
		t.Fatalf("link: %v", err)
	}
	if link.GitHubUserID != ghOctocat || link.Login != "octocat" || link.AvatarURL == "" || link.UserID != v2UserA {
		t.Fatalf("link = %+v", link)
	}
	if got, err := svc.GitHubUserLink(ctx, v2Business, v2UserA); err != nil || got.Login != "octocat" {
		t.Fatalf("GitHubUserLink = %+v, %v", got, err)
	}

	// Linking again replaces the account rather than adding a second one.
	if link, err = linkGitHub(ctx, t, svc, oauth, v2Business, v2UserA, hubot()); err != nil || link.Login != "hubot" {
		t.Fatalf("relink = %+v, %v", link, err)
	}

	if err := svc.UnlinkGitHub(ctx, v2Business, v2UserA); err != nil {
		t.Fatalf("UnlinkGitHub: %v", err)
	}
	if _, err := svc.GitHubUserLink(ctx, v2Business, v2UserA); !errors.Is(err, domain.ErrGitHubNotLinked) {
		t.Fatalf("after unlink err = %v, want ErrGitHubNotLinked", err)
	}
	if err := svc.UnlinkGitHub(ctx, v2Business, v2UserA); err != nil {
		t.Fatalf("unlinking twice: %v", err)
	}
}

func TestCompleteGitHubLinkRejectsBadStates(t *testing.T) {
	svc, ctx, _, oauth, clock := ghLinkSvc(t)
	valid := ghLinkState(t, svc, v2Business, v2UserA)
	code := oauth.authorize(valid, octocat())

	installState := ghState(t, svc, v2Business, v2UserA)
	cases := map[string]struct {
		state      string
		businessID int64
		userID     int64
		advance    time.Duration
	}{
		"tampered business":   {strings.Replace(valid, ".5201.", ".5202.", 1), ghOtherBusiness, v2UserA, 0},
		"other business":      {valid, ghOtherBusiness, v2UserA, 0},
		"other member":        {valid, v2Business, v2UserB, 0},
		"truncated signature": {valid[:len(valid)-2], v2Business, v2UserA, 0},
		"install state":       {installState, v2Business, v2UserA, 0},
		"expired":             {valid, v2Business, v2UserA, githubStateTTL + time.Second},
	}
	start := clock.t
	for name, tc := range cases {
		clock.t = start.Add(tc.advance)
		if _, err := svc.CompleteGitHubLink(ctx, tc.businessID, tc.userID, code, tc.state); !errors.Is(err, domain.ErrGitHubInvalidState) {
			t.Errorf("%s: err = %v, want ErrGitHubInvalidState", name, err)
		}
	}
	clock.t = start

	// A link state is not an install state either.
	if _, err := svc.CompleteGitHubInstall(ctx, v2Business, v2UserA, 9001, valid); !errors.Is(err, domain.ErrGitHubInvalidState) {
		t.Errorf("link state completing an install: err = %v, want ErrGitHubInvalidState", err)
	}

	// A code GitHub issued for another link attempt fails PKCE.
	other := ghLinkState(t, svc, v2Business, v2UserA)
	if _, err := svc.CompleteGitHubLink(ctx, v2Business, v2UserA, oauth.authorize(other, hubot()), valid); !errors.Is(err, domain.ErrGitHubAuthorizationFailed) {
		t.Errorf("code from another state: err = %v, want ErrGitHubAuthorizationFailed", err)
	}
	if _, err := svc.GitHubUserLink(ctx, v2Business, v2UserA); !errors.Is(err, domain.ErrGitHubNotLinked) {
		t.Fatalf("a rejected link was stored: %v", err)
	}

	if link, err := svc.CompleteGitHubLink(ctx, v2Business, v2UserA, code, valid); err != nil || link.Login != "octocat" {
		t.Fatalf("valid state = %+v, %v", link, err)
	}
}

func TestGitHubAccountLinksToOneMemberPerBusiness(t *testing.T) {
	svc, ctx, _, oauth, _ := ghLinkSvc(t)
	if _, err := linkGitHub(ctx, t, svc, oauth, v2Business, v2UserA, octocat()); err != nil {
		t.Fatalf("link A: %v", err)
	}
	if _, err := linkGitHub(ctx, t, svc, oauth, v2Business, v2UserB, octocat()); !errors.Is(err, domain.ErrGitHubAccountLinkedElsewhere) {
		t.Fatalf("second member, same account: err = %v, want ErrGitHubAccountLinkedElsewhere", err)
	}
	if _, err := svc.GitHubUserLink(ctx, v2Business, v2UserB); !errors.Is(err, domain.ErrGitHubNotLinked) {
		t.Fatalf("refused link left a row: %v", err)
	}

	// Another business may link the same GitHub account to its own member,
	// and neither business sees the other's link.
	if _, err := linkGitHub(ctx, t, svc, oauth, ghOtherBusiness, v2UserB, octocat()); err != nil {
		t.Fatalf("other business link: %v", err)
	}
	if _, err := svc.GitHubUserLink(ctx, ghOtherBusiness, v2UserA); !errors.Is(err, domain.ErrGitHubNotLinked) {
		t.Fatalf("other business sees A's link: %v", err)
	}
	if err := svc.UnlinkGitHub(ctx, ghOtherBusiness, v2UserA); err != nil {
		t.Fatalf("UnlinkGitHub: %v", err)
	}
	if link, err := svc.GitHubUserLink(ctx, v2Business, v2UserA); err != nil || link.GitHubUserID != ghOctocat {
		t.Fatalf("unlink in another business touched this one: %+v, %v", link, err)
	}
}

func TestGitHubLinkNeedsConfigurationAndAConnectedBusiness(t *testing.T) {
	svc, ctx, _, _ := ghSvc(t)
	if _, err := svc.StartGitHubLink(ctx, v2Business, v2UserA); !errors.Is(err, domain.ErrGitHubLinkNotConfigured) {
		t.Fatalf("without OAuth credentials err = %v, want ErrGitHubLinkNotConfigured", err)
	}
	svc.SetGitHubOAuth(newFakeGitHubOAuth())
	if _, err := svc.StartGitHubLink(ctx, v2Business, v2UserA); !errors.Is(err, domain.ErrGitHubNotConnected) {
		t.Fatalf("unconnected business err = %v, want ErrGitHubNotConnected", err)
	}
	connectGitHub(ctx, t, svc)
	conn, err := svc.GitHubConnection(ctx, v2Business)
	if err != nil || !conn.LinkingEnabled {
		t.Fatalf("connection = %+v, %v; want linking enabled", conn, err)
	}
}
