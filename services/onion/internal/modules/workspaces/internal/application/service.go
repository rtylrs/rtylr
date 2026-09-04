package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/hubprofile"
	"github.com/voxire/onion/internal/platform/labourcost"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/orgchart"
	"github.com/voxire/onion/internal/platform/wsbus"
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

// IsWorkspaceMember reports whether userID is on the workspace's member
// roster or owns it. Used by the realtime hub to authorize
// `workspace:<id>` and `project:<id>` topic subscriptions.
func (s *Service) IsWorkspaceMember(ctx context.Context, userID, workspaceID int64) bool {
	return s.repo.IsWorkspaceMember(ctx, userID, workspaceID)
}

// IsProjectMember reports whether userID can access the project — either
// because they're a project member, a workspace member of the project's
// workspace, or its owner. Used for realtime topic auth.
func (s *Service) IsProjectMember(ctx context.Context, userID, projectID int64) bool {
	return s.repo.IsProjectMember(ctx, userID, projectID)
}

// IsTaskMember reports whether userID can subscribe to a task realtime topic.
func (s *Service) IsTaskMember(ctx context.Context, userID, taskID int64) bool {
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return false
	}
	if task.ProjectID != 0 {
		return s.repo.IsProjectMember(ctx, userID, task.ProjectID)
	}
	if task.WorkspaceID != 0 {
		return s.repo.IsWorkspaceMember(ctx, userID, task.WorkspaceID)
	}
	return task.OwnerID == userID
}

// ─────────────────────────────────────────────
// Authorization
// ─────────────────────────────────────────────

// projectRoleFor resolves a caller's effective role on a project, or "" when
// they hold none.
//
// A flow_project_members row is only one of three ways to reach a project:
// IsProjectMember also admits the project's owner and any member of the
// project's workspace, and neither necessarily has a row — CreateProject
// writes none for the owner. Resolving on the members table alone would deny
// both groups access they have today, so the lookup falls back through
// ownership and then the workspace roster.
func (s *Service) projectRoleFor(ctx context.Context, projectID, userID int64) string {
	if projectID == 0 || userID == 0 {
		return ""
	}
	if role, err := s.repo.GetProjectMemberRole(ctx, projectID, userID); err == nil && role != "" {
		return role
	}

	ownerID, workspaceID, err := s.repo.GetProjectOwnerWorkspace(ctx, projectID)
	if err != nil {
		return ""
	}
	if ownerID == userID {
		return domain.FlowRoleAdmin
	}
	if role := s.managerProxyRoleFor(ctx, projectID, userID); role != "" {
		return role
	}
	return s.workspaceRoleFor(ctx, workspaceID, userID)
}

// managerProxyRoleFor grants a department manager member-equivalent rights on a
// project someone reporting to them is involved in (M3), or "" when the proxy
// does not apply.
//
// Member and not admin: the manager can do what a member can, which is the
// decided scope. Admin would let them archive or transfer ownership of a
// project nobody added them to.
//
// This needs the project's business, which projectRoleFor does not carry, so it
// reads it rather than taking it on trust. A project outside the caller's
// business must not resolve through the proxy.
func (s *Service) managerProxyRoleFor(ctx context.Context, projectID, userID int64) string {
	if s.orgchart == nil {
		return ""
	}
	p, err := s.repo.GetProjectProxyFacts(ctx, projectID)
	if err != nil {
		return ""
	}
	if s.managesSomeoneOn(ctx, p, userID, p.BusinessID) {
		return domain.FlowRoleMember
	}
	return ""
}

// workspaceRoleFor resolves a caller's role on a workspace. The owner holds no
// flow_workspace_members row either, so they are resolved separately.
func (s *Service) workspaceRoleFor(ctx context.Context, workspaceID, userID int64) string {
	if workspaceID == 0 || userID == 0 {
		return ""
	}
	if ownerID, err := s.repo.GetWorkspaceOwner(ctx, workspaceID); err == nil && ownerID == userID {
		return domain.FlowRoleAdmin
	}
	role, err := s.repo.GetUserWorkspaceRole(ctx, workspaceID, userID)
	if err != nil {
		return ""
	}
	return role
}

// hubProfileBypasses reports whether the central Flow app permission permits
// business-wide reporting visibility. It is intentionally not consulted by
// project, workspace, or task resource checks.
func (s *Service) hubProfileBypasses(ctx context.Context, userID, businessID int64) bool {
	if s.hubProfiles == nil {
		return false
	}
	profile, err := s.hubProfiles.GetProfile(ctx, userID, businessID)
	if err != nil {
		return false
	}
	if profile.EffectiveAccess == nil {
		return false
	}
	for _, app := range profile.EffectiveAccess.Apps {
		if app.AppID != "flow" {
			continue
		}
		for _, permission := range app.Permissions {
			if permission == "flow:team:manage" {
				return true
			}
		}
	}
	return false
}

// RequireProjectRole reports whether userID holds at least minRole on a
// project, returning domain.ErrUnauthorized when they do not. Flow project
// membership and project scope are the only source of this authorization.
func (s *Service) RequireProjectRole(ctx context.Context, projectID, userID, _ int64, minRole string) error {
	if domain.RoleAllows(s.projectRoleFor(ctx, projectID, userID), minRole) {
		return nil
	}
	return domain.ErrUnauthorized
}

// RequireWorkspaceRole reports whether userID holds at least minRole on a
// workspace, including the workspace owner who has no membership row.
func (s *Service) RequireWorkspaceRole(ctx context.Context, workspaceID, userID, _ int64, minRole string) error {
	if domain.RoleAllows(s.workspaceRoleFor(ctx, workspaceID, userID), minRole) {
		return nil
	}
	return domain.ErrUnauthorized
}

// RequireTaskRole is RequireProjectRole addressed by task rather than project.
// A task need not belong to a project, so the fallback order mirrors
// IsTaskMember: project, then workspace, then sole ownership. Business-wide
// visibility does not bypass task, project, or workspace scope.
func (s *Service) RequireTaskRole(ctx context.Context, taskID, userID, businessID int64, minRole string) error {
	if taskID == 0 || userID == 0 {
		return domain.ErrUnauthorized
	}
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return domain.ErrUnauthorized
	}

	if task.ProjectID != 0 {
		return s.RequireProjectRole(ctx, task.ProjectID, userID, businessID, minRole)
	}
	if task.WorkspaceID != 0 {
		if domain.RoleAllows(s.workspaceRoleFor(ctx, task.WorkspaceID, userID), minRole) {
			return nil
		}
		return domain.ErrUnauthorized
	}
	if task.OwnerID == userID {
		return nil
	}
	return domain.ErrUnauthorized
}

// RequireCommentRole authorizes against the project owning a comment's task.
func (s *Service) RequireCommentRole(ctx context.Context, commentID, userID, businessID int64, minRole string) error {
	taskID, err := s.repo.GetCommentTaskID(ctx, commentID)
	if err != nil {
		return domain.ErrUnauthorized
	}
	return s.RequireTaskRole(ctx, taskID, userID, businessID, minRole)
}

// RequireAttachmentRole authorizes against the project owning an attachment's task.
func (s *Service) RequireAttachmentRole(ctx context.Context, attachmentID, userID, businessID int64, minRole string) error {
	taskID, err := s.repo.GetAttachmentTaskID(ctx, attachmentID)
	if err != nil {
		return domain.ErrUnauthorized
	}
	return s.RequireTaskRole(ctx, taskID, userID, businessID, minRole)
}

// RequireTimeLogRole authorizes against the project owning a time log's task.
func (s *Service) RequireTimeLogRole(ctx context.Context, timeLogID, userID, businessID int64, minRole string) error {
	taskID, err := s.repo.GetTimeLogTaskID(ctx, timeLogID)
	if err != nil {
		return domain.ErrUnauthorized
	}
	return s.RequireTaskRole(ctx, taskID, userID, businessID, minRole)
}

// ─────────────────────────────────────────────
// Projects
// ─────────────────────────────────────────────

func (s *Service) ListProjects(ctx context.Context, workspaceID, businessID, limit, offset int64) ([]domain.Project, int64, error) {
	return s.repo.GetProjectsByWorkspace(ctx, workspaceID, businessID, limit, offset)
}

func (s *Service) CreateProject(ctx context.Context, input domain.ProjectInput) (domain.Project, error) {
	defer invalidateWorkspacesScope(ctx, input.BusinessID)
	// Ownerless projects break RequireProjectRole's ownership fallback, so the
	// service owns this invariant instead of trusting every caller.
	if input.OwnerID == 0 {
		input.OwnerID = input.CreatedBy
	}
	if input.Visibility == "" {
		input.Visibility = domain.VisibilityPrivate
	}
	if !domain.ValidVisibility(input.Visibility) {
		return domain.Project{}, errors.New("invalid project visibility")
	}
	id, err := s.repo.CreateProject(ctx, input)
	if err != nil {
		return domain.Project{}, err
	}

	// Seed default columns for every new project
	_ = s.repo.SeedDefaultColumns(ctx, id, input.WorkspaceID, input.CreatedBy)

	return domain.Project{
		ID:          id,
		WorkspaceID: input.WorkspaceID,
		OwnerID:     input.OwnerID,
		BusinessID:  input.BusinessID,
		Name:        input.Name,
		Description: input.Description,
		Icon:        input.Icon,
		Color:       input.Color,
		Status:      "active",
		Visibility:  input.Visibility,
		IsPersonal:  input.IsPersonal,
		StartDate:   input.StartDate,
		DueDate:     input.DueDate,
		CreatedBy:   input.CreatedBy,
	}, nil
}

func (s *Service) UpdateProject(ctx context.Context, input domain.UpdateProjectInput) error {
	defer invalidateWorkspacesScope(ctx, input.BusinessID)
	return s.repo.UpdateProject(ctx, input)
}

func (s *Service) DeleteProject(ctx context.Context, id, businessID int64) error {
	defer invalidateWorkspacesScope(ctx, businessID)
	// Archive now carries recycle-bin semantics: stamp the permanent-deletion
	// deadline from the retention setting so every client sees the same clock.
	deadline := binPurgeDeadline(time.Now().UTC())
	return s.repo.ArchiveProject(ctx, id, businessID, &deadline)
}

func (s *Service) ListProjectMembers(ctx context.Context, projectID, limit, offset int64) ([]domain.ProjectMember, int64, error) {
	return s.repo.ListProjectMembers(ctx, projectID, limit, offset)
}

func (s *Service) AddProjectMember(ctx context.Context, input domain.AddProjectMemberInput) error {
	return s.repo.AddProjectMember(ctx, input)
}

func (s *Service) RemoveProjectMember(ctx context.Context, input domain.RemoveProjectMemberInput) error {
	return s.repo.RemoveProjectMember(ctx, input)
}

func (s *Service) UpdateProjectMemberRole(ctx context.Context, input domain.UpdateProjectMemberRoleInput) error {
	return s.repo.UpdateProjectMemberRole(ctx, input)
}

func (s *Service) TransferProjectOwnership(ctx context.Context, input domain.TransferProjectOwnershipInput) error {
	return s.repo.TransferProjectOwnership(ctx, input)
}

func (s *Service) MoveProject(ctx context.Context, input domain.MoveProjectInput) error {
	return s.repo.MoveProject(ctx, input)
}

func (s *Service) InviteProjectMemberByEmail(ctx context.Context, projectID int64, email, role string, addedBy int64) (bool, error) {
	userID, found, err := s.repo.LookupUserByEmail(ctx, email)
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}
	return true, s.repo.AddProjectMember(ctx, domain.AddProjectMemberInput{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
		AddedBy:   addedBy,
	})
}

// ─────────────────────────────────────────────
// Tasks
// ─────────────────────────────────────────────

func (s *Service) ListTasks(ctx context.Context, filter domain.ListTasksFilter) ([]domain.Task, int64, error) {
	if filter.ProjectID == 0 && filter.WorkspaceID != 0 {
		return s.repo.GetTasksByWorkspace(ctx, filter)
	}
	return s.repo.GetTasksByProject(ctx, filter)
}

func (s *Service) CreateTask(ctx context.Context, input domain.TaskInput) (domain.Task, error) {
	defer invalidateWorkspacesScope(ctx, input.BusinessID)

	// R13: an omitted project means "file into my My Tasks bucket", which is
	// lazily provisioned on first use. Every task belongs to exactly one
	// project; "outside a project" is a UI concept.
	if input.ProjectID == 0 {
		bucketID, _, ensureErr := s.EnsureMyTasksProject(ctx, input.OwnerID, input.BusinessID, 0)
		if ensureErr != nil {
			return domain.Task{}, ensureErr
		}
		input.ProjectID = bucketID
	}

	// T1: assignment carries access with it.
	if input.AssigneeID != nil && *input.AssigneeID != input.OwnerID {
		if grantErr := s.EnsureAssigneeMembership(ctx, input.ProjectID, input.BusinessID, *input.AssigneeID); grantErr != nil {
			return domain.Task{}, grantErr
		}
	}

	// A new task lands in the assignee's my-tasks list, whose cached pages were
	// computed before it existed.
	defer invalidateUserScope(ctx, assigneeID(input.AssigneeID))
	id, err := s.repo.CreateTask(ctx, input)
	if err != nil {
		return domain.Task{}, err
	}

	if actErr := s.repo.LogActivity(ctx, id, input.ReporterID, "created_task", "", input.Title); actErr != nil {
		logger.Ctx(ctx).Error("CreateTask: failed to log activity", logger.Error(actErr))
	}

	return domain.Task{
		ID:           id,
		ProjectID:    input.ProjectID,
		WorkspaceID:  input.WorkspaceID,
		OwnerID:      input.OwnerID,
		BusinessID:   input.BusinessID,
		ParentTaskID: input.ParentTaskID,
		Title:        input.Title,
		Description:  input.Description,
		Status:       "todo",
		Priority:     input.Priority,
		AssigneeID:   input.AssigneeID,
		ReporterID:   input.ReporterID,
		DueDate:      input.DueDate,
		StartDate:    input.StartDate,
		Labels:       input.Labels,
	}, nil
}

func (s *Service) GetTaskDetail(ctx context.Context, taskID int64) (domain.TaskDetail, error) {
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return domain.TaskDetail{}, err
	}

	subtasks, _, err := s.repo.GetSubtasks(ctx, taskID, internalListCap, 0)
	if err != nil {
		return domain.TaskDetail{}, err
	}

	comments, err := s.repo.GetCommentsByTask(ctx, taskID)
	if err != nil {
		return domain.TaskDetail{}, err
	}

	activity, err := s.repo.GetActivityByTask(ctx, taskID, 50)
	if err != nil {
		return domain.TaskDetail{}, err
	}

	dependencyMap, err := s.repo.ListTaskDependencies(ctx, []int64{taskID}, task.BusinessID)
	if err != nil {
		return domain.TaskDetail{}, err
	}
	task.Dependencies = dependencyMap[taskID]
	if task.Dependencies == nil {
		task.Dependencies = []int64{}
	}

	return domain.TaskDetail{
		Task:     task,
		Subtasks: subtasks,
		Comments: comments,
		Activity: activity,
	}, nil
}

func (s *Service) UpdateTask(ctx context.Context, input domain.UpdateTaskInput, userID int64) error {
	// Registered ahead of the cache evictions below so that it runs after all
	// of them: announcing tells the affected clients to refetch, and that
	// refetch must not be able to find a pre-update my-tasks page still cached.
	// Stays nil unless the update both changed the assignee and landed.
	var announced *assigneeChange
	defer func() {
		if announced != nil {
			s.announceAssigneeChange(ctx, *announced)
		}
	}()

	defer invalidateWorkspacesScope(ctx, input.BusinessID)
	existing, err := s.repo.GetTaskByID(ctx, input.ID)
	if err != nil {
		return err
	}

	// Any edit changes what the current assignee's my-tasks list should show,
	// title and status included, so their cached pages go regardless of whether
	// this update touches the assignee.
	defer invalidateUserScope(ctx, assigneeID(existing.AssigneeID))

	// This is a partial update — only record a change for a field the request
	// actually supplied. An omitted field arrives as the zero value (empty
	// string / nil) and means "leave unchanged", so comparing it against the
	// existing value would otherwise log a spurious "changed to empty" entry.
	//
	// The entries are computed here, from existing versus input, and written
	// only once the update has landed. Each LogActivity is its own write with
	// no transaction around it, so writing them first left false history behind
	// every time the update then failed: the task stayed as it was and its
	// Activity tab claimed changes that never happened.
	var activity []taskActivityEntry

	if input.Status != "" && existing.Status != input.Status {
		activity = append(activity, taskActivityEntry{"status_changed", existing.Status, input.Status})
	}

	if input.Priority != "" && existing.Priority != input.Priority {
		activity = append(activity, taskActivityEntry{"priority_changed", existing.Priority, input.Priority})
	}

	var reassignment *assigneeChange
	if input.AssigneeID != nil {
		existingAssignee := int64(0)
		if existing.AssigneeID != nil {
			existingAssignee = *existing.AssigneeID
		}
		newAssignee := *input.AssigneeID

		if existingAssignee != newAssignee {
			activity = append(activity, taskActivityEntry{
				"assignee_changed",
				strconv.FormatInt(existingAssignee, 10),
				strconv.FormatInt(newAssignee, 10),
			})
			// Both sides of a reassignment need evicting. The previous
			// assignee's eviction is already deferred above; without this one
			// the new assignee keeps serving pages that predate the task
			// arriving in their list.
			defer invalidateUserScope(ctx, newAssignee)

			// The title comes from the request when the same call renames the
			// task, so the inbox row names what the assignee will actually see.
			title := existing.Title
			if input.Title != "" {
				title = input.Title
			}
			reassignment = &assigneeChange{
				taskID:       input.ID,
				workspaceID:  existing.WorkspaceID,
				taskTitle:    title,
				project:      existing.ProjectName,
				projectColor: existing.ProjectColor,
				actorID:      userID,
				previous:     existingAssignee,
				next:         newAssignee,
			}
		}
	}

	if updateErr := s.repo.UpdateTask(ctx, input); updateErr != nil {
		return updateErr
	}

	// The only place UpdateTask writes activity, and it is past the point of
	// failure. A logging failure is reported and swallowed as before: the edit
	// itself succeeded, and losing a history row is not a reason to tell the
	// caller their change did not land.
	for _, entry := range activity {
		if actErr := s.repo.LogActivity(ctx, input.ID, userID, entry.kind, entry.from, entry.to); actErr != nil {
			logger.Ctx(ctx).Error("UpdateTask: failed to log activity",
				logger.Any("kind", entry.kind), logger.Error(actErr))
		}
	}

	if input.Status != "" && existing.Status != input.Status {
		notificationTask := existing
		notificationTask.Status = input.Status
		if input.Title != "" {
			notificationTask.Title = input.Title
		}
		if input.AssigneeID != nil {
			notificationTask.AssigneeID = input.AssigneeID
		}
		s.announceStatusChange(ctx, notificationTask, userID, input.Status)
	}

	announced = reassignment
	return nil
}

// taskActivityEntry is one history row an update will produce, held between
// computing it from the request and writing it after the update succeeds.
type taskActivityEntry struct {
	kind string
	from string
	to   string
}

// assigneeChange is what a reassignment needs in order to announce itself once
// the write has landed: who lost the task, who picked it up, and enough of the
// task to render an inbox row without reading it back.
type assigneeChange struct {
	taskID       int64
	workspaceID  int64
	taskTitle    string
	project      string
	projectColor string
	actorID      int64
	previous     int64
	next         int64
}

// announceAssigneeChange is the realtime half of a reassignment.
//
// Both sides get a task event on their own `user:<id>` topic. The previous
// assignee's my-tasks list has to drop the task and the new assignee's has to
// pick it up, and neither of them is subscribed to the `task:<id>` or
// `project:<id>` topics the write already publishes on. A client sitting on My
// tasks or Today subscribes to `workspace:<id>` and `user:<id>` only, so a
// reassignment reached it through no topic at all.
//
// The new assignee also gets an inbox row, which is what makes "you were
// assigned a task" visible outside the flow board. That row is skipped when
// they turned task-assignment notifications off, and when they assigned the
// task to themselves and already know. The realtime event is deliberately not
// gated on either: it is an invalidation hint that keeps a list correct rather
// than a notification, and muting an inbox should not leave My tasks stale.
func (s *Service) announceAssigneeChange(ctx context.Context, change assigneeChange) {
	publishUserTaskEvent(change.previous, change.taskID, change.actorID)
	publishUserTaskEvent(change.next, change.taskID, change.actorID)

	if change.next == 0 || change.next == change.actorID {
		return
	}

	settings, err := s.repo.GetUserSettings(ctx, change.next)
	if err != nil {
		logger.Ctx(ctx).Error("announceAssigneeChange: failed to read notification settings, no inbox row created",
			logger.Error(err))
		return
	}
	if !settings.TaskAssigned {
		return
	}

	notification := domain.CreateNotificationInput{
		UserID:       change.next,
		Type:         assignedNotification,
		Title:        assignedNotificationTitle,
		TaskID:       &change.taskID,
		TaskTitle:    change.taskTitle,
		ActorID:      &change.actorID,
		Project:      change.project,
		ProjectColor: change.projectColor,
	}
	if change.workspaceID != 0 {
		notification.WorkspaceID = &change.workspaceID
	}

	// CreateNotification publishes `notification.created` to the same user
	// topic, so the inbox badge updates off the same call.
	if notifyErr := s.CreateNotification(ctx, notification); notifyErr != nil {
		logger.Ctx(ctx).Error("announceAssigneeChange: failed to create the assignment notification",
			logger.Error(notifyErr))
	}
}

func (s *Service) notifyTaskRecipient(
	ctx context.Context,
	task domain.Task,
	recipientID, actorID int64,
	notificationType, title, description string,
	settingEnabled func(domain.UserSettings) bool,
) {
	if recipientID == 0 || recipientID == actorID {
		return
	}

	settings, err := s.repo.GetUserSettings(ctx, recipientID)
	if err != nil {
		logger.Ctx(ctx).Error("notifyTaskRecipient: failed to read notification settings",
			logger.Int("owner_id", int(task.OwnerID)),
			logger.Int("task_id", int(task.ID)),
			logger.Int("recipient_id", int(recipientID)),
			logger.Error(err))
		return
	}
	if !settingEnabled(settings) {
		return
	}

	taskID := task.ID
	actor := actorID
	notification := domain.CreateNotificationInput{
		UserID:       recipientID,
		Type:         notificationType,
		Title:        title,
		Description:  description,
		TaskID:       &taskID,
		TaskTitle:    task.Title,
		ActorID:      &actor,
		Project:      task.ProjectName,
		ProjectColor: task.ProjectColor,
	}
	if task.WorkspaceID != 0 {
		workspaceID := task.WorkspaceID
		notification.WorkspaceID = &workspaceID
	}

	if err := s.CreateNotification(ctx, notification); err != nil {
		logger.Ctx(ctx).Error("notifyTaskRecipient: failed to create notification",
			logger.Int("owner_id", int(task.OwnerID)),
			logger.Int("task_id", int(task.ID)),
			logger.Int("recipient_id", int(recipientID)),
			logger.String("type", notificationType),
			logger.Error(err))
	}
}

func mentionedUsernames(content string) []string {
	usernames := make([]string, 0)
	seen := make(map[string]struct{})
	for index := 0; index < len(content); index++ {
		if content[index] != '@' || (index > 0 && isMentionPrefixChar(content[index-1])) {
			continue
		}

		start := index + 1
		if start >= len(content) || !isMentionFirstChar(content[start]) {
			continue
		}
		end := start
		for end < len(content) && isMentionTokenChar(content[end]) {
			end++
		}
		if end == start {
			continue
		}

		username := strings.ToLower(content[start:end])
		if _, found := seen[username]; found {
			index = end - 1
			continue
		}
		seen[username] = struct{}{}
		usernames = append(usernames, username)
		index = end - 1
	}
	return usernames
}

func isMentionTokenChar(character byte) bool {
	return (character >= 'a' && character <= 'z') ||
		(character >= 'A' && character <= 'Z') ||
		(character >= '0' && character <= '9') ||
		character == '.' || character == '_' || character == '-'
}

func isMentionFirstChar(character byte) bool {
	return (character >= 'a' && character <= 'z') ||
		(character >= 'A' && character <= 'Z') ||
		(character >= '0' && character <= '9')
}

func isMentionPrefixChar(character byte) bool {
	return isMentionTokenChar(character) || character == '@'
}

func (s *Service) announceCommentNotifications(ctx context.Context, task domain.Task, actorID int64, content string) {
	s.notifyTaskRecipient(ctx, task, assigneeID(task.AssigneeID), actorID,
		commentNotification, commentNotificationTitle, content,
		func(settings domain.UserSettings) bool { return settings.TaskComments })

	for _, username := range mentionedUsernames(content) {
		recipientID, found, err := s.repo.LookupUserByUsername(ctx, username, task.BusinessID)
		if err != nil {
			logger.Ctx(ctx).Error("announceCommentNotifications: failed to resolve mention",
				logger.Int("owner_id", int(task.OwnerID)),
				logger.Int("task_id", int(task.ID)),
				logger.String("username", username),
				logger.Error(err))
			continue
		}
		if !found {
			continue
		}
		s.notifyTaskRecipient(ctx, task, recipientID, actorID,
			mentionNotification, mentionNotificationTitle, content,
			func(settings domain.UserSettings) bool { return settings.Mentions })
	}
}

func (s *Service) announceStatusChange(ctx context.Context, task domain.Task, actorID int64, status string) {
	s.notifyTaskRecipient(ctx, task, assigneeID(task.AssigneeID), actorID,
		statusNotification, statusNotificationTitle, status,
		func(settings domain.UserSettings) bool { return settings.StatusChanges })
}

// publishUserTaskEvent sends a task event to one person's own user topic.
//
// This is the only place a task event reaches a `user:<id>` topic, and it only
// ever addresses the two people a reassignment is about, never a third party's
// topic. Topic authorization is checked once at subscribe time rather than per
// event, and the `user` scope resolves to "the caller must be that user", so a
// self-scoped topic needs no membership reasoning of its own.
//
// userID 0 means the task was unassigned, or was assigned from nobody. There is
// no topic to publish to.
func publishUserTaskEvent(userID, taskID, actorID int64) {
	if userID == 0 {
		return
	}
	wsbus.Publish(wsbus.Event{
		Topic:    fmt.Sprintf("user:%d", userID),
		Type:     taskUpdatedEvent,
		EntityID: taskID,
		ActorID:  actorID,
	})
}

func (s *Service) UpdateTaskStatus(ctx context.Context, input domain.UpdateTaskStatusInput, userID int64) error {
	existing, err := s.repo.GetTaskByID(ctx, input.ID)
	if err != nil {
		return err
	}

	// The my-tasks list filters on status and hides done and cancelled work, so
	// a status change can add or remove the task from the assignee's list.
	defer invalidateUserScope(ctx, assigneeID(existing.AssigneeID))

	if existing.Status != input.Status {
		if actErr := s.repo.LogActivity(ctx, input.ID, userID, "status_changed", existing.Status, input.Status); actErr != nil {
			logger.Ctx(ctx).Error("UpdateTaskStatus: failed to log activity", logger.Error(actErr))
		}
	}

	if updateErr := s.repo.UpdateTaskStatus(ctx, input); updateErr != nil {
		return updateErr
	}

	if existing.Status != input.Status {
		notificationTask := existing
		notificationTask.Status = input.Status
		s.announceStatusChange(ctx, notificationTask, userID, input.Status)
	}
	return nil
}

func (s *Service) DeleteTask(ctx context.Context, taskID, businessID int64) error {
	defer invalidateWorkspacesScope(ctx, businessID)

	// Read the assignee before the row goes. Once the task is deleted there is
	// nothing left to say whose cached my-tasks pages still list it.
	existing, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		logger.Ctx(ctx).Error("DeleteTask: failed to read task before delete, my-tasks cache left to expire",
			logger.Error(err))
	} else {
		defer invalidateUserScope(ctx, assigneeID(existing.AssigneeID))
	}

	return s.repo.DeleteTask(ctx, taskID, businessID)
}

type myTasksPage struct {
	Tasks []domain.Task
	Total int64
}

func (s *Service) GetMyTasks(ctx context.Context, userID, page, limit int64) ([]domain.Task, int64, error) {
	offset := (page - 1) * limit
	result, err := cacher.Of[myTasksPage]().GetOrComputeSingle(ctx,
		// page/limit are part of the cache key, not just the fetch - otherwise
		// page 2 would return page 1's cached rows until the 2-minute TTL expires.
		cacheKeys.Scoped("user", userID, fmt.Sprintf("workspace_tasks:page=%d:limit=%d", page, limit)),
		func(ctx context.Context) (myTasksPage, error) {
			// workspaceID=0: this is the cross-workspace "My tasks" board, which
			// intentionally spans every workspace the user belongs to.
			tasks, total, fetchErr := s.repo.GetMyTasks(ctx, userID, 0, limit, offset)
			return myTasksPage{Tasks: tasks, Total: total}, fetchErr
		},
		cacher.WithExpiry(ctx, workspacesTTL),
	)
	if err != nil {
		return nil, 0, err
	}
	return result.Tasks, result.Total, nil
}

func (s *Service) SearchTasks(ctx context.Context, workspaceID int64, q string, limit, offset int64) ([]domain.Task, int64, error) {
	return s.repo.SearchTasks(ctx, workspaceID, q, limit, offset)
}

// ─────────────────────────────────────────────
// Comments
// ─────────────────────────────────────────────

func (s *Service) AddComment(ctx context.Context, taskID, userID int64, content string) (domain.TaskComment, error) {
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return domain.TaskComment{}, err
	}

	id, err := s.repo.CreateComment(ctx, taskID, userID, content)
	if err != nil {
		return domain.TaskComment{}, err
	}

	if actErr := s.repo.LogActivity(ctx, taskID, userID, "comment_added", "", content); actErr != nil {
		logger.Ctx(ctx).Error("AddComment: failed to log activity", logger.Error(actErr))
	}

	comment := domain.TaskComment{
		ID:      id,
		TaskID:  taskID,
		UserID:  userID,
		Content: content,
	}
	if stored, readErr := s.repo.GetCommentByID(ctx, id); readErr != nil {
		logger.Ctx(ctx).Error("AddComment: failed to read stored comment", logger.Error(readErr))
	} else {
		comment = stored
	}

	s.announceCommentNotifications(ctx, task, userID, content)
	return comment, nil
}

func (s *Service) DeleteComment(ctx context.Context, commentID, userID int64) error {
	return s.repo.DeleteComment(ctx, commentID, userID)
}

// GetCommentTaskID returns the task a comment belongs to. Handlers call
// this before DeleteComment so the realtime event can route to the
// right topic after the row is gone.
func (s *Service) GetCommentTaskID(ctx context.Context, commentID int64) int64 {
	taskID, err := s.repo.GetCommentTaskID(ctx, commentID)
	if err != nil {
		return 0
	}
	return taskID
}

// ─────────────────────────────────────────────
// Dashboard
// ─────────────────────────────────────────────

func (s *Service) GetDashboard(ctx context.Context, workspaceID, businessID, userID int64) (domain.DashboardStats, error) {
	stats, err := s.repo.GetDashboardStats(ctx, workspaceID, businessID)
	if err != nil {
		return domain.DashboardStats{}, err
	}

	// Scope "my tasks" to this workspace so the dashboard never surfaces work
	// assigned in a different workspace (issue #783).
	myTasks, _, err := s.repo.GetMyTasks(ctx, userID, workspaceID, internalListCap, 0)
	if err != nil {
		return domain.DashboardStats{}, err
	}

	stats.MyTasks = myTasks

	return stats, nil
}

// ─────────────────────────────────────────────
// Task Activity (standalone)
// ─────────────────────────────────────────────

func (s *Service) GetTaskActivity(ctx context.Context, taskID int64) ([]domain.TaskActivity, error) {
	return s.repo.GetActivityByTask(ctx, taskID, 100)
}

// ─────────────────────────────────────────────
// Task Subtasks (standalone)
// ─────────────────────────────────────────────

func (s *Service) GetSubtasks(ctx context.Context, taskID, limit, offset int64) ([]domain.Task, int64, error) {
	return s.repo.GetSubtasks(ctx, taskID, limit, offset)
}

func (s *Service) CreateSubtask(ctx context.Context, input domain.TaskInput) (domain.Task, error) {
	return s.CreateTask(ctx, input)
}

// ─────────────────────────────────────────────
// Task Attachments
// ─────────────────────────────────────────────

func (s *Service) GetAttachments(ctx context.Context, taskID int64) ([]domain.TaskAttachment, error) {
	return s.repo.GetAttachmentsByTask(ctx, taskID)
}

func (s *Service) CreateAttachment(ctx context.Context, input domain.TaskAttachmentInput) (int64, error) {
	return s.repo.CreateAttachment(ctx, input)
}

func (s *Service) DeleteAttachment(ctx context.Context, attachmentID, businessID int64) error {
	return s.repo.DeleteAttachment(ctx, attachmentID, businessID)
}

// ─────────────────────────────────────────────
// Time Logs
// ─────────────────────────────────────────────

func (s *Service) GetTimeLogs(ctx context.Context, taskID int64) ([]domain.TimeLog, error) {
	return s.repo.GetTimeLogsByTask(ctx, taskID)
}

func (s *Service) CreateTimeLog(ctx context.Context, input domain.TimeLogInput) (int64, error) {
	id, err := s.repo.CreateTimeLog(ctx, input)
	if err != nil {
		return 0, err
	}
	if recalcErr := s.repo.RecalcActualHours(ctx, input.TaskID, input.UserID); recalcErr != nil {
		logger.Ctx(ctx).Error("CreateTimeLog: failed to recalc actual_hours", logger.Error(recalcErr))
	}
	return id, nil
}

func (s *Service) DeleteTimeLog(ctx context.Context, id, userID, taskID int64) error {
	if err := s.repo.DeleteTimeLog(ctx, id, userID); err != nil {
		return err
	}
	if recalcErr := s.repo.RecalcActualHours(ctx, taskID, userID); recalcErr != nil {
		logger.Ctx(ctx).Error("DeleteTimeLog: failed to recalc actual_hours", logger.Error(recalcErr))
	}
	return nil
}

// DeleteTimeLogFlat deletes a time log by ID, looking up the task_id automatically.
func (s *Service) DeleteTimeLogFlat(ctx context.Context, id, userID int64) error {
	taskID, err := s.repo.GetTimeLogTaskID(ctx, id)
	if err != nil {
		return err
	}
	return s.DeleteTimeLog(ctx, id, userID, taskID)
}

// ─────────────────────────────────────────────
// Labels
// ─────────────────────────────────────────────

func (s *Service) ListLabels(ctx context.Context, workspaceID, limit, offset int64) ([]domain.Label, int64, error) {
	return s.repo.GetLabelsByWorkspace(ctx, workspaceID, limit, offset)
}

func (s *Service) CreateLabel(ctx context.Context, input domain.LabelInput) (domain.Label, error) {
	id, err := s.repo.CreateLabel(ctx, input)
	if err != nil {
		return domain.Label{}, err
	}

	return domain.Label{
		ID:          id,
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		Color:       input.Color,
	}, nil
}

func (s *Service) UpdateLabel(ctx context.Context, id, actorID, businessID int64, input domain.LabelInput) error {
	return s.repo.UpdateLabel(ctx, id, actorID, businessID, input)
}

func (s *Service) DeleteLabel(ctx context.Context, id, actorID, businessID int64) error {
	return s.repo.DeleteLabel(ctx, id, actorID, businessID)
}

func (s *Service) SetTaskLabels(ctx context.Context, taskID, actorID int64, labelIDs []int64) error {
	return s.repo.SetTaskLabels(ctx, taskID, actorID, labelIDs)
}

// ─────────────────────────────────────────────
// Cycles
// ─────────────────────────────────────────────

func (s *Service) ListCycles(ctx context.Context, workspaceID, limit, offset int64) ([]domain.Cycle, int64, error) {
	return s.repo.ListCycles(ctx, workspaceID, limit, offset)
}

func (s *Service) CreateCycle(ctx context.Context, input domain.CycleInput) (domain.Cycle, error) {
	id, err := s.repo.CreateCycle(ctx, input)
	if err != nil {
		return domain.Cycle{}, err
	}
	status := "upcoming"
	now := time.Now().UTC()
	if now.After(input.StartDate) && now.Before(input.EndDate) {
		status = "active"
	}
	return domain.Cycle{
		ID:          id,
		WorkspaceID: input.WorkspaceID,
		ProjectID:   input.ProjectID,
		Name:        input.Name,
		Description: input.Description,
		Status:      status,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
		CreatedBy:   input.CreatedBy,
	}, nil
}

func (s *Service) UpdateCycle(ctx context.Context, input domain.UpdateCycleInput) error {
	return s.repo.UpdateCycle(ctx, input)
}

// ─────────────────────────────────────────────
// Milestones
// ─────────────────────────────────────────────

func (s *Service) ListMilestones(ctx context.Context, workspaceID int64, projectID *int64, limit, offset int64) ([]domain.Milestone, int64, error) {
	return s.repo.ListMilestones(ctx, workspaceID, projectID, limit, offset)
}

func (s *Service) CreateMilestone(ctx context.Context, input domain.MilestoneInput) (domain.Milestone, error) {
	id, err := s.repo.CreateMilestone(ctx, input)
	if err != nil {
		return domain.Milestone{}, err
	}
	return domain.Milestone{
		ID:          id,
		ProjectID:   input.ProjectID,
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		Description: input.Description,
		Status:      "planning",
		DueDate:     input.DueDate,
		CreatedBy:   input.CreatedBy,
	}, nil
}

func (s *Service) UpdateMilestone(ctx context.Context, input domain.UpdateMilestoneInput) error {
	return s.repo.UpdateMilestone(ctx, input)
}

// ─────────────────────────────────────────────
// Views
// ─────────────────────────────────────────────

// ListViews returns the business's shared views plus the caller's own private
// ones. callerID is required rather than optional: passing 0 would hide every
// private view from its own creator.
func (s *Service) ListViews(ctx context.Context, businessID, workspaceID, callerID, limit, offset int64) ([]domain.View, int64, error) {
	return s.repo.ListViews(ctx, businessID, workspaceID, callerID, limit, offset)
}

func (s *Service) CreateView(ctx context.Context, input domain.ViewInput) (domain.View, error) {
	id, err := s.repo.CreateView(ctx, input)
	if err != nil {
		return domain.View{}, err
	}
	return domain.View{
		ID:          id,
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		Icon:        input.Icon,
		Filters:     input.Filters,
		GroupBy:     input.GroupBy,
		SortBy:      input.SortBy,
		SortDir:     input.SortDir,
		CreatedBy:   input.CreatedBy,
	}, nil
}

func (s *Service) UpdateView(ctx context.Context, input domain.UpdateViewInput) error {
	return s.repo.UpdateView(ctx, input)
}

func (s *Service) DeleteView(ctx context.Context, id, actorID, businessID int64) error {
	return s.repo.DeleteView(ctx, id, actorID, businessID)
}

// ─────────────────────────────────────────────
// Insights
// ─────────────────────────────────────────────

func (s *Service) GetInsights(ctx context.Context, workspaceID int64) (domain.Insights, error) {
	return s.repo.GetInsights(ctx, workspaceID)
}

// ─────────────────────────────────────────────
// Team Workload
// ─────────────────────────────────────────────

func (s *Service) GetTeamWorkload(ctx context.Context, workspaceID, limit, offset int64) ([]domain.TeamMemberWorkload, int64, error) {
	return s.repo.GetTeamWorkload(ctx, workspaceID, limit, offset)
}

// ─────────────────────────────────────────────
// Roadmap
// ─────────────────────────────────────────────

func (s *Service) GetRoadmap(ctx context.Context, workspaceID, limit, offset int64) ([]domain.Project, int64, error) {
	return s.repo.GetRoadmap(ctx, workspaceID, limit, offset)
}

// ─────────────────────────────────────────────
// Notifications
// ─────────────────────────────────────────────

func (s *Service) ListNotifications(ctx context.Context, userID, limit, offset int64) ([]domain.Notification, int64, error) {
	return s.repo.ListNotifications(ctx, userID, limit, offset)
}

func (s *Service) MarkNotificationRead(ctx context.Context, id, userID int64) error {
	return s.repo.MarkNotificationRead(ctx, id, userID)
}

func (s *Service) MarkAllNotificationsRead(ctx context.Context, userID int64) error {
	return s.repo.MarkAllNotificationsRead(ctx, userID)
}

func (s *Service) CreateNotification(ctx context.Context, input domain.CreateNotificationInput) error {
	if err := s.repo.CreateNotification(ctx, input); err != nil {
		return err
	}
	// Push a hint to the recipient's user topic so their inbox badge
	// and list update in real time. The entity id is the recipient
	// rather than the notification row — consumers invalidate their
	// `['notifications']` query rather than patching by id.
	if input.UserID != 0 {
		wsbus.Publish(wsbus.Event{
			Topic:    fmt.Sprintf("user:%d", input.UserID),
			Type:     "notification.created",
			EntityID: input.UserID,
		})
	}
	return nil
}

// ─────────────────────────────────────────────
// User Settings
// ─────────────────────────────────────────────

func (s *Service) GetUserSettings(ctx context.Context, userID int64) (domain.UserSettings, error) {
	return s.repo.GetUserSettings(ctx, userID)
}

func (s *Service) UpdateUserSettings(ctx context.Context, settings domain.UserSettings) error {
	return s.repo.UpsertUserSettings(ctx, settings)
}

// ─────────────────────────────────────────────
// Project task counts (for roadmap)
// ─────────────────────────────────────────────

func (s *Service) GetProjectTaskCounts(ctx context.Context, projectID int64) (total, done int) {
	return s.repo.GetProjectTaskCounts(ctx, projectID)
}

func (s *Service) GetProjectTaskCountsByProjectIDs(ctx context.Context, projectIDs []int64) (map[int64]domain.ProjectTaskCounts, error) {
	return s.repo.GetProjectTaskCountsByProjectIDs(ctx, projectIDs)
}

// ─────────────────────────────────────────────
// Project Columns
// ─────────────────────────────────────────────

func (s *Service) GetColumnsByProject(ctx context.Context, projectID, limit, offset int64) ([]domain.ProjectColumn, int64, error) {
	return s.repo.GetColumnsByProject(ctx, projectID, limit, offset)
}

func (s *Service) CreateColumn(ctx context.Context, input domain.ProjectColumnInput) (domain.ProjectColumn, error) {
	id, err := s.repo.CreateColumn(ctx, input)
	if err != nil {
		return domain.ProjectColumn{}, err
	}
	cols, _, err := s.repo.GetColumnsByProject(ctx, input.ProjectID, internalListCap, 0)
	if err != nil {
		return domain.ProjectColumn{}, err
	}
	for _, c := range cols {
		if c.ID == id {
			return c, nil
		}
	}
	return domain.ProjectColumn{
		ID: id, ProjectID: input.ProjectID, WorkspaceID: input.WorkspaceID,
		Name: input.Name, Color: input.Color, SortOrder: input.SortOrder,
	}, nil
}

func (s *Service) UpdateColumn(ctx context.Context, input domain.UpdateProjectColumnInput) error {
	return s.repo.UpdateColumn(ctx, input)
}

func (s *Service) DeleteColumn(ctx context.Context, id, actorID int64) error {
	return s.repo.DeleteColumn(ctx, id, actorID)
}

// MoveTaskToColumn is the board's drag-and-drop write path, reached from the
// flow project API. It reassigns the task to a different column, and the column
// a task sits in is part of every my-tasks payload, so the assignee's cached
// pages have to go with it.
func (s *Service) MoveTaskToColumn(ctx context.Context, taskID, columnID, actorID int64) error {
	existing, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		logger.Ctx(ctx).Error("MoveTaskToColumn: failed to read task, my-tasks cache left to expire",
			logger.Error(err))
	} else {
		defer invalidateUserScope(ctx, assigneeID(existing.AssigneeID))
	}

	return s.repo.MoveTaskToColumn(ctx, taskID, columnID, actorID)
}

// GetColumnProject returns the project id of the given column, or 0 if
// the lookup fails. Used by the realtime publish path that needs a topic
// without the caller supplying the project id.
func (s *Service) GetColumnProject(ctx context.Context, columnID int64) int64 {
	return s.repo.GetColumnProject(ctx, columnID)
}

// GetProjectWorkspace returns the workspace id the given project belongs
// to. Used by realtime publish sites that route events to workspace-level
// subscribers without having to load the full project row.
func (s *Service) GetProjectWorkspace(ctx context.Context, projectID int64) int64 {
	return s.repo.GetProjectWorkspace(ctx, projectID)
}

// EnsureMyTasksProject returns the id of the caller's personal "My Tasks"
// project (and the workspace it lives in), creating it — plus its default
// kanban columns — on first call. This is the board the user lands on in Flow
// and where quick-created tasks go. Idempotent: it reuses the existing project,
// migrating the legacy "Inbox" name to "My Tasks" on access. Pass workspaceID 0
// to use (and lazily create) the caller's personal workspace. businessID must
// be the caller's resolved tenant id (tenant.ResolveBusinessID), not their
// user id — this is what the created project's business_id column stores.
func (s *Service) EnsureMyTasksProject(
	ctx context.Context,
	userID, businessID, workspaceID int64,
) (projectID, resolvedWorkspaceID int64, err error) {
	if workspaceID == 0 {
		wsID, wsErr := s.repo.CreatePersonalWorkspace(ctx, userID, businessID)
		if wsErr != nil {
			return 0, 0, wsErr
		}
		workspaceID = wsID
	} else if !s.repo.IsWorkspaceMember(ctx, userID, workspaceID) {
		return 0, 0, fmt.Errorf("workspace access denied")
	}

	if existingID, existingName := s.repo.GetMyTasksProject(ctx, userID, workspaceID); existingID != 0 {
		if existingName == "Inbox" {
			if renameErr := s.repo.RenameProject(ctx, existingID, userID, workspaceID, "My Tasks"); renameErr != nil {
				logger.Ctx(ctx).Warn("migrate Inbox->My Tasks failed", logger.Error(renameErr))
			} else {
				invalidateWorkspacesScope(ctx, businessID)
			}
		}
		return existingID, workspaceID, nil
	}

	id, err := s.repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: workspaceID,
		OwnerID:     userID,
		BusinessID:  businessID,
		Name:        "My Tasks",
		Description: "Your personal task board. Quick-created tasks land here.",
		Icon:        "✅",
		Color:       "#64C5B9",
		Visibility:  domain.VisibilityPrivate,
		IsPersonal:  true,
		CreatedBy:   userID,
	})
	if err != nil {
		return 0, 0, err
	}
	// Seed default kanban columns so the Board view works out of the box.
	if seedErr := s.repo.SeedDefaultColumns(ctx, id, workspaceID, userID); seedErr != nil {
		// Non-fatal — user can still create columns manually.
		_ = seedErr
	}
	invalidateWorkspacesScope(ctx, businessID)
	return id, workspaceID, nil
}

func (s *Service) SeedDefaultColumns(ctx context.Context, projectID, workspaceID, createdBy int64) error {
	return s.repo.SeedDefaultColumns(ctx, projectID, workspaceID, createdBy)
}
