package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

// githubRepositoryListCap is the most repositories the adapter lists for one
// installation. A listing that reaches it may be partial.
const githubRepositoryListCap = 1000

// githubWebhookPayload is the part of a GitHub App delivery the handlers
// read. Anything else in the body is left unparsed.
type githubWebhookPayload struct {
	Action       string `json:"action"`
	Installation *struct {
		ID          int64      `json:"id"`
		SuspendedAt *time.Time `json:"suspended_at"`
	} `json:"installation"`
	Sender struct {
		ID int64 `json:"id"`
	} `json:"sender"`
	RepositorySelection string `json:"repository_selection"`
	RepositoriesRemoved []struct {
		ID int64 `json:"id"`
	} `json:"repositories_removed"`
}

// githubWebhookHandler fills in what one event changes for a stored
// installation. Pull request and branch events (EPIC 2) register here too.
type githubWebhookHandler func(ctx context.Context, s *Service, p githubWebhookPayload, inst domain.GitHubInstallation, effect *domain.GitHubWebhookEffect) error

var githubWebhookHandlers = map[string]githubWebhookHandler{
	"installation":              handleGitHubInstallationEvent,
	"installation_repositories": handleGitHubInstallationRepositoriesEvent,
}

// HandleGitHubWebhook applies one delivery whose signature the caller has
// verified. A delivery for an installation no business holds is ignored.
// Events without a handler only record "last webhook received".
func (s *Service) HandleGitHubWebhook(ctx context.Context, deliveryID, event string, body []byte) (domain.GitHubWebhookOutcome, error) {
	var p githubWebhookPayload
	if err := json.Unmarshal(body, &p); err != nil {
		return 0, fmt.Errorf("%w: %w", domain.ErrGitHubWebhookPayload, err)
	}
	effect := domain.GitHubWebhookEffect{
		DeliveryID: deliveryID,
		Event:      event,
		ReceivedAt: s.githubNow().UTC().Truncate(time.Second),
	}
	// A revoked authorization belongs to the GitHub user, not to an
	// installation, so it is applied to every business that linked them.
	if event == "github_app_authorization" {
		if p.Action == "revoked" && p.Sender.ID > 0 {
			effect.RevokeGitHubUserID = p.Sender.ID
		}
		return s.applyGitHubWebhook(ctx, effect)
	}
	if p.Installation == nil || p.Installation.ID <= 0 {
		return domain.GitHubWebhookIgnored, nil
	}
	inst, err := s.repo.GetGitHubInstallationByInstallationID(ctx, p.Installation.ID)
	if errors.Is(err, domain.ErrGitHubNotConnected) {
		return domain.GitHubWebhookIgnored, nil
	}
	if err != nil {
		return 0, err
	}
	effect.BusinessID = inst.BusinessID
	effect.InstallationID = inst.InstallationID
	if handle, ok := githubWebhookHandlers[event]; ok {
		if err := handle(ctx, s, p, inst, &effect); err != nil {
			return 0, err
		}
	}
	return s.applyGitHubWebhook(ctx, effect)
}

func (s *Service) applyGitHubWebhook(ctx context.Context, effect domain.GitHubWebhookEffect) (domain.GitHubWebhookOutcome, error) {
	duplicate, err := s.repo.ApplyGitHubWebhook(ctx, effect)
	if err != nil {
		return 0, err
	}
	if duplicate {
		return domain.GitHubWebhookDuplicate, nil
	}
	return domain.GitHubWebhookProcessed, nil
}

func (s *Service) githubNow() time.Time {
	if s.github.now != nil {
		return s.github.now()
	}
	return time.Now()
}

// handleGitHubInstallationEvent: an uninstall removes the installation and
// every project selection from it, as disconnecting does; suspend and
// unsuspend are stored.
func handleGitHubInstallationEvent(_ context.Context, _ *Service, p githubWebhookPayload, _ domain.GitHubInstallation, effect *domain.GitHubWebhookEffect) error {
	switch p.Action {
	case "deleted":
		effect.DeleteInstallation = true
	case "suspend":
		at := effect.ReceivedAt
		if p.Installation.SuspendedAt != nil {
			at = p.Installation.SuspendedAt.UTC()
		}
		effect.SetSuspended, effect.SuspendedAt = true, &at
	case "unsuspend":
		effect.SetSuspended, effect.SuspendedAt = true, nil
	}
	return nil
}

// handleGitHubInstallationRepositoriesEvent drops repositories the
// installation lost from every project selection.
func handleGitHubInstallationRepositoriesEvent(ctx context.Context, s *Service, p githubWebhookPayload, inst domain.GitHubInstallation, effect *domain.GitHubWebhookEffect) error {
	for _, repo := range p.RepositoriesRemoved {
		if repo.ID > 0 {
			effect.RemoveRepositoryIDs = append(effect.RemoveRepositoryIDs, repo.ID)
		}
	}
	// GitHub sends an empty repositories_removed when the installation moves
	// from all repositories to a selection, so compare the picked
	// repositories with what the installation can still reach.
	if len(effect.RemoveRepositoryIDs) > 0 || p.RepositorySelection != "selected" {
		return nil
	}
	lost, err := s.githubPickedRepositoriesOutsideInstallation(ctx, inst)
	effect.RemoveRepositoryIDs = lost
	return err
}

func (s *Service) githubPickedRepositoriesOutsideInstallation(ctx context.Context, inst domain.GitHubInstallation) ([]int64, error) {
	picked, err := s.repo.ListInstallationProjectRepositoryIDs(ctx, inst.BusinessID, inst.InstallationID)
	if err != nil || len(picked) == 0 || !s.githubConfigured() {
		return nil, err
	}
	available, err := s.github.app.ListInstallationRepositories(ctx, inst.InstallationID)
	if err != nil {
		return nil, err
	}
	if len(available) >= githubRepositoryListCap {
		return nil, nil
	}
	reachable := make(map[int64]bool, len(available))
	for _, repo := range available {
		reachable[repo.ID] = true
	}
	var lost []int64
	for _, id := range picked {
		if !reachable[id] {
			lost = append(lost, id)
		}
	}
	return lost, nil
}
