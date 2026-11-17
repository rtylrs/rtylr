package application

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

const (
	githubStateTTL = 10 * time.Minute
	// githubClockSkew tolerates GitHub's clock running behind ours when
	// comparing an installation's created_at with the state's issue time.
	githubClockSkew = 2 * time.Minute
)

// GitHubApp is the GitHub App API as the application needs it. The adapter
// signs app JWTs and mints installation tokens itself; neither reaches here.
type GitHubApp interface {
	InstallURL(state string) string
	GetInstallation(ctx context.Context, installationID int64) (domain.GitHubInstallationInfo, error)
	ListInstallationRepositories(ctx context.Context, installationID int64) ([]domain.GitHubRepository, error)
	DeleteInstallation(ctx context.Context, installationID int64) error
}

type githubStore interface {
	GetGitHubInstallation(ctx context.Context, businessID int64) (domain.GitHubInstallation, error)
	CreateGitHubInstallation(ctx context.Context, in domain.GitHubInstallation) error
	UpdateGitHubInstallationAccount(ctx context.Context, businessID, installationID int64, login, accountType string) error
	DeleteGitHubInstallation(ctx context.Context, businessID int64) error
	ListProjectRepositories(ctx context.Context, businessID, projectID int64) ([]domain.ProjectRepository, error)
	ReplaceProjectRepositories(ctx context.Context, businessID, projectID, installationID, callerID int64, repos []domain.GitHubRepository) error
}

type githubIntegration struct {
	app      GitHubApp
	stateKey []byte
	now      func() time.Time
}

// SetGitHubApp wires the GitHub App adapter. stateKey signs install states.
// Without it every GitHub operation reports domain.ErrGitHubNotConfigured.
func (s *Service) SetGitHubApp(app GitHubApp, stateKey []byte, now func() time.Time) {
	if app == nil || len(stateKey) == 0 {
		s.github = githubIntegration{}
		return
	}
	if now == nil {
		now = time.Now
	}
	s.github = githubIntegration{app: app, stateKey: stateKey, now: now}
}

func (s *Service) githubConfigured() bool { return s.github.app != nil }

// GitHubConnection returns the business's connection with its health read
// live from GitHub.
func (s *Service) GitHubConnection(ctx context.Context, businessID int64) (domain.GitHubConnection, error) {
	conn := domain.GitHubConnection{Configured: s.githubConfigured()}
	inst, err := s.repo.GetGitHubInstallation(ctx, businessID)
	if errors.Is(err, domain.ErrGitHubNotConnected) {
		return conn, nil
	}
	if err != nil {
		return conn, err
	}
	conn.Connected = true
	conn.AccountLogin = inst.AccountLogin
	conn.AccountType = inst.AccountType
	conn.ConnectedAt = inst.CreatedAt
	conn.Status = s.githubInstallationStatus(ctx, inst.InstallationID)
	return conn, nil
}

func (s *Service) githubInstallationStatus(ctx context.Context, installationID int64) string {
	if !s.githubConfigured() {
		return domain.GitHubStatusUnknown
	}
	info, err := s.github.app.GetInstallation(ctx, installationID)
	switch {
	case errors.Is(err, domain.ErrGitHubInstallationNotFound):
		return domain.GitHubStatusRemoved
	case err != nil:
		logger.Ctx(ctx).Warn("github installation health check failed", logger.Error(err))
		return domain.GitHubStatusUnknown
	case info.SuspendedAt != nil:
		return domain.GitHubStatusSuspended
	default:
		return domain.GitHubStatusActive
	}
}

// StartGitHubInstall returns the GitHub install URL carrying a signed state.
// The caller must already be a business admin.
func (s *Service) StartGitHubInstall(businessID, userID int64) (string, error) {
	if !s.githubConfigured() {
		return "", domain.ErrGitHubNotConfigured
	}
	state, err := s.signGitHubState(businessID, userID)
	if err != nil {
		return "", err
	}
	return s.github.app.InstallURL(state), nil
}

// CompleteGitHubInstall stores the installation GitHub redirected back with.
// The caller must already be a business admin.
//
// The state proves this business and user started the install; the app JWT
// lookup proves the installation exists for this app. Neither proves the
// installation is the one this user just created, since installation ids are
// sequential, so a new claim must also be newer than the state. Claiming an
// older installation is therefore rejected: uninstall it on GitHub and connect
// again. Completing again for the installation the business already holds
// only refreshes its account details.
func (s *Service) CompleteGitHubInstall(ctx context.Context, businessID, userID, installationID int64, state string) (domain.GitHubConnection, error) {
	if !s.githubConfigured() {
		return domain.GitHubConnection{}, domain.ErrGitHubNotConfigured
	}
	issuedAt, err := s.verifyGitHubState(state, businessID, userID)
	if err != nil {
		return domain.GitHubConnection{}, err
	}
	info, err := s.github.app.GetInstallation(ctx, installationID)
	if err != nil {
		return domain.GitHubConnection{}, err
	}
	existing, err := s.repo.GetGitHubInstallation(ctx, businessID)
	switch {
	case err == nil && existing.InstallationID != installationID:
		return domain.GitHubConnection{}, domain.ErrGitHubAlreadyConnected
	case err == nil:
		// Already this business's installation, so freshness does not apply.
		err = s.repo.UpdateGitHubInstallationAccount(ctx, businessID, installationID, info.AccountLogin, info.AccountType)
	case errors.Is(err, domain.ErrGitHubNotConnected):
		if info.CreatedAt.Before(issuedAt.Add(-githubClockSkew)) {
			return domain.GitHubConnection{}, domain.ErrGitHubInstallationStale
		}
		err = s.repo.CreateGitHubInstallation(ctx, domain.GitHubInstallation{
			BusinessID:     businessID,
			InstallationID: installationID,
			AccountLogin:   info.AccountLogin,
			AccountType:    info.AccountType,
			InstalledBy:    userID,
		})
	}
	if err != nil {
		return domain.GitHubConnection{}, err
	}
	return s.GitHubConnection(ctx, businessID)
}

// DisconnectGitHub removes the stored installation and every project's
// selection, then asks GitHub to uninstall the app. The local removal stands
// even if GitHub refuses; the bool reports whether GitHub was updated too.
// The caller must already be a business admin.
func (s *Service) DisconnectGitHub(ctx context.Context, businessID int64) (bool, error) {
	inst, err := s.repo.GetGitHubInstallation(ctx, businessID)
	if err != nil {
		return false, err
	}
	if err := s.repo.DeleteGitHubInstallation(ctx, businessID); err != nil {
		return false, err
	}
	if !s.githubConfigured() {
		return false, nil
	}
	err = s.github.app.DeleteInstallation(ctx, inst.InstallationID)
	if err != nil && !errors.Is(err, domain.ErrGitHubInstallationNotFound) {
		logger.Ctx(ctx).Warn("github uninstall failed", logger.Error(err))
		return false, nil
	}
	return true, nil
}

// ListGitHubRepositories lists what the business's installation can access,
// for a project admin picking repositories.
func (s *Service) ListGitHubRepositories(ctx context.Context, projectID, callerID, businessID int64) ([]domain.GitHubRepository, error) {
	inst, err := s.projectGitHubInstallation(ctx, projectID, callerID, businessID, domain.FlowRoleAdmin)
	if err != nil {
		return nil, err
	}
	return s.github.app.ListInstallationRepositories(ctx, inst.InstallationID)
}

// ListProjectRepositories returns a project's picked repositories.
func (s *Service) ListProjectRepositories(ctx context.Context, projectID, callerID, businessID int64) ([]domain.ProjectRepository, error) {
	if err := s.requireProjectInBusiness(ctx, projectID, callerID, businessID, domain.FlowRoleViewer); err != nil {
		return nil, err
	}
	return s.repo.ListProjectRepositories(ctx, businessID, projectID)
}

// SetProjectRepositories replaces a project's selection. Every id must be one
// the installation can access right now, so a stale or forged id is refused
// rather than stored.
func (s *Service) SetProjectRepositories(ctx context.Context, projectID, callerID, businessID int64, repositoryIDs []int64) ([]domain.ProjectRepository, error) {
	inst, err := s.projectGitHubInstallation(ctx, projectID, callerID, businessID, domain.FlowRoleAdmin)
	if err != nil {
		return nil, err
	}
	wanted := make([]int64, 0, len(repositoryIDs))
	seen := make(map[int64]bool, len(repositoryIDs))
	for _, id := range repositoryIDs {
		if !seen[id] {
			seen[id] = true
			wanted = append(wanted, id)
		}
	}
	if len(wanted) > domain.MaxProjectRepositories {
		return nil, domain.ErrGitHubTooManyRepositories
	}
	picked := make([]domain.GitHubRepository, 0, len(wanted))
	if len(wanted) > 0 {
		available, err := s.github.app.ListInstallationRepositories(ctx, inst.InstallationID)
		if err != nil {
			return nil, err
		}
		byID := make(map[int64]domain.GitHubRepository, len(available))
		for _, repo := range available {
			byID[repo.ID] = repo
		}
		for _, id := range wanted {
			repo, ok := byID[id]
			if !ok {
				return nil, domain.ErrGitHubRepositoryNotAllowed
			}
			picked = append(picked, repo)
		}
	}
	if err := s.repo.ReplaceProjectRepositories(ctx, businessID, projectID, inst.InstallationID, callerID, picked); err != nil {
		return nil, err
	}
	return s.repo.ListProjectRepositories(ctx, businessID, projectID)
}

func (s *Service) projectGitHubInstallation(ctx context.Context, projectID, callerID, businessID int64, minRole string) (domain.GitHubInstallation, error) {
	if err := s.requireProjectInBusiness(ctx, projectID, callerID, businessID, minRole); err != nil {
		return domain.GitHubInstallation{}, err
	}
	if !s.githubConfigured() {
		return domain.GitHubInstallation{}, domain.ErrGitHubNotConfigured
	}
	return s.repo.GetGitHubInstallation(ctx, businessID)
}

// requireProjectInBusiness checks the role and that the project belongs to
// the business. A business admin passes RequireProjectRole for any id, so the
// ownership check is what keeps a foreign project id out of this tenant's rows.
func (s *Service) requireProjectInBusiness(ctx context.Context, projectID, callerID, businessID int64, minRole string) error {
	if err := s.RequireProjectRole(ctx, projectID, callerID, businessID, minRole); err != nil {
		return domain.ErrUnauthorized
	}
	if _, err := s.repo.GetProjectV2ByID(ctx, projectID, businessID); err != nil {
		return domain.ErrUnauthorized
	}
	return nil
}

// Install state: business.user.issuedUnix.nonce.signature, HMAC-SHA256 over the
// first four parts. GitHub echoes it back verbatim to the Setup URL.
func (s *Service) signGitHubState(businessID, userID int64) (string, error) {
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	payload := fmt.Sprintf("%d.%d.%d.%s", businessID, userID, s.github.now().Unix(), hex.EncodeToString(nonce))
	return payload + "." + s.githubStateMAC(payload), nil
}

func (s *Service) githubStateMAC(payload string) string {
	mac := hmac.New(sha256.New, s.github.stateKey)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *Service) verifyGitHubState(state string, businessID, userID int64) (time.Time, error) {
	cut := strings.LastIndex(state, ".")
	if cut <= 0 {
		return time.Time{}, domain.ErrGitHubInvalidState
	}
	payload, sig := state[:cut], state[cut+1:]
	if !hmac.Equal([]byte(sig), []byte(s.githubStateMAC(payload))) {
		return time.Time{}, domain.ErrGitHubInvalidState
	}
	parts := strings.Split(payload, ".")
	if len(parts) != 4 {
		return time.Time{}, domain.ErrGitHubInvalidState
	}
	stateBusiness, errB := strconv.ParseInt(parts[0], 10, 64)
	stateUser, errU := strconv.ParseInt(parts[1], 10, 64)
	issuedUnix, errT := strconv.ParseInt(parts[2], 10, 64)
	if errB != nil || errU != nil || errT != nil || stateBusiness != businessID || stateUser != userID {
		return time.Time{}, domain.ErrGitHubInvalidState
	}
	issued := time.Unix(issuedUnix, 0)
	now := s.github.now()
	if now.Sub(issued) > githubStateTTL || issued.After(now.Add(time.Minute)) {
		return time.Time{}, domain.ErrGitHubInvalidState
	}
	return issued, nil
}
