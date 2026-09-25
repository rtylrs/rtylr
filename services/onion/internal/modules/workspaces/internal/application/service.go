package application

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/hubprofile"
	"github.com/voxire/onion/internal/platform/labourcost"
	"github.com/voxire/onion/internal/platform/orgchart"
)

const (
	workspacesScope = "workspaces"
	workspacesTTL   = 2 * time.Minute

	// internalListCap bounds list reads made on behalf of another feature
	// (the dashboard's "my tasks" widget, a task's subtasks embedded in its
	// detail view, a project's columns re-fetched right after creating one)
	// rather than a paginated list RPC. Matches the cap these reads already
	// used before pagination existed.
	internalListCap = int64(500)

	// taskUpdatedEvent is the realtime type the task write paths already
	// publish on the task and project topics. The user-topic publish reuses it
	// instead of minting a `task.assigned`: the events carry no payload,
	// consumers branch on the `task.` prefix to invalidate their task queries,
	// and a second type would be one more string for every consumer to learn
	// carrying no extra information.
	taskUpdatedEvent = "task.updated"

	// assignedNotification is the flow_notifications.type enum value for "this
	// task is yours now".
	assignedNotification = "assigned"

	// assignedNotificationTitle is the headline stored on the inbox row. The
	// column holds a plain string, so the copy is written once here rather than
	// resolved per reader; the task title, project and actor travel in their own
	// columns alongside it.
	assignedNotificationTitle = "Assigned to you"

	commentNotification      = "comment"
	commentNotificationTitle = "New comment on your task"
	mentionNotification      = "mention"
	mentionNotificationTitle = "You were mentioned in a task"
	statusNotification       = "status_change"
	statusNotificationTitle  = "Task status changed"
)

func invalidateWorkspacesScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(workspacesScope, businessID))
}

// invalidateUserScope clears the per-user cache bucket, `cache:user:<id>:*`.
// GetMyTasks caches there because the list is scoped to the person reading it
// rather than to the business that owns the task, which puts it out of reach of
// invalidateWorkspacesScope: the two patterns share no segment.
//
// This has to stay a prefix delete. Page and limit are baked into every
// my-tasks key, so deleting one exact key would leave every other cached page
// serving pre-mutation rows until the TTL expires.
//
// userID 0 means "unassigned" on the task write paths; there is no reader to
// evict, so skip the SCAN rather than wipe `cache:user:0:*`.
func invalidateUserScope(ctx context.Context, userID int64) {
	if userID == 0 {
		return
	}
	cacher.InvalidateScope(ctx, cacheKeys.UserPrefix(userID))
}

// assigneeID flattens an optional assignee to a plain id, 0 meaning unassigned,
// so the task write paths can hand it straight to invalidateUserScope.
func assigneeID(id *int64) int64 {
	if id == nil {
		return 0
	}
	return *id
}

type Service struct {
	repo *infrastructure.Repository

	// labourCosts is optional because workers and tests can construct the
	// module without payroll. The no-op implementation keeps those consumers
	// safe while the API bootstrap wires the HR-owned provider.
	labourCosts labourcost.Provider

	// hubProfiles resolves the central effective-access payload used by Flow's
	// business-wide reporting visibility. Project and workspace authorization
	// remains owned by Flow resource membership.
	hubProfiles hubprofile.Resolver

	// orgchart resolves who reports to a caller, backing the dept-manager
	// visibility proxy (M3). Optional in exactly the way hubProfiles is: nil
	// until SetOrgchartResolver is called, in which case the proxy is skipped
	// and visibility behaves as it did before it existed. A nil provider must
	// never widen access, only fail to widen it.
	orgchart orgchart.ManagerResolver
}

func NewService(_ context.Context, db *sql.DB) *Service {
	repo := infrastructure.NewRepository(db)

	return &Service{
		repo:        repo,
		labourCosts: noopLabourCostProvider{},
	}
}

// SetLabourCostProvider wires the HR-owned aggregate port. Consumers receive
// cents and rate-coverage counts only; hourly rates never enter workspaces.
func (s *Service) SetLabourCostProvider(provider labourcost.Provider) {
	if provider != nil {
		s.labourCosts = provider
	}
}

// GetProjectLabourCost returns the live tenant-scoped labour aggregate for a
// project. The result contains no employee compensation fields.
func (s *Service) GetProjectLabourCost(ctx context.Context, businessID, projectID int64) (labourcost.ProjectCost, error) {
	return s.labourCosts.GetProjectLabourCost(ctx, businessID, projectID)
}

type noopLabourCostProvider struct{}

func (noopLabourCostProvider) GetProjectLabourCost(_ context.Context, businessID, projectID int64) (labourcost.ProjectCost, error) {
	return labourcost.ProjectCost{BusinessID: businessID, ProjectID: projectID}, nil
}

// SetHubProfileResolver wires the central effective-access lookup used by
// business-wide reporting visibility. It does not bypass Flow resource scope.
func (s *Service) SetHubProfileResolver(r hubprofile.Resolver) {
	s.hubProfiles = r
}

// SetOrgchartResolver wires the reporting-line lookup behind the dept-manager
// visibility proxy. Until it is called the proxy is inert.
func (s *Service) SetOrgchartResolver(r orgchart.ManagerResolver) {
	s.orgchart = r
}

// managedUserIDs returns the users reporting to the caller, or nil when the
// proxy cannot apply.
//
// A resolver error is swallowed to nil rather than propagated. That is
// deliberate: this widens visibility, so failing to resolve must degrade to the
// caller's own projects, never to an error on a list they are entitled to see.
func (s *Service) managedUserIDs(ctx context.Context, callerID, businessID int64) []int64 {
	if s.orgchart == nil || callerID <= 0 || businessID <= 0 {
		return nil
	}
	ids, err := s.orgchart.ManagedUserIDs(ctx, callerID, businessID)
	if err != nil {
		return nil
	}
	return ids
}

// ─────────────────────────────────────────────
// Workspaces
// ─────────────────────────────────────────────

type workspacesPage struct {
	Workspaces []domain.Workspace
	Total      int64
}

func (s *Service) ListWorkspaces(ctx context.Context, businessID, page, limit int64) ([]domain.Workspace, int64, error) {
	offset := (page - 1) * limit
	result, err := cacher.Of[workspacesPage]().GetOrComputeSingle(ctx,
		// page/limit are part of the cache key, not just the fetch - otherwise
		// page 2 would return page 1's cached rows until the TTL expires.
		cacheKeys.Scoped(workspacesScope, businessID, fmt.Sprintf("list:page=%d:limit=%d", page, limit)),
		func(ctx context.Context) (workspacesPage, error) {
			workspaces, total, fetchErr := s.repo.GetWorkspacesByOwner(ctx, businessID, limit, offset)
			return workspacesPage{Workspaces: workspaces, Total: total}, fetchErr
		},
		cacher.WithExpiry(ctx, workspacesTTL),
	)
	if err != nil {
		return nil, 0, err
	}
	return result.Workspaces, result.Total, nil
}

func (s *Service) CreateWorkspace(ctx context.Context, input domain.WorkspaceInput) (domain.Workspace, error) {
	defer invalidateWorkspacesScope(ctx, input.BusinessID)
	count, err := s.repo.CountWorkspacesByOwner(ctx, input.BusinessID)
	if err != nil {
		return domain.Workspace{}, err
	}

	isDefault := count == 0

	id, err := s.repo.CreateWorkspace(ctx, input, isDefault)
	if err != nil {
		return domain.Workspace{}, err
	}

	return domain.Workspace{
		ID:          id,
		BusinessID:  input.BusinessID,
		OwnerID:     input.OwnerID,
		Name:        input.Name,
		Description: input.Description,
		Icon:        input.Icon,
		Color:       input.Color,
		IsDefault:   isDefault,
		CreatedBy:   input.CreatedBy,
	}, nil
}

func (s *Service) UpdateWorkspace(ctx context.Context, input domain.UpdateWorkspaceInput) error {
	defer invalidateWorkspacesScope(ctx, input.BusinessID)
	return s.repo.UpdateWorkspace(ctx, input)
}

func (s *Service) DeleteWorkspace(ctx context.Context, id, businessID int64) error {
	defer invalidateWorkspacesScope(ctx, businessID)
	return s.repo.DeleteWorkspace(ctx, id, businessID)
}

func (s *Service) AddWorkspaceMember(ctx context.Context, input domain.AddWorkspaceMemberInput) error {
	return s.repo.AddWorkspaceMember(ctx, input)
}

func (s *Service) RemoveWorkspaceMember(ctx context.Context, input domain.RemoveWorkspaceMemberInput) error {
	return s.repo.RemoveWorkspaceMember(ctx, input)
}

func (s *Service) ListWorkspaceMembers(ctx context.Context, workspaceID, limit, offset int64) ([]domain.WorkspaceMember, int64, error) {
	return s.repo.ListWorkspaceMembers(ctx, workspaceID, limit, offset)
}

func (s *Service) UpdateWorkspaceMemberRole(ctx context.Context, input domain.UpdateWorkspaceMemberRoleInput) error {
	return s.repo.UpdateWorkspaceMemberRole(ctx, input)
}

func (s *Service) CreatePersonalWorkspace(ctx context.Context, userID, businessID int64) (int64, error) {
	return s.repo.CreatePersonalWorkspace(ctx, userID, businessID)
}

// ResolveWorkspaceID converts a workspace_id coming from the frontend into
// a numeric DB ID. The frontend uses the sentinel string "personal" to mean
// "the user's Personal workspace", which is created on demand. Numeric IDs
// are accepted only for workspaces the caller owns or belongs to. businessID
// must be the caller's resolved tenant id, used only on the "personal" path.
func (s *Service) ResolveWorkspaceID(ctx context.Context, userID, businessID int64, raw string) (int64, error) {
	if raw == "" {
		return 0, nil
	}
	if raw == "personal" {
		return s.repo.CreatePersonalWorkspace(ctx, userID, businessID)
	}
	workspaceID, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || workspaceID <= 0 {
		return 0, fmt.Errorf("invalid workspace id")
	}
	if !s.repo.IsWorkspaceMember(ctx, userID, workspaceID) {
		return 0, fmt.Errorf("workspace access denied")
	}
	return workspaceID, nil
}
