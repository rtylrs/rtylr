package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	flowgithubv2 "github.com/voxire/voxie/gen/go/flow/github/v2"
)

// GitHubAPIV2 serves T-366: the business's GitHub App installation, each
// project's repository selection and each member's linked GitHub account.
// Connecting and disconnecting take a business admin, the same gate as Dash
// Extensions connections; picking repositories takes a project admin; linking
// an account takes Flow access only.
type GitHubAPIV2 struct{ Service *application.Service }

func NewGitHubAPIV2(s *application.Service) *GitHubAPIV2 { return &GitHubAPIV2{Service: s} }

func githubError(ctx context.Context, err error, fallback string) error {
	switch {
	case errors.Is(err, domain.ErrUnauthorized):
		return onekiterrors.PermissionDenied()
	case errors.Is(err, domain.ErrGitHubNotConfigured):
		return onekiterrors.New(http.StatusServiceUnavailable, "GitHub is not configured on this server")
	case errors.Is(err, domain.ErrGitHubNotConnected):
		return onekiterrors.New(http.StatusConflict, "GitHub is not connected")
	case errors.Is(err, domain.ErrGitHubInvalidState):
		return onekiterrors.New(http.StatusBadRequest, "This GitHub link has expired or is not yours. Start connecting again.")
	case errors.Is(err, domain.ErrGitHubInstallationNotFound):
		return onekiterrors.New(http.StatusBadRequest, "GitHub does not know this installation")
	case errors.Is(err, domain.ErrGitHubInstallationStale):
		return onekiterrors.New(http.StatusConflict, "The app was already installed on that GitHub account. Uninstall it on GitHub, then connect again.")
	case errors.Is(err, domain.ErrGitHubInstallationClaimed):
		return onekiterrors.New(http.StatusConflict, "That GitHub installation is connected to another business")
	case errors.Is(err, domain.ErrGitHubAlreadyConnected):
		return onekiterrors.New(http.StatusConflict, "Disconnect the current GitHub account before connecting another")
	case errors.Is(err, domain.ErrGitHubRepositoryNotAllowed):
		return onekiterrors.New(http.StatusBadRequest, "A repository is not available to the GitHub installation")
	case errors.Is(err, domain.ErrGitHubTooManyRepositories):
		return onekiterrors.New(http.StatusBadRequest, "A project can use at most 100 repositories")
	case errors.Is(err, domain.ErrGitHubLinkNotConfigured):
		return onekiterrors.New(http.StatusServiceUnavailable, "Linking GitHub accounts is not configured on this server")
	case errors.Is(err, domain.ErrGitHubAuthorizationFailed):
		return onekiterrors.New(http.StatusBadRequest, "GitHub did not accept the authorization. Link your account again.")
	case errors.Is(err, domain.ErrGitHubAccountLinkedElsewhere):
		return onekiterrors.New(http.StatusConflict, "That GitHub account is linked to another member of this business")
	case errors.Is(err, domain.ErrGitHubUnavailable):
		logger.Ctx(ctx).Warn(fallback, logger.Error(err))
		return onekiterrors.New(http.StatusBadGateway, "GitHub did not respond. Try again.")
	default:
		return workspaceInternalError(ctx, fallback, err)
	}
}

func requireBusinessAdmin(ctx context.Context) (int64, *authPkg.AuthUser, error) {
	if user, ok := authPkg.GetAuthUser(ctx); !ok || !authPkg.IsBusinessAdmin(user) {
		return 0, nil, onekiterrors.PermissionDenied()
	}
	return resolveBusinessIDV2(ctx)
}

func toGitHubConnection(c domain.GitHubConnection, canManage bool) *flowgithubv2.GitHubConnection {
	out := &flowgithubv2.GitHubConnection{
		Configured:   c.Configured,
		Connected:    c.Connected,
		AccountLogin: c.AccountLogin,
		AccountType:  c.AccountType,
		Status:       c.Status,
		CanManage:    canManage,
	}
	if c.Connected {
		out.ConnectedAt = c.ConnectedAt.UTC().Format(time.RFC3339)
		out.LinkingEnabled = c.LinkingEnabled
		out.LastWebhookEvent = c.LastWebhookEvent
		if c.LastWebhookAt != nil {
			out.LastWebhookAt = c.LastWebhookAt.UTC().Format(time.RFC3339)
		}
	}
	return out
}

// requireFlowMember admits anyone with Flow access. Linking a GitHub account
// is personal, so it needs no project or business role.
func requireFlowMember(ctx context.Context) (int64, *authPkg.AuthUser, error) {
	user, ok := authPkg.GetAuthUser(ctx)
	if !ok || user == nil || !authPkg.HasAnyAppScope(user, authorization.AppFlow) {
		return 0, nil, onekiterrors.PermissionDenied()
	}
	return resolveBusinessIDV2(ctx)
}

func toGitHubAccountLink(link domain.GitHubUserLink) *flowgithubv2.GitHubAccountLink {
	return &flowgithubv2.GitHubAccountLink{
		Linked:       true,
		GithubUserId: link.GitHubUserID,
		Login:        link.Login,
		AvatarUrl:    link.AvatarURL,
		LinkedAt:     link.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func toProjectRepositories(repos []domain.ProjectRepository) []*flowgithubv2.ProjectRepository {
	out := make([]*flowgithubv2.ProjectRepository, 0, len(repos))
	for _, r := range repos {
		out = append(out, &flowgithubv2.ProjectRepository{
			RepositoryId: r.RepositoryID,
			FullName:     r.FullName,
			Private:      r.Private,
			CreatedAt:    r.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return out
}

func (a *GitHubAPIV2) GetGitHubConnection(ctx context.Context, _ *flowgithubv2.GetGitHubConnectionRequest) (*flowgithubv2.GetGitHubConnectionResponse, error) {
	businessID, user, err := resolveBusinessIDV2(ctx)
	if err != nil {
		return nil, err
	}
	conn, err := a.Service.GitHubConnection(ctx, businessID)
	if err != nil {
		return nil, githubError(ctx, err, "GetGitHubConnection failed")
	}
	return &flowgithubv2.GetGitHubConnectionResponse{Connection: toGitHubConnection(conn, authPkg.IsBusinessAdmin(user))}, nil
}

func (a *GitHubAPIV2) StartGitHubInstall(ctx context.Context, _ *flowgithubv2.StartGitHubInstallRequest) (*flowgithubv2.StartGitHubInstallResponse, error) {
	businessID, user, err := requireBusinessAdmin(ctx)
	if err != nil {
		return nil, err
	}
	installURL, err := a.Service.StartGitHubInstall(businessID, user.UserID)
	if err != nil {
		return nil, githubError(ctx, err, "StartGitHubInstall failed")
	}
	return &flowgithubv2.StartGitHubInstallResponse{InstallUrl: installURL}, nil
}

func (a *GitHubAPIV2) CompleteGitHubInstall(ctx context.Context, req *flowgithubv2.CompleteGitHubInstallRequest) (*flowgithubv2.CompleteGitHubInstallResponse, error) {
	businessID, user, err := requireBusinessAdmin(ctx)
	if err != nil {
		return nil, err
	}
	conn, err := a.Service.CompleteGitHubInstall(ctx, businessID, user.UserID, req.GetInstallationId(), req.GetState())
	if err != nil {
		return nil, githubError(ctx, err, "CompleteGitHubInstall failed")
	}
	return &flowgithubv2.CompleteGitHubInstallResponse{Connection: toGitHubConnection(conn, true)}, nil
}

func (a *GitHubAPIV2) DisconnectGitHub(ctx context.Context, _ *flowgithubv2.DisconnectGitHubRequest) (*flowgithubv2.DisconnectGitHubResponse, error) {
	businessID, _, err := requireBusinessAdmin(ctx)
	if err != nil {
		return nil, err
	}
	uninstalled, err := a.Service.DisconnectGitHub(ctx, businessID)
	if err != nil {
		return nil, githubError(ctx, err, "DisconnectGitHub failed")
	}
	return &flowgithubv2.DisconnectGitHubResponse{Uninstalled: uninstalled}, nil
}

func (a *GitHubAPIV2) ListGitHubRepositories(ctx context.Context, req *flowgithubv2.ListGitHubRepositoriesRequest) (*flowgithubv2.ListGitHubRepositoriesResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:edit"); err != nil {
		return nil, err
	}
	businessID, user, err := resolveBusinessIDV2(ctx)
	if err != nil {
		return nil, err
	}
	repos, err := a.Service.ListGitHubRepositories(ctx, req.GetProjectId(), user.UserID, businessID)
	if err != nil {
		return nil, githubError(ctx, err, "ListGitHubRepositories failed")
	}
	out := make([]*flowgithubv2.GitHubRepository, 0, len(repos))
	for _, r := range repos {
		out = append(out, &flowgithubv2.GitHubRepository{Id: r.ID, FullName: r.FullName, Private: r.Private})
	}
	return &flowgithubv2.ListGitHubRepositoriesResponse{Repositories: out}, nil
}

func (a *GitHubAPIV2) ListProjectRepositories(ctx context.Context, req *flowgithubv2.ListProjectRepositoriesRequest) (*flowgithubv2.ListProjectRepositoriesResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:view"); err != nil {
		return nil, err
	}
	businessID, user, err := resolveBusinessIDV2(ctx)
	if err != nil {
		return nil, err
	}
	repos, err := a.Service.ListProjectRepositories(ctx, req.GetProjectId(), user.UserID, businessID)
	if err != nil {
		return nil, githubError(ctx, err, "ListProjectRepositories failed")
	}
	return &flowgithubv2.ListProjectRepositoriesResponse{Repositories: toProjectRepositories(repos)}, nil
}

func (a *GitHubAPIV2) SetProjectRepositories(ctx context.Context, req *flowgithubv2.SetProjectRepositoriesRequest) (*flowgithubv2.SetProjectRepositoriesResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:edit"); err != nil {
		return nil, err
	}
	businessID, user, err := resolveBusinessIDV2(ctx)
	if err != nil {
		return nil, err
	}
	repos, err := a.Service.SetProjectRepositories(ctx, req.GetProjectId(), user.UserID, businessID, req.GetRepositoryIds())
	if err != nil {
		return nil, githubError(ctx, err, "SetProjectRepositories failed")
	}
	return &flowgithubv2.SetProjectRepositoriesResponse{Repositories: toProjectRepositories(repos)}, nil
}

func (a *GitHubAPIV2) GetMyGitHubLink(ctx context.Context, _ *flowgithubv2.GetMyGitHubLinkRequest) (*flowgithubv2.GetMyGitHubLinkResponse, error) {
	businessID, user, err := requireFlowMember(ctx)
	if err != nil {
		return nil, err
	}
	link, err := a.Service.GitHubUserLink(ctx, businessID, user.UserID)
	if errors.Is(err, domain.ErrGitHubNotLinked) {
		return &flowgithubv2.GetMyGitHubLinkResponse{Link: &flowgithubv2.GitHubAccountLink{}}, nil
	}
	if err != nil {
		return nil, githubError(ctx, err, "GetMyGitHubLink failed")
	}
	return &flowgithubv2.GetMyGitHubLinkResponse{Link: toGitHubAccountLink(link)}, nil
}

func (a *GitHubAPIV2) StartGitHubLink(ctx context.Context, _ *flowgithubv2.StartGitHubLinkRequest) (*flowgithubv2.StartGitHubLinkResponse, error) {
	businessID, user, err := requireFlowMember(ctx)
	if err != nil {
		return nil, err
	}
	authorizeURL, err := a.Service.StartGitHubLink(ctx, businessID, user.UserID)
	if err != nil {
		return nil, githubError(ctx, err, "StartGitHubLink failed")
	}
	return &flowgithubv2.StartGitHubLinkResponse{AuthorizeUrl: authorizeURL}, nil
}

func (a *GitHubAPIV2) CompleteGitHubLink(ctx context.Context, req *flowgithubv2.CompleteGitHubLinkRequest) (*flowgithubv2.CompleteGitHubLinkResponse, error) {
	businessID, user, err := requireFlowMember(ctx)
	if err != nil {
		return nil, err
	}
	link, err := a.Service.CompleteGitHubLink(ctx, businessID, user.UserID, req.GetCode(), req.GetState())
	if err != nil {
		return nil, githubError(ctx, err, "CompleteGitHubLink failed")
	}
	return &flowgithubv2.CompleteGitHubLinkResponse{Link: toGitHubAccountLink(link)}, nil
}

func (a *GitHubAPIV2) UnlinkGitHub(ctx context.Context, _ *flowgithubv2.UnlinkGitHubRequest) (*flowgithubv2.UnlinkGitHubResponse, error) {
	businessID, user, err := requireFlowMember(ctx)
	if err != nil {
		return nil, err
	}
	if err := a.Service.UnlinkGitHub(ctx, businessID, user.UserID); err != nil {
		return nil, githubError(ctx, err, "UnlinkGitHub failed")
	}
	return &flowgithubv2.UnlinkGitHubResponse{}, nil
}
