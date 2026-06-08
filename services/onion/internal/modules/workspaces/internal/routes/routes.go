package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, flowService *application.Service) {
	// ─── Workspaces ───────────────────────────────────────────────────────────

	http.RegisterGETWithValidation[handlers.ListWorkspacesRequest, handlers.ListWorkspacesResponse](
		ctx, router, handlers.ListWorkspacesRoutes(),
		&handlers.ListWorkspacesHandler{FlowService: flowService},
	)

	http.RegisterPOSTWithValidation[handlers.CreateWorkspaceRequest, handlers.CreateWorkspaceResponse](
		ctx, router, handlers.CreateWorkspaceRoutes(),
		&handlers.CreateWorkspaceHandler{FlowService: flowService},
	)

	http.RegisterPUTWithValidation[handlers.UpdateWorkspaceRequest, handlers.UpdateWorkspaceResponse](
		ctx, router, handlers.UpdateWorkspaceRoutes(),
		&handlers.UpdateWorkspaceHandler{FlowService: flowService},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteWorkspaceRequest, handlers.DeleteWorkspaceResponse](
		ctx, router, handlers.DeleteWorkspaceRoutes(),
		&handlers.DeleteWorkspaceHandler{FlowService: flowService},
	)

	http.RegisterPOSTWithValidation[handlers.AddWorkspaceMemberRequest, handlers.AddWorkspaceMemberResponse](
		ctx, router, handlers.AddWorkspaceMemberRoutes(),
		&handlers.AddWorkspaceMemberHandler{FlowService: flowService},
	)

	http.RegisterDELETEWithValidation[handlers.RemoveWorkspaceMemberRequest, handlers.RemoveWorkspaceMemberResponse](
		ctx, router, handlers.RemoveWorkspaceMemberRoutes(),
		&handlers.RemoveWorkspaceMemberHandler{FlowService: flowService},
	)

	// ─── Projects ─────────────────────────────────────────────────────────────

	http.RegisterGETWithValidation[handlers.ListProjectsRequest, handlers.ListProjectsResponse](
		ctx, router, handlers.ListProjectsRoutes(),
		&handlers.ListProjectsHandler{FlowService: flowService},
	)

	http.RegisterPOSTWithValidation[handlers.CreateProjectRequest, handlers.CreateProjectResponse](
		ctx, router, handlers.CreateProjectRoutes(),
		&handlers.CreateProjectHandler{FlowService: flowService},
	)

	http.RegisterPUTWithValidation[handlers.UpdateProjectRequest, handlers.UpdateProjectResponse](
		ctx, router, handlers.UpdateProjectRoutes(),
		&handlers.UpdateProjectHandler{FlowService: flowService},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteProjectRequest, handlers.DeleteProjectResponse](
		ctx, router, handlers.DeleteProjectRoutes(),
		&handlers.DeleteProjectHandler{FlowService: flowService},
	)

	// ─── Tasks ────────────────────────────────────────────────────────────────

	http.RegisterGETWithValidation[handlers.ListTasksRequest, handlers.ListTasksResponse](
		ctx, router, handlers.ListTasksRoutes(),
		&handlers.ListTasksHandler{FlowService: flowService},
	)

	http.RegisterPOSTWithValidation[handlers.CreateTaskRequest, handlers.CreateTaskResponse](
		ctx, router, handlers.CreateTaskRoutes(),
		&handlers.CreateTaskHandler{FlowService: flowService},
	)

	http.RegisterGETWithValidation[handlers.GetTaskDetailRequest, handlers.GetTaskDetailResponse](
		ctx, router, handlers.GetTaskDetailRoutes(),
		&handlers.GetTaskDetailHandler{FlowService: flowService},
	)

	http.RegisterPUTWithValidation[handlers.UpdateTaskRequest, handlers.UpdateTaskResponse](
		ctx, router, handlers.UpdateTaskRoutes(),
		&handlers.UpdateTaskHandler{FlowService: flowService},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteTaskRequest, handlers.DeleteTaskResponse](
		ctx, router, handlers.DeleteTaskRoutes(),
		&handlers.DeleteTaskHandler{FlowService: flowService},
	)

	http.RegisterPUTWithValidation[handlers.UpdateTaskStatusRequest, handlers.UpdateTaskStatusResponse](
		ctx, router, handlers.UpdateTaskStatusRoutes(),
		&handlers.UpdateTaskStatusHandler{FlowService: flowService},
	)

	http.RegisterGET[handlers.GetMyTasksRequest, handlers.GetMyTasksResponse](
		ctx, router, handlers.GetMyTasksRoutes(),
		&handlers.GetMyTasksHandler{FlowService: flowService},
	)

	http.RegisterGET[handlers.GetMyTasksProjectRequest, handlers.GetMyTasksProjectResponse](
		ctx, router, handlers.GetMyTasksProjectRoutes(),
		&handlers.GetMyTasksProjectHandler{FlowService: flowService},
	)

	http.RegisterGETWithValidation[handlers.SearchTasksRequest, handlers.SearchTasksResponse](
		ctx, router, handlers.SearchTasksRoutes(),
		&handlers.SearchTasksHandler{FlowService: flowService},
	)

	// ─── Comments ─────────────────────────────────────────────────────────────

	http.RegisterPOSTWithValidation[handlers.AddCommentRequest, handlers.AddCommentResponse](
		ctx, router, handlers.AddCommentRoutes(),
		&handlers.AddCommentHandler{FlowService: flowService},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteCommentRequest, handlers.DeleteCommentResponse](
		ctx, router, handlers.DeleteCommentRoutes(),
		&handlers.DeleteCommentHandler{FlowService: flowService},
	)

	// ─── Task Activity ────────────────────────────────────────────────────────

	http.RegisterGET[handlers.GetTaskActivityRequest, handlers.GetTaskActivityResponse](
		ctx, router, handlers.GetTaskActivityRoutes(),
		&handlers.GetTaskActivityHandler{FlowService: flowService},
	)

	http.RegisterGETWithValidation[handlers.GetTaskActivityByQPRequest, handlers.GetTaskActivityByQPResponse](
		ctx, router, handlers.GetTaskActivityByQPRoutes(),
		&handlers.GetTaskActivityByQPHandler{FlowService: flowService},
	)

	// ─── Task Subtasks ───────────────────────────────────────────────────────

	http.RegisterGET[handlers.GetSubtasksRequest, handlers.GetSubtasksResponse](
		ctx, router, handlers.GetSubtasksRoutes(),
		&handlers.GetSubtasksHandler{FlowService: flowService},
	)

	http.RegisterPOSTWithValidation[handlers.CreateSubtaskRequest, handlers.CreateSubtaskResponse](
		ctx, router, handlers.CreateSubtaskRoutes(),
		&handlers.CreateSubtaskHandler{FlowService: flowService},
	)

	// ─── Task Attachments ────────────────────────────────────────────────────

	http.RegisterGET[handlers.GetAttachmentsRequest, handlers.GetAttachmentsResponse](
		ctx, router, handlers.GetAttachmentsRoutes(),
		&handlers.GetAttachmentsHandler{FlowService: flowService},
	)

	http.RegisterPOSTWithValidation[handlers.CreateAttachmentRequest, handlers.CreateAttachmentResponse](
		ctx, router, handlers.CreateAttachmentRoutes(),
		&handlers.CreateAttachmentHandler{FlowService: flowService},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteAttachmentRequest, handlers.DeleteAttachmentResponse](
		ctx, router, handlers.DeleteAttachmentRoutes(),
		&handlers.DeleteAttachmentHandler{FlowService: flowService},
	)

	// ─── Time Logs ───────────────────────────────────────────────────────────

	http.RegisterGET[handlers.GetTimeLogsRequest, handlers.GetTimeLogsResponse](
		ctx, router, handlers.GetTimeLogsRoutes(),
		&handlers.GetTimeLogsHandler{FlowService: flowService},
	)

	http.RegisterPOSTWithValidation[handlers.CreateTimeLogRequest, handlers.CreateTimeLogResponse](
		ctx, router, handlers.CreateTimeLogRoutes(),
		&handlers.CreateTimeLogHandler{FlowService: flowService},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteTimeLogRequest, handlers.DeleteTimeLogResponse](
		ctx, router, handlers.DeleteTimeLogRoutes(),
		&handlers.DeleteTimeLogHandler{FlowService: flowService},
	)

	http.RegisterGETWithValidation[handlers.GetTimeLogsFlatRequest, handlers.GetTimeLogsFlatResponse](
		ctx, router, handlers.GetTimeLogsFlatRoutes(),
		&handlers.GetTimeLogsFlatHandler{FlowService: flowService},
	)

	http.RegisterPOSTWithValidation[handlers.CreateTimeLogFlatRequest, handlers.CreateTimeLogFlatResponse](
		ctx, router, handlers.CreateTimeLogFlatRoutes(),
		&handlers.CreateTimeLogFlatHandler{FlowService: flowService},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteTimeLogFlatRequest, handlers.DeleteTimeLogFlatResponse](
		ctx, router, handlers.DeleteTimeLogFlatRoutes(),
		&handlers.DeleteTimeLogFlatHandler{FlowService: flowService},
	)

	// ─── Dashboard ────────────────────────────────────────────────────────────

	http.RegisterGETWithValidation[handlers.GetDashboardRequest, handlers.GetDashboardResponse](
		ctx, router, handlers.GetDashboardRoutes(),
		&handlers.GetDashboardHandler{FlowService: flowService},
	)

	// ─── Labels ───────────────────────────────────────────────────────────────

	http.RegisterGETWithValidation[handlers.ListLabelsRequest, handlers.ListLabelsResponse](
		ctx, router, handlers.ListLabelsRoutes(),
		&handlers.ListLabelsHandler{FlowService: flowService},
	)

	http.RegisterPOSTWithValidation[handlers.CreateLabelRequest, handlers.CreateLabelResponse](
		ctx, router, handlers.CreateLabelRoutes(),
		&handlers.CreateLabelHandler{FlowService: flowService},
	)

	http.RegisterPUTWithValidation[handlers.UpdateLabelRequest, handlers.UpdateLabelResponse](
		ctx, router, handlers.UpdateLabelRoutes(),
		&handlers.UpdateLabelHandler{FlowService: flowService},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteLabelRequest, handlers.DeleteLabelResponse](
		ctx, router, handlers.DeleteLabelRoutes(),
		&handlers.DeleteLabelHandler{FlowService: flowService},
	)

	http.RegisterPUTWithValidation[handlers.SetTaskLabelsRequest, handlers.SetTaskLabelsResponse](
		ctx, router, handlers.SetTaskLabelsRoutes(),
		&handlers.SetTaskLabelsHandler{FlowService: flowService},
	)

	// ─── Workspace Members (extended) ─────────────────────────────────────────

	http.RegisterGETWithValidation[handlers.ListWorkspaceMembersRequest, handlers.ListWorkspaceMembersResponse](
		ctx, router, handlers.ListWorkspaceMembersRoutes(),
		&handlers.ListWorkspaceMembersHandler{FlowService: flowService},
	)

	http.RegisterPUTWithValidation[handlers.UpdateWorkspaceMemberRoleRequest, handlers.UpdateWorkspaceMemberRoleResponse](
		ctx, router, handlers.UpdateWorkspaceMemberRoleRoutes(),
		&handlers.UpdateWorkspaceMemberRoleHandler{FlowService: flowService},
	)

	// ─── Project Members ──────────────────────────────────────────────────────

	http.RegisterGETWithValidation[handlers.ListProjectMembersRequest, handlers.ListProjectMembersResponse](
		ctx, router, handlers.ListProjectMembersRoutes(),
		&handlers.ListProjectMembersHandler{FlowService: flowService},
	)

	http.RegisterPOSTWithValidation[handlers.AddProjectMemberRequest, handlers.AddProjectMemberResponse](
		ctx, router, handlers.AddProjectMemberRoutes(),
		&handlers.AddProjectMemberHandler{FlowService: flowService},
	)

	http.RegisterDELETEWithValidation[handlers.RemoveProjectMemberRequest, handlers.RemoveProjectMemberResponse](
		ctx, router, handlers.RemoveProjectMemberRoutes(),
		&handlers.RemoveProjectMemberHandler{FlowService: flowService},
	)

	http.RegisterPUTWithValidation[handlers.UpdateProjectMemberRoleRequest, handlers.UpdateProjectMemberRoleResponse](
		ctx, router, handlers.UpdateProjectMemberRoleRoutes(),
		&handlers.UpdateProjectMemberRoleHandler{FlowService: flowService},
	)

	http.RegisterPUTWithValidation[handlers.TransferProjectOwnershipRequest, handlers.TransferProjectOwnershipResponse](
		ctx, router, handlers.TransferProjectOwnershipRoutes(),
		&handlers.TransferProjectOwnershipHandler{FlowService: flowService},
	)

	http.RegisterPUTWithValidation[handlers.MoveProjectRequest, handlers.MoveProjectResponse](
		ctx, router, handlers.MoveProjectRoutes(),
		&handlers.MoveProjectHandler{FlowService: flowService},
	)

	http.RegisterPOSTWithValidation[handlers.InviteProjectMemberByEmailRequest, handlers.InviteProjectMemberByEmailResponse](
		ctx, router, handlers.InviteProjectMemberByEmailRoutes(),
		&handlers.InviteProjectMemberByEmailHandler{FlowService: flowService},
	)

	// ─── Project Columns ──────────────────────────────────────────────────────────

	http.RegisterGETWithValidation[handlers.ListColumnsRequest, handlers.ListColumnsResponse](
		ctx, router, handlers.ListColumnsRoutes(),
		&handlers.ListColumnsHandler{FlowService: flowService},
	)

	http.RegisterPOSTWithValidation[handlers.CreateColumnRequest, handlers.CreateColumnResponse](
		ctx, router, handlers.CreateColumnRoutes(),
		&handlers.CreateColumnHandler{FlowService: flowService},
	)

	http.RegisterPUTWithValidation[handlers.UpdateColumnRequest, handlers.UpdateColumnResponse](
		ctx, router, handlers.UpdateColumnRoutes(),
		&handlers.UpdateColumnHandler{FlowService: flowService},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteColumnRequest, handlers.DeleteColumnResponse](
		ctx, router, handlers.DeleteColumnRoutes(),
		&handlers.DeleteColumnHandler{FlowService: flowService},
	)

	http.RegisterPUTWithValidation[handlers.MoveTaskToColumnRequest, handlers.MoveTaskToColumnResponse](
		ctx, router, handlers.MoveTaskToColumnRoutes(),
		&handlers.MoveTaskToColumnHandler{FlowService: flowService},
	)

	// ─── Cycles ───────────────────────────────────────────────────────────────────

	http.RegisterGETWithValidation[handlers.ListCyclesRequest, handlers.ListCyclesResponse](
		ctx, router, handlers.ListCyclesRoutes(),
		&handlers.ListCyclesHandler{FlowService: flowService},
	)

	http.RegisterPOSTWithValidation[handlers.CreateCycleRequest, handlers.CreateCycleResponse](
		ctx, router, handlers.CreateCycleRoutes(),
		&handlers.CreateCycleHandler{FlowService: flowService},
	)

	http.RegisterPUTWithValidation[handlers.UpdateCycleRequest, handlers.UpdateCycleResponse](
		ctx, router, handlers.UpdateCycleRoutes(),
		&handlers.UpdateCycleHandler{FlowService: flowService},
	)

	// ─── Milestones ───────────────────────────────────────────────────────────────

	http.RegisterGETWithValidation[handlers.ListMilestonesRequest, handlers.ListMilestonesResponse](
		ctx, router, handlers.ListMilestonesRoutes(),
		&handlers.ListMilestonesHandler{FlowService: flowService},
	)

	http.RegisterPOSTWithValidation[handlers.CreateMilestoneRequest, handlers.CreateMilestoneResponse](
		ctx, router, handlers.CreateMilestoneRoutes(),
		&handlers.CreateMilestoneHandler{FlowService: flowService},
	)

	http.RegisterPUTWithValidation[handlers.UpdateMilestoneRequest, handlers.UpdateMilestoneResponse](
		ctx, router, handlers.UpdateMilestoneRoutes(),
		&handlers.UpdateMilestoneHandler{FlowService: flowService},
	)

	// ─── Views ────────────────────────────────────────────────────────────────────

	http.RegisterGETWithValidation[handlers.ListViewsRequest, handlers.ListViewsResponse](
		ctx, router, handlers.ListViewsRoutes(),
		&handlers.ListViewsHandler{FlowService: flowService},
	)

	http.RegisterPOSTWithValidation[handlers.CreateViewRequest, handlers.CreateViewResponse](
		ctx, router, handlers.CreateViewRoutes(),
		&handlers.CreateViewHandler{FlowService: flowService},
	)

	http.RegisterPUTWithValidation[handlers.UpdateViewRequest, handlers.UpdateViewResponse](
		ctx, router, handlers.UpdateViewRoutes(),
		&handlers.UpdateViewHandler{FlowService: flowService},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteViewRequest, handlers.DeleteViewResponse](
		ctx, router, handlers.DeleteViewRoutes(),
		&handlers.DeleteViewHandler{FlowService: flowService},
	)

	// ─── Insights ─────────────────────────────────────────────────────────────────

	http.RegisterGETWithValidation[handlers.GetInsightsRequest, handlers.GetInsightsResponse](
		ctx, router, handlers.GetInsightsRoutes(),
		&handlers.GetInsightsHandler{FlowService: flowService},
	)

	http.RegisterGETWithValidation[handlers.GetTeamWorkloadRequest, handlers.GetTeamWorkloadResponse](
		ctx, router, handlers.GetTeamWorkloadRoutes(),
		&handlers.GetTeamWorkloadHandler{FlowService: flowService},
	)

	http.RegisterGETWithValidation[handlers.GetRoadmapRequest, handlers.GetRoadmapResponse](
		ctx, router, handlers.GetRoadmapRoutes(),
		&handlers.GetRoadmapHandler{FlowService: flowService},
	)

	// ─── Notifications ────────────────────────────────────────────────────────────

	http.RegisterGET[handlers.ListNotificationsRequest, handlers.ListNotificationsResponse](
		ctx, router, handlers.ListNotificationsRoutes(),
		&handlers.ListNotificationsHandler{FlowService: flowService},
	)

	http.RegisterPUTWithValidation[handlers.MarkNotificationReadRequest, handlers.MarkNotificationReadResponse](
		ctx, router, handlers.MarkNotificationReadRoutes(),
		&handlers.MarkNotificationReadHandler{FlowService: flowService},
	)

	http.RegisterPUTWithValidation[handlers.MarkAllNotificationsReadRequest, handlers.MarkAllNotificationsReadResponse](
		ctx, router, handlers.MarkAllNotificationsReadRoutes(),
		&handlers.MarkAllNotificationsReadHandler{FlowService: flowService},
	)

	// ─── Settings ─────────────────────────────────────────────────────────────────

	http.RegisterGET[handlers.GetUserSettingsRequest, handlers.GetUserSettingsResponse](
		ctx, router, handlers.GetUserSettingsRoutes(),
		&handlers.GetUserSettingsHandler{FlowService: flowService},
	)

	http.RegisterPUTWithValidation[handlers.UpdateUserSettingsRequest, handlers.UpdateUserSettingsResponse](
		ctx, router, handlers.UpdateUserSettingsRoutes(),
		&handlers.UpdateUserSettingsHandler{FlowService: flowService},
	)
}
