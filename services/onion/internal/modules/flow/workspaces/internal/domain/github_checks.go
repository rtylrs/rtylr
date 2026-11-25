package domain

import "time"

// Pull requests and CI checks from GitHub webhooks (T-376).

// GitHubPullRequestUpdate is what a pull_request delivery says about a PR.
type GitHubPullRequestUpdate struct {
	RepositoryID int64
	RepoFullName string
	Number       int64
	Title        string
	// State is open, draft, closed or merged.
	State   string
	HeadSHA string
	HeadRef string
	HTMLURL string
}

// GitHubCheckUpdate is one check on a commit: a check run, or a commit status
// (which other CI systems post), keyed by ExternalID.
type GitHubCheckUpdate struct {
	RepositoryID int64
	HeadSHA      string
	ExternalID   string
	Name         string
	// Status is queued, in_progress or completed.
	Status     string
	Conclusion string
	DetailsURL string
}

// GitHubCheck is a stored check.
type GitHubCheck struct {
	Name       string
	Status     string
	Conclusion string
	DetailsURL string
	UpdatedAt  time.Time
}

// TaskPullRequest is a PR linked to a task, with its checks when known.
type TaskPullRequest struct {
	URL          string
	Ref          string
	RepoFullName string
	Number       int64
	Title        string
	State        string
	// Known is false for a linked PR no webhook has described yet.
	Known       bool
	ChecksState string
	Checks      []GitHubCheck
}

// Overall check states, worst first.
const (
	ChecksFailing = "failing"
	ChecksRunning = "running"
	ChecksPassing = "passing"
	ChecksNone    = "none"
)

var failingConclusions = map[string]bool{
	"failure": true, "cancelled": true, "timed_out": true, "action_required": true, "startup_failure": true,
}

// ChecksState rolls checks up: any failure is failing, else anything still
// queued or running is running, else passing; no checks is none. Neutral,
// skipped and stale count as passing.
func ChecksState(checks []GitHubCheck) string {
	if len(checks) == 0 {
		return ChecksNone
	}
	running := false
	for _, c := range checks {
		if c.Status == "completed" && failingConclusions[c.Conclusion] {
			return ChecksFailing
		}
		if c.Status != "completed" {
			running = true
		}
	}
	if running {
		return ChecksRunning
	}
	return ChecksPassing
}

// WorseChecksState returns whichever state is worse.
func WorseChecksState(a, b string) string {
	rank := map[string]int{ChecksFailing: 3, ChecksRunning: 2, ChecksPassing: 1, ChecksNone: 0, "": 0}
	if rank[b] > rank[a] {
		return b
	}
	return a
}
