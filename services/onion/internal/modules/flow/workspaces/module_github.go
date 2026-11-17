package workspaces

import (
	"context"
	"crypto/sha256"
	"os"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application"
	githubapp "github.com/voxire/onion/internal/modules/flow/workspaces/internal/infrastructure/github"
	"github.com/voxire/onion/internal/platform/logger"
)

// configureGitHubApp wires the GitHub App from the environment. All three app
// variables unset leaves GitHub off; a partial or unparseable set is logged
// and also leaves it off rather than failing startup. The OAuth client id and
// secret turn on member account linking.
func configureGitHubApp(ctx context.Context, service *application.Service) {
	appID := strings.TrimSpace(os.Getenv("GITHUB_APP_ID"))
	slug := strings.TrimSpace(os.Getenv("GITHUB_APP_SLUG"))
	pem := githubPrivateKeyPEM(os.Getenv("GITHUB_APP_PRIVATE_KEY"))
	if appID == "" && slug == "" && len(pem) == 0 {
		return
	}
	clientID := strings.TrimSpace(os.Getenv("GITHUB_APP_CLIENT_ID"))
	clientSecret := strings.TrimSpace(os.Getenv("GITHUB_APP_CLIENT_SECRET"))
	if (clientID == "") != (clientSecret == "") {
		logger.Ctx(ctx).Error("GitHub account linking needs both GITHUB_APP_CLIENT_ID and GITHUB_APP_CLIENT_SECRET; linking stays off")
		clientID, clientSecret = "", ""
	}
	client, err := githubapp.NewClient(githubapp.Config{
		AppID:         appID,
		Slug:          slug,
		PrivateKeyPEM: pem,
		ClientID:      clientID,
		ClientSecret:  clientSecret,
	})
	if err != nil {
		logger.Ctx(ctx).Error("GitHub App is not configured correctly", logger.Error(err))
		return
	}
	// Install and link states are signed with a key derived from the app key,
	// so no separate secret is needed and rotating the app key voids open
	// states.
	stateKey := sha256.Sum256(append([]byte("rtylr-flow-github-install-state\x00"), pem...))
	service.SetGitHubApp(client, stateKey[:], time.Now)
	if client.LinkingEnabled() {
		service.SetGitHubOAuth(client)
	}
}

// githubWebhookSecret is the app's webhook secret. Empty refuses every
// webhook delivery rather than accepting unsigned ones.
func githubWebhookSecret() []byte {
	return []byte(strings.TrimSpace(os.Getenv("GITHUB_APP_WEBHOOK_SECRET")))
}

// githubPrivateKeyPEM accepts the PEM with real newlines or with "\n"
// escapes, since many secret stores flatten multi-line values.
func githubPrivateKeyPEM(raw string) []byte {
	raw = strings.TrimSpace(raw)
	if raw != "" && !strings.Contains(raw, "\n") {
		raw = strings.ReplaceAll(raw, `\n`, "\n")
	}
	return []byte(raw)
}
