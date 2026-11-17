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
)

// GitHubInstallation is the installation a business stored.
type GitHubInstallation struct {
	ID             int64
	BusinessID     int64
	InstallationID int64
	AccountLogin   string
	AccountType    string
	InstalledBy    int64
	CreatedAt      time.Time
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
