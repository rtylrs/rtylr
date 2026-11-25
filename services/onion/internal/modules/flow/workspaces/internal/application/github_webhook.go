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

	Repository *struct {
		ID       int64  `json:"id"`
		FullName string `json:"full_name"`
	} `json:"repository"`
	PullRequest *struct {
		Number  int64  `json:"number"`
		Title   string `json:"title"`
		State   string `json:"state"`
		Draft   bool   `json:"draft"`
		Merged  bool   `json:"merged"`
		HTMLURL string `json:"html_url"`
		Head    struct {
			SHA string `json:"sha"`
			Ref string `json:"ref"`
		} `json:"head"`
	} `json:"pull_request"`
	CheckRun *struct {
		ID         int64  `json:"id"`
		Name       string `json:"name"`
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
		HeadSHA    string `json:"head_sha"`
		DetailsURL string `json:"details_url"`
		HTMLURL    string `json:"html_url"`
	} `json:"check_run"`
	// Commit status fields (the "status" event) sit at the top level.
	SHA       string `json:"sha"`
	Context   string `json:"context"`
	State     string `json:"state"`
	TargetURL string `json:"target_url"`
}

// githubWebhookHandler fills in what one event changes for a stored
// installation. Pull request and branch events (EPIC 2) register here too.
type githubWebhookHandler func(ctx context.Context, s *Service, p githubWebhookPayload, inst domain.GitHubInstallation, effect *domain.GitHubWebhookEffect) error

var githubWebhookHandlers = map[string]githubWebhookHandler{
	"installation":              handleGitHubInstallationEvent,
	"installation_repositories": handleGitHubInstallationRepositoriesEvent,
	"pull_request":              handleGitHubPullRequestEvent,
	"check_run":                 handleGitHubCheckRunEvent,
	"status":                    handleGitHubStatusEvent,
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
	s.publishLinkedTaskUpdates(ctx, effect)
	return domain.GitHubWebhookProcessed, nil
}

// publishLinkedTaskUpdates tells open screens of the tasks linked to the
// changed pull requests to refetch. Actor 0 is GitHub, so nobody's own
// screen skips the event. A failed lookup only delays the refresh.
func (s *Service) publishLinkedTaskUpdates(ctx context.Context, effect domain.GitHubWebhookEffect) {
	if len(effect.PullRequests) == 0 && len(effect.Checks) == 0 {
		return
	}
	tasks, err := s.repo.TasksLinkedToGitHubChange(ctx, effect)
	if err != nil {
		return
	}
	for _, t := range tasks {
		publishTaskUpdated(t.ID, t.ProjectID, 0)
	}
}

// githubPRState folds GitHub's state, draft and merged flags into one value.
func githubPRState(state string, draft, merged bool) string {
	switch {
	case merged:
		return "merged"
	case state == "closed":
		return "closed"
	case draft:
		return "draft"
	}
	return "open"
}

func handleGitHubPullRequestEvent(_ context.Context, _ *Service, p githubWebhookPayload, _ domain.GitHubInstallation, effect *domain.GitHubWebhookEffect) error {
	if p.PullRequest == nil || p.Repository == nil || p.Repository.ID <= 0 || p.PullRequest.Number <= 0 {
		return nil
	}
	pr := p.PullRequest
	effect.PullRequests = append(effect.PullRequests, domain.GitHubPullRequestUpdate{
		RepositoryID: p.Repository.ID,
		RepoFullName: p.Repository.FullName,
		Number:       pr.Number,
		Title:        truncateRunes(pr.Title, 500),
		State:        githubPRState(pr.State, pr.Draft, pr.Merged),
		HeadSHA:      pr.Head.SHA,
		HeadRef:      truncateRunes(pr.Head.Ref, 255),
		HTMLURL:      pr.HTMLURL,
	})
	return nil
}

func handleGitHubCheckRunEvent(_ context.Context, _ *Service, p githubWebhookPayload, _ domain.GitHubInstallation, effect *domain.GitHubWebhookEffect) error {
	run := p.CheckRun
	if run == nil || p.Repository == nil || p.Repository.ID <= 0 || run.ID <= 0 || run.HeadSHA == "" {
		return nil
	}
	details := run.DetailsURL
	if details == "" {
		details = run.HTMLURL
	}
	effect.Checks = append(effect.Checks, domain.GitHubCheckUpdate{
		RepositoryID: p.Repository.ID,
		HeadSHA:      run.HeadSHA,
		ExternalID:   fmt.Sprintf("run:%d", run.ID),
		Name:         truncateRunes(run.Name, 255),
		Status:       run.Status,
		Conclusion:   run.Conclusion,
		DetailsURL:   details,
	})
	return nil
}

// handleGitHubStatusEvent stores a commit status (posted by CI systems that do
// not use check runs) as a check keyed by its context.
func handleGitHubStatusEvent(_ context.Context, _ *Service, p githubWebhookPayload, _ domain.GitHubInstallation, effect *domain.GitHubWebhookEffect) error {
	if p.Repository == nil || p.Repository.ID <= 0 || p.SHA == "" || p.Context == "" {
		return nil
	}
	status, conclusion := "completed", ""
	switch p.State {
	case "pending":
		status = "in_progress"
	case "success":
		conclusion = "success"
	case "failure", "error":
		conclusion = "failure"
	default:
		return nil
	}
	effect.Checks = append(effect.Checks, domain.GitHubCheckUpdate{
		RepositoryID: p.Repository.ID,
		HeadSHA:      p.SHA,
		ExternalID:   truncateRunes("status:"+p.Context, 191),
		Name:         truncateRunes(p.Context, 255),
		Status:       status,
		Conclusion:   conclusion,
		DetailsURL:   p.TargetURL,
	})
	return nil
}

func truncateRunes(value string, max int) string {
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
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
