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

// configureGitHubApp wires the GitHub App from the environment. All three
// variables unset leaves GitHub off; a partial or unparseable set is logged
// and also leaves it off rather than failing startup.
func configureGitHubApp(ctx context.Context, service *application.Service) {
	appID := strings.TrimSpace(os.Getenv("GITHUB_APP_ID"))
	slug := strings.TrimSpace(os.Getenv("GITHUB_APP_SLUG"))
	pem := githubPrivateKeyPEM(os.Getenv("GITHUB_APP_PRIVATE_KEY"))
	if appID == "" && slug == "" && len(pem) == 0 {
		return
	}
	client, err := githubapp.NewClient(githubapp.Config{AppID: appID, Slug: slug, PrivateKeyPEM: pem})
	if err != nil {
		logger.Ctx(ctx).Error("GitHub App is not configured correctly", logger.Error(err))
		return
	}
	// The install state is signed with a key derived from the app key, so no
	// separate secret is needed and rotating the app key voids open states.
	stateKey := sha256.Sum256(append([]byte("rtylr-flow-github-install-state\x00"), pem...))
	service.SetGitHubApp(client, stateKey[:], time.Now)
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
