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
}
