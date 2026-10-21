package application

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
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
	repo repository

	workdayRoster  WorkdayRosterSource
	workdayLines   WorkDayLineStore
	memberProfiles InsightsMemberProfileSource

	// labourCosts is optional because workers and tests can construct the
	// module without payroll. The no-op implementation keeps those consumers
	// safe while the API bootstrap wires the HR-owned provider.
	labourCosts labourcost.Provider

	// hubProfiles resolves central access for Flow reporting and business
	// administrators, while member-level resource access remains Flow-owned.
	hubProfiles hubprofile.Resolver

	// orgchart resolves who reports to a caller, backing the dept-manager
	// visibility proxy (M3). Optional in exactly the way hubProfiles is: nil
	// until SetOrgchartResolver is called, in which case the proxy is skipped
	// and visibility behaves as it did before it existed. A nil provider must
	// never widen access, only fail to widen it.
	orgchart orgchart.ManagerResolver
}

// SetWorkdayRosterSource wires the owner-owned HR and organization directory
// data used to resolve team work days.
func (s *Service) SetWorkdayRosterSource(source WorkdayRosterSource) {
	if source != nil {
		s.workdayRoster = source
	}
}

func (s *Service) SetWorkDayLineStore(store WorkDayLineStore) {
	if store != nil {
		s.workdayLines = store
	}
}

func (s *Service) SetInsightsMemberProfileSource(source InsightsMemberProfileSource) {
	if source != nil {
		s.memberProfiles = source
	}
}

func NewService(_ context.Context, repo repository) *Service {

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

type repository interface {
	AddProjectMember(ctx context.Context, input domain.AddProjectMemberInput) error
	AddTaskDependency(ctx context.Context, input domain.AddTaskDependencyInput) error
	AddWorkspaceMember(ctx context.Context, input domain.AddWorkspaceMemberInput) error
	ArchiveProject(ctx context.Context, id, businessID int64, purgeAt *time.Time) error
	CountWorkTypeChildren(ctx context.Context, ownerID, businessID, parentID int64) (int64, error)
	CountWorkspacesByOwner(ctx context.Context, businessID int64) (int64, error)
	CreateAttachment(ctx context.Context, input domain.TaskAttachmentInput) (int64, error)
	CreateColumn(ctx context.Context, input domain.ProjectColumnInput) (int64, error)
	CreateComment(ctx context.Context, taskID, userID int64, content string) (int64, error)
	CreateCycle(ctx context.Context, input domain.CycleInput) (int64, error)
	CreateCycleV2(ctx context.Context, input domain.CycleInput) (int64, error)
	CreateLabel(ctx context.Context, input domain.LabelInput) (int64, error)
	CreateLabelV2(ctx context.Context, businessID int64, name, color string) (int64, error)
	CreateMilestone(ctx context.Context, input domain.MilestoneInput) (int64, error)
	CreateMilestoneV2(ctx context.Context, input domain.MilestoneInput) (int64, error)
	CreateNotification(ctx context.Context, input domain.CreateNotificationInput) error
	CreatePersonalWorkspace(ctx context.Context, userID, businessID int64) (int64, error)
	CreateProject(ctx context.Context, input domain.ProjectInput) (int64, error)
	CreateStatus(ctx context.Context, businessID, createdBy int64, in domain.StatusInput) (domain.Status, error)
	CreateTask(ctx context.Context, input domain.TaskInput) (int64, error)
	CreateTimeLog(ctx context.Context, input domain.TimeLogInput) (int64, error)
	CreateView(ctx context.Context, input domain.ViewInput) (int64, error)
	CreateViewV2(ctx context.Context, input domain.ViewInput) (int64, error)
	CreateWorkDayLine(ctx context.Context, input domain.WorkDayLineInput) (domain.WorkDayLine, error)
	CreateWorkType(ctx context.Context, input domain.WorkTypeInput) (domain.WorkType, error)
	CreateWorkspace(ctx context.Context, input domain.WorkspaceInput, isDefault bool) (int64, error)
	DeleteAttachment(ctx context.Context, attachmentID, businessID int64) error
	DeleteColumn(ctx context.Context, id, actorID int64) error
	DeleteComment(ctx context.Context, commentID, userID int64) error
	DeleteLabel(ctx context.Context, id, actorID, businessID int64) error
	DeleteStatus(ctx context.Context, id, businessID int64) (int64, error)
	DeleteTask(ctx context.Context, taskID, businessID int64) error
	DeleteTimeLog(ctx context.Context, id, userID int64) error
	DeleteView(ctx context.Context, id, actorID, businessID int64) error
	DeleteWorkDayLine(ctx context.Context, id, ownerID, businessID int64) (bool, error)
	DeleteWorkType(ctx context.Context, id, ownerID, businessID int64) (bool, error)
	DeleteWorkspace(ctx context.Context, id, businessID int64) error
	GetActivityByTask(ctx context.Context, taskID int64, limit int) ([]domain.TaskActivity, error)
	GetAttachmentTaskID(ctx context.Context, attachmentID int64) (int64, error)
	GetAttachmentsByTask(ctx context.Context, taskID int64) ([]domain.TaskAttachment, error)
	GetColumnProject(ctx context.Context, columnID int64) int64
	GetColumnsByProject(ctx context.Context, projectID, limit, offset int64) ([]domain.ProjectColumn, int64, error)
	GetCommentByID(ctx context.Context, commentID int64) (domain.TaskComment, error)
	GetCommentTaskID(ctx context.Context, commentID int64) (int64, error)
	GetCommentsByTask(ctx context.Context, taskID int64) ([]domain.TaskComment, error)
	GetDashboardStats(ctx context.Context, workspaceID, businessID int64) (domain.DashboardStats, error)
	GetDashboardStatsV2(ctx context.Context, businessID int64) (domain.DashboardStats, error)
	GetFocusPins(ctx context.Context, userID, businessID int64, taskIDs []int64) (map[int64]bool, error)
	GetInsights(ctx context.Context, workspaceID int64) (domain.Insights, error)
	ListInsightsTaskIDs(ctx context.Context, workspaceID, afterID int64, limit int32) ([]int64, error)
	GetLabelsByWorkspace(ctx context.Context, workspaceID, limit, offset int64) ([]domain.Label, int64, error)
	GetMyTasks(ctx context.Context, userID, workspaceID, limit, offset int64) ([]domain.Task, int64, error)
	GetMyTasksForBusiness(ctx context.Context, userID, businessID, workspaceID, limit, offset int64) ([]domain.Task, int64, error)
	GetMyTasksProject(ctx context.Context, userID, workspaceID int64) (int64, string)
	GetProjectMemberRole(ctx context.Context, projectID, userID int64) (string, error)
	GetProjectOwnerWorkspace(ctx context.Context, projectID int64) (int64, int64, error)
	GetProjectProxyFacts(ctx context.Context, projectID int64) (domain.Project, error)
	GetProjectTaskCounts(ctx context.Context, projectID int64) (total, done int)
	GetProjectTaskCountsByProjectIDs(ctx context.Context, projectIDs []int64) (map[int64]domain.ProjectTaskCounts, error)
	GetProjectV2ByID(ctx context.Context, id, businessID int64) (domain.Project, error)
	GetProjectWorkspace(ctx context.Context, projectID int64) int64
	GetProjectsByCreatorOrMembership(ctx context.Context, businessID, callerID, limit, offset int64) ([]domain.Project, int64, error)
	GetProjectsByWorkspace(ctx context.Context, workspaceID, businessID, limit, offset int64) ([]domain.Project, int64, error)
	GetRoadmap(ctx context.Context, workspaceID, limit, offset int64) ([]domain.Project, int64, error)
	GetStatusByID(ctx context.Context, id, businessID int64) (domain.Status, error)
	GetSubtasks(ctx context.Context, parentTaskID, limit, offset int64) ([]domain.Task, int64, error)
	GetTaskByID(ctx context.Context, taskID int64) (domain.Task, error)
	GetTasksByBusinessV2(ctx context.Context, filter domain.ListTasksFilter, callerID int64, includeAll bool) ([]domain.Task, int64, error)
	GetTasksByProject(ctx context.Context, filter domain.ListTasksFilter) ([]domain.Task, int64, error)
	GetTasksByWorkspace(ctx context.Context, filter domain.ListTasksFilter) ([]domain.Task, int64, error)
	GetTeamWorkload(ctx context.Context, workspaceID, limit, offset int64) ([]domain.TeamMemberWorkload, int64, error)
	GetTimeLogTaskID(ctx context.Context, id int64) (int64, error)
	GetTimeLogsByTask(ctx context.Context, taskID int64) ([]domain.TimeLog, error)
	GetUserSettings(ctx context.Context, userID int64) (domain.UserSettings, error)
	GetUserWorkspaceRole(ctx context.Context, workspaceID, userID int64) (string, error)
	GetWorkDay(ctx context.Context, ownerID, businessID, employeeID int64, date string) (domain.WorkDay, bool, error)
	GetWorkDayByID(ctx context.Context, id, businessID int64) (domain.WorkDay, bool, error)
	GetWorkDayLineOwner(ctx context.Context, id, ownerID, businessID int64) (domain.WorkDayLineOwner, bool, error)
	GetWorkType(ctx context.Context, id, ownerID, businessID int64) (domain.WorkType, error)
	GetWorkspaceOwner(ctx context.Context, workspaceID int64) (int64, error)
	GetWorkspacesByOwner(ctx context.Context, businessID, limit, offset int64) ([]domain.Workspace, int64, error)
	IsProjectMember(ctx context.Context, userID, projectID int64) bool
	IsWorkspaceMember(ctx context.Context, userID, workspaceID int64) bool
	ListArchivedProjectsV2(ctx context.Context, businessID, callerID int64, includeAll bool, limit, offset int64) ([]domain.Project, int64, error)
	ListCycles(ctx context.Context, workspaceID, limit, offset int64) ([]domain.Cycle, int64, error)
	ListCyclesV2(ctx context.Context, projectID, businessID, callerID int64, includeAll bool, limit, offset int64) ([]domain.Cycle, int64, error)
	ListLabelsV2(ctx context.Context, businessID, limit, offset int64) ([]domain.Label, int64, error)
	ListMilestones(ctx context.Context, workspaceID int64, projectID *int64, limit, offset int64) ([]domain.Milestone, int64, error)
	ListMilestonesV2(ctx context.Context, businessID int64, projectID *int64, callerID int64, includeAll bool, limit, offset int64) ([]domain.Milestone, int64, error)
	ListNotifications(ctx context.Context, userID, limit, offset int64) ([]domain.Notification, int64, error)
	ListProjectMembers(ctx context.Context, projectID, limit, offset int64) ([]domain.ProjectMember, int64, error)
	ListProjectsV2(ctx context.Context, businessID, callerID int64, includeAll bool, managedUserIDs []int64, status, visibility, q string, limit, offset int64) ([]domain.Project, int64, error)
	ListPublicProjectsByOwnerV2(
		ctx context.Context,
		businessID, ownerID int64,
		limit int32,
	) ([]domain.ProfileProject, error)
	ListStatusesByProject(ctx context.Context, businessID, projectID int64) ([]domain.Status, error)
	ListTaskDependencies(ctx context.Context, taskIDs []int64, businessID int64) (map[int64][]int64, error)
	ListTeamWorkDays(ctx context.Context, businessID int64, employeeIDs, hiddenIDs, userIDs []int64, from, to string, limit int32, offset int64) ([]domain.WorkDay, int64, error)
	ListViews(ctx context.Context, businessID, workspaceID, callerID, limit, offset int64) ([]domain.View, int64, error)
	ListViewsV2(ctx context.Context, businessID, callerID, limit, offset int64) ([]domain.View, int64, error)
	ListWorkDayLines(ctx context.Context, businessID, workDayID int64) ([]domain.WorkDayLine, int32, error)
	ListWorkDays(ctx context.Context, ownerID, businessID, employeeID int64, from, to string, limit int32, offset int64) ([]domain.WorkDay, int64, error)
	ListWorkDaysByEmployeeIDs(ctx context.Context, ownerID, businessID int64, employeeIDs []int64, from, to string, limit int32, offset int64) ([]domain.WorkDay, int64, error)
	ListWorkTypes(ctx context.Context, ownerID, businessID int64, includeArchived bool) ([]domain.WorkType, error)
	ListWorkspaceMembers(ctx context.Context, workspaceID, limit, offset int64) ([]domain.WorkspaceMember, int64, error)
	LogActivity(ctx context.Context, taskID, userID int64, action, oldValue, newValue string) error
	LookupUserByEmail(ctx context.Context, email string) (int64, bool, error)
	LookupUserByUsername(ctx context.Context, username string, businessID int64) (int64, bool, error)
	MarkAllNotificationsRead(ctx context.Context, userID int64) error
	MarkNotificationRead(ctx context.Context, id, userID int64) error
	MoveProject(ctx context.Context, input domain.MoveProjectInput) error
	MoveTaskToColumn(ctx context.Context, taskID, columnID, actorID int64) error
	MoveTaskToColumnOrdered(ctx context.Context, taskID, columnID, actorID int64, orderedTaskIDs []int64) error
	PurgeProjectCascade(ctx context.Context, id, businessID int64) ([]string, error)
	RecalcActualHours(ctx context.Context, taskID, actorID int64) error
	RemoveProjectMember(ctx context.Context, input domain.RemoveProjectMemberInput) error
	RemoveTaskDependency(ctx context.Context, input domain.RemoveTaskDependencyInput) error
	RemoveWorkspaceMember(ctx context.Context, input domain.RemoveWorkspaceMemberInput) error
	RenameProject(ctx context.Context, projectID, ownerID, workspaceID int64, name string) error
	RestoreProjectV2(ctx context.Context, id, businessID int64) error
	SearchTasks(ctx context.Context, workspaceID int64, q string, limit, offset int64) ([]domain.Task, int64, error)
	SearchTasksV2(ctx context.Context, businessID, callerID int64, includeAll bool, q string, limit, offset int64) ([]domain.Task, int64, error)
	SeedDefaultColumns(ctx context.Context, projectID, workspaceID, createdBy int64) error
	SetTaskFocusPin(ctx context.Context, taskID, userID, businessID int64, pinned bool) error
	SetTaskLabels(ctx context.Context, taskID, actorID int64, labelIDs []int64) error
	SetTaskStatusID(ctx context.Context, taskID, businessID, statusID int64, legacyStatus string) error
	SumWorkDayLineMinutesByDayIDs(ctx context.Context, businessID int64, workDayIDs []int64) (map[int64]int32, error)
	TransferProjectOwnership(ctx context.Context, input domain.TransferProjectOwnershipInput) error
	UpdateColumn(ctx context.Context, input domain.UpdateProjectColumnInput) error
	UpdateCycle(ctx context.Context, input domain.UpdateCycleInput) error
	UpdateLabel(ctx context.Context, id, actorID, businessID int64, input domain.LabelInput) error
	UpdateMilestone(ctx context.Context, input domain.UpdateMilestoneInput) error
	UpdateMilestoneV2(ctx context.Context, input domain.UpdateMilestoneInput) error
	UpdateProject(ctx context.Context, input domain.UpdateProjectInput) error
	UpdateProjectMemberRole(ctx context.Context, input domain.UpdateProjectMemberRoleInput) error
	UpdateProjectVisibility(ctx context.Context, id, businessID int64, visibility string) error
	UpdateStatus(ctx context.Context, id, businessID, projectID int64, in domain.StatusInput) (domain.Status, error)
	UpdateTask(ctx context.Context, input domain.UpdateTaskInput) error
	UpdateTaskStatus(ctx context.Context, input domain.UpdateTaskStatusInput) error
	UpdateView(ctx context.Context, input domain.UpdateViewInput) error
	UpdateWorkDayLine(ctx context.Context, input domain.WorkDayLineInput) (domain.WorkDayLine, error)
	UpdateWorkType(ctx context.Context, input domain.WorkTypeInput) (domain.WorkType, error)
	UpdateWorkspace(ctx context.Context, input domain.UpdateWorkspaceInput) error
	UpdateWorkspaceMemberRole(ctx context.Context, input domain.UpdateWorkspaceMemberRoleInput) error
	UpsertUserSettings(ctx context.Context, s domain.UserSettings) error
	UpsertWorkDay(ctx context.Context, input domain.WorkDayInput) (domain.WorkDay, error)
}
