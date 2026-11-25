package domain

import (
	"errors"
	"time"
)

// Installation health as read from GitHub.
const (
	GitHubStatusActive    = "active"
	GitHubStatusSuspended = "suspended"
	GitHubStatusRemoved   = "removed"
	GitHubStatusUnknown   = "unknown"
)

// MaxProjectRepositories bounds one project's selection.
const MaxProjectRepositories = 100

var (
	ErrGitHubNotConfigured        = errors.New("github app is not configured")
	ErrGitHubNotConnected         = errors.New("github is not connected")
	ErrGitHubInvalidState         = errors.New("github install state is invalid or expired")
	ErrGitHubInstallationNotFound = errors.New("github installation not found")
	ErrGitHubInstallationStale    = errors.New("github installation predates this install attempt")
	ErrGitHubInstallationClaimed  = errors.New("github installation belongs to another business")
	ErrGitHubAlreadyConnected     = errors.New("business already has a different github installation")
	ErrGitHubRepositoryNotAllowed = errors.New("repository is not available to the installation")
	ErrGitHubTooManyRepositories  = errors.New("too many repositories")
	ErrGitHubUnavailable          = errors.New("github request failed")

	ErrGitHubLinkNotConfigured      = errors.New("github account linking is not configured")
	ErrGitHubNotLinked              = errors.New("github account is not linked")
	ErrGitHubAuthorizationFailed    = errors.New("github did not accept the authorization")
	ErrGitHubAccountLinkedElsewhere = errors.New("github account is linked to another member of this business")
	ErrGitHubWebhookPayload         = errors.New("github webhook payload is invalid")
)

// GitHubInstallation is the installation a business stored.
type GitHubInstallation struct {
	ID             int64
	BusinessID     int64
	InstallationID int64
	AccountLogin   string
	AccountType    string
	InstalledBy    int64
	// SuspendedAt is set from installation suspend webhooks.
	SuspendedAt      *time.Time
	LastWebhookAt    *time.Time
	LastWebhookEvent string
	CreatedAt        time.Time
}

// GitHubInstallationInfo is what GitHub reports about an installation.
type GitHubInstallationInfo struct {
	ID           int64
	AccountLogin string
	AccountType  string
	CreatedAt    time.Time
	SuspendedAt  *time.Time
}

// GitHubConnection is a business's connection as shown to a caller.
type GitHubConnection struct {
	Configured   bool
	Connected    bool
	AccountLogin string
	AccountType  string
	Status       string
	ConnectedAt  time.Time
	// LastWebhookAt is nil until a signed webhook for the installation arrives.
	LastWebhookAt    *time.Time
	LastWebhookEvent string
	LinkingEnabled   bool
}

// GitHubRepository is a repository an installation can access.
type GitHubRepository struct {
	ID       int64
	FullName string
	Private  bool
}

// ProjectRepository is a repository a project picked.
type ProjectRepository struct {
	RepositoryID int64
	FullName     string
	Private      bool
	CreatedAt    time.Time
}

// GitHubUser is the GitHub account behind a user access token.
type GitHubUser struct {
	ID        int64
	Login     string
	AvatarURL string
}

// GitHubUserLink is a member's GitHub account in one business.
type GitHubUserLink struct {
	BusinessID   int64
	UserID       int64
	GitHubUserID int64
	Login        string
	AvatarURL    string
	CreatedAt    time.Time
}

// GitHubWebhookOutcome tells the webhook handler which 2xx to answer with.
type GitHubWebhookOutcome int

const (
	// GitHubWebhookProcessed: the delivery was recorded and applied.
	GitHubWebhookProcessed GitHubWebhookOutcome = iota
	// GitHubWebhookDuplicate: this delivery id was already processed.
	GitHubWebhookDuplicate
	// GitHubWebhookIgnored: no stored installation matches the delivery.
	GitHubWebhookIgnored
)

// GitHubWebhookEffect is what one delivery changes, applied in the same
// transaction that records the delivery id.
type GitHubWebhookEffect struct {
	DeliveryID string
	Event      string
	// BusinessID and InstallationID are zero for app-level events.
	BusinessID     int64
	InstallationID int64
	ReceivedAt     time.Time

	DeleteInstallation bool
	// SetSuspended writes SuspendedAt, nil meaning unsuspended.
	SetSuspended        bool
	SuspendedAt         *time.Time
	RemoveRepositoryIDs []int64
	RevokeGitHubUserID  int64

	// PullRequests and Checks come from pull_request, check_run and status
	// deliveries (T-376).
	PullRequests []GitHubPullRequestUpdate
	Checks       []GitHubCheckUpdate
}
