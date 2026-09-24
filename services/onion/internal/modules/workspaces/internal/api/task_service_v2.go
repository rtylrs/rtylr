package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	flowtaskv2 "github.com/voxire/voxie/gen/go/flow/task/v2"
)

// TaskAPIV2 serves the workspace-free flow.task.v2 contracts.
//
// Deviations from the v1 handlers, each traceable to a spec ruling:
//   - createTask treats an omitted project_id as "file into my My Tasks
//     bucket" (R13), and stamps assignee membership (T1) so nobody is handed
//     work they cannot see.
//   - updateTask grants membership on reassignment the same way.
//   - list/search are business-wide and visibility-gated row by row (S5).
type TaskAPIV2 struct {
	Service *application.Service
}

func NewTaskAPIV2(service *application.Service) *TaskAPIV2 {
	return &TaskAPIV2{Service: service}
}

func (a *TaskAPIV2) resolveAuth(ctx context.Context) (*authPkg.AuthUser, int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, 0, onekiterrors.PermissionDenied()
	}
	businessID, err := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if err != nil {
		return nil, 0, onekiterrors.PermissionDenied()
	}
	return authUser, businessID, nil
}

func toCommentV2(c domain.TaskComment) *flowtaskv2.TaskComment {
	return &flowtaskv2.TaskComment{
		Id: c.ID, TaskId: c.TaskID, UserId: c.UserID,
		ActorName: c.ActorName, ActorInitials: c.ActorInitials,
		Content: c.Content, IsEdited: c.IsEdited,
		CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toActivityV2(a domain.TaskActivity) *flowtaskv2.TaskActivity {
	return &flowtaskv2.TaskActivity{
		Id: a.ID, TaskId: a.TaskID, UserId: a.UserID,
		ActorName: a.ActorName, ActorInitials: a.ActorInitials,
		Action: a.Action, OldValue: a.OldValue, NewValue: a.NewValue,
		CreatedAt: a.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toAttachmentV2(a domain.TaskAttachment) *flowtaskv2.Attachment {
	return &flowtaskv2.Attachment{
		Id: a.ID, TaskId: a.TaskID, Name: a.Name, Url: a.URL,
		Size: a.Size, UploadedBy: a.UploadedBy,
		CreatedAt: a.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toTimeLogV2(l domain.TimeLog) *flowtaskv2.TimeLog {
	return &flowtaskv2.TimeLog{
		Id: l.ID, TaskId: l.TaskID, UserId: l.UserID,
		DurationMinutes: int64(l.DurationMinutes), Description: l.Description,
		LoggedAt:  l.LoggedAt.Format("2006-01-02T15:04:05Z"),
		CreatedAt: l.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func subtasksToCompactV2(tasks []domain.Task) []*flowtaskv2.TaskPayloadCompact {
	out := make([]*flowtaskv2.TaskPayloadCompact, 0, len(tasks))
	for _, tk := range tasks {
		out = append(out, &flowtaskv2.TaskPayloadCompact{
			Id: tk.ID, Title: tk.Title, Status: tk.Status,
			Priority: tk.Priority,
			DueDate: func() string {
				if tk.DueDate != nil {
					return tk.DueDate.Format("2006-01-02")
				}
				return ""
			}(),
			CreatedAt: tk.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	return out
}

func toTaskV2(t domain.Task) *flowtaskv2.FlowTask {
	dependencies := t.Dependencies
	if dependencies == nil {
		dependencies = []int64{}
	}
	out := &flowtaskv2.FlowTask{
		Id:               t.ID,
		ProjectId:        t.ProjectID,
		OwnerId:          t.OwnerID,
		Title:            t.Title,
		Description:      t.Description,
		Status:           t.Status,
		Priority:         t.Priority,
		ReporterId:       t.ReporterID,
		SortOrder:        int64(t.SortOrder),
		Labels:           t.Labels,
		CreatedAt:        t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Identifier:       t.Identifier,
		ProjectName:      t.ProjectName,
		ProjectColor:     t.ProjectColor,
		AssigneeName:     t.AssigneeName,
		AssigneeInitials: t.AssigneeInitials,
		SubtaskCount:     t.SubtaskCount,
		StatusName:       t.StatusName,
		StatusColor:      t.StatusColor,
		Dependencies:     dependencies,
	}
	if t.ParentTaskID != nil {
		v := *t.ParentTaskID
		out.ParentTaskId = &v
	}
	if t.AssigneeID != nil {
		v := *t.AssigneeID
		out.AssigneeId = &v
	}
	if t.StatusID != nil {
		v := *t.StatusID
		out.StatusId = &v
	}
	if t.DueDate != nil {
		v := t.DueDate.Format("2006-01-02")
		out.DueDate = &v
	}
	if t.StartDate != nil {
		v := t.StartDate.Format("2006-01-02")
		out.StartDate = &v
	}
	if t.CompletedAt != nil {
		v := t.CompletedAt.Format("2006-01-02T15:04:05Z")
		out.CompletedAt = &v
	}
	if t.EstimatedHours != nil {
		v := float64(*t.EstimatedHours)
		out.EstimatedHours = &v
	}
	if t.ActualHours != nil {
		v := float64(*t.ActualHours)
		out.ActualHours = &v
	}
	if t.EffortMinutes != nil {
		v := *t.EffortMinutes
		out.EffortMinutes = &v
	}
	out.FocusPin = t.FocusPin
	if t.SectionID != nil {
		v := *t.SectionID
		out.SectionId = &v
	}
	if t.CycleID != nil {
		v := *t.CycleID
		out.CycleId = &v
	}
	if t.MilestoneID != nil {
		v := *t.MilestoneID
		out.MilestoneId = &v
	}
	if t.Archived {
		out.Archived = true
	}
	return out
}

func tasksToProtoV2(tasks []domain.Task) []*flowtaskv2.FlowTask {
	out := make([]*flowtaskv2.FlowTask, 0, len(tasks))
	for i := range tasks {
		out = append(out, toTaskV2(tasks[i]))
	}
	return out
}

func (a *TaskAPIV2) ListTasks(ctx context.Context, req *flowtaskv2.ListTasksRequest) (*flowtaskv2.ListTasksResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:view"); err != nil {
		return nil, err
	}
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}

	filter := domain.ListTasksFilter{
		ProjectID: req.GetProjectId(),
		Status:    req.GetStatus(),
		Priority:  req.GetPriority(),
		Limit:     int64(req.GetLimit()),
		Offset:    0,
	}
	if req.GetAssigneeId() != 0 {
		v := req.GetAssigneeId()
		filter.AssigneeID = &v
	}
	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	filter.Limit, filter.Offset = int64(limit), offset

	tasks, total, err := a.Service.ListTasksV2(ctx, caller.UserID, businessID, filter)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list tasks")
	}
	return &flowtaskv2.ListTasksResponse{
		Tasks:      tasksToProtoV2(tasks),
		Pagination: buildPagination(page, limit, total, len(tasks)),
	}, nil
}

func (a *TaskAPIV2) GetMyTasks(ctx context.Context, req *flowtaskv2.GetMyTasksRequest) (*flowtaskv2.GetMyTasksResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:view"); err != nil {
		return nil, err
	}
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	page, limit, _ := resolvePagination(req.GetPage(), req.GetLimit())
	tasks, total, err := a.Service.GetMyTasksForBusiness(ctx, caller.UserID, businessID, int64(page), int64(limit))
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get my tasks")
	}
	return &flowtaskv2.GetMyTasksResponse{
		Tasks:      tasksToProtoV2(tasks),
		Pagination: buildPagination(page, limit, total, len(tasks)),
	}, nil
}

func (a *TaskAPIV2) GetMyTasksProject(ctx context.Context, _ *flowtaskv2.GetMyTasksProjectRequest) (*flowtaskv2.GetMyTasksProjectResponse, error) {
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	projectID, _, err := a.Service.EnsureMyTasksProject(ctx, caller.UserID, businessID, 0)
	if err != nil || projectID == 0 {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to resolve My Tasks project")
	}
	return &flowtaskv2.GetMyTasksProjectResponse{ProjectId: projectID}, nil
}

func (a *TaskAPIV2) SearchTasks(ctx context.Context, req *flowtaskv2.SearchTasksRequest) (*flowtaskv2.SearchTasksResponse, error) {
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	tasks, total, err := a.Service.SearchTasksV2(ctx, caller.UserID, businessID, req.GetQ(), int64(limit), offset)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to search tasks")
	}
	return &flowtaskv2.SearchTasksResponse{
		Tasks:      tasksToProtoV2(tasks),
		Pagination: buildPagination(page, limit, total, len(tasks)),
	}, nil
}

func (a *TaskAPIV2) CreateTask(ctx context.Context, req *flowtaskv2.CreateTaskRequest) (*flowtaskv2.CreateTaskResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:create"); err != nil {
		return nil, err
	}
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireTaskAssignmentPermission(ctx, req.AssigneeId, caller.UserID); err != nil {
		return nil, err
	}

	projectID := req.GetProjectId()

	// When project_id is omitted the service provisions the My Tasks bucket and
	// files the task there; the role gate runs against the resolved project.
	if projectID != 0 {
		if roleErr := a.Service.RequireProjectRole(ctx, projectID, caller.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
			return nil, onekiterrors.PermissionDenied()
		}
	}

	dueDate, err := parseTaskDate(req.GetDueDate())
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid due_date")
	}
	startDate, err := parseTaskDate(req.GetStartDate())
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid start_date")
	}

	input := domain.TaskInput{
		ProjectID:      projectID,
		OwnerID:        caller.UserID,
		BusinessID:     businessID,
		ParentTaskID:   req.ParentTaskId,
		Title:          req.GetTitle(),
		Description:    req.GetDescription(),
		Priority:       req.GetPriority(),
		AssigneeID:     req.AssigneeId,
		ReporterID:     caller.UserID,
		DueDate:        dueDate,
		StartDate:      startDate,
		EstimatedHours: req.EstimatedHours,
		Labels:         req.GetLabels(),
	}

	task, createErr := a.Service.CreateTask(ctx, input)
	if createErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create task")
	}

	payload := toTaskV2(task)
	publishTaskEventByProject("task.created", projectID, task.ID, caller.UserID)
	publishBusinessEvent("business.task_created", businessID, task.ID, caller.UserID)
	return &flowtaskv2.CreateTaskResponse{Task: payload}, nil
}

func (a *TaskAPIV2) GetTaskDetail(ctx context.Context, req *flowtaskv2.GetTaskDetailRequest) (*flowtaskv2.GetTaskDetailResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:view"); err != nil {
		return nil, err
	}
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if roleErr := a.Service.RequireTaskRole(ctx, req.GetId(), caller.UserID, businessID, domain.FlowRoleViewer); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	detail, err := a.Service.GetTaskDetailForViewer(ctx, req.GetId(), caller.UserID, businessID)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get task detail")
	}

	comments := make([]*flowtaskv2.TaskComment, 0, len(detail.Comments))
	for _, c := range detail.Comments {
		comments = append(comments, toCommentV2(c))
	}
	activity := make([]*flowtaskv2.TaskActivity, 0, len(detail.Activity))
	for _, act := range detail.Activity {
		activity = append(activity, toActivityV2(act))
	}
	return &flowtaskv2.GetTaskDetailResponse{
		Task:     toTaskV2(detail.Task),
		Subtasks: tasksToProtoV2(detail.Subtasks),
		Comments: comments,
		Activity: activity,
	}, nil
}

func (a *TaskAPIV2) UpdateTask(ctx context.Context, req *flowtaskv2.UpdateTaskRequest) (*flowtaskv2.UpdateTaskResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:edit"); err != nil {
		return nil, err
	}
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if roleErr := a.Service.RequireTaskRole(ctx, req.GetId(), caller.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	body := req.GetBody()
	if body == nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Missing update body")
	}
	if err := requireTaskAssignmentPermission(ctx, body.AssigneeId, caller.UserID); err != nil {
		return nil, err
	}

	dueDate, err := parseTaskDate(body.GetDueDate())
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid due_date")
	}
	startDate, err := parseTaskDate(body.GetStartDate())
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid start_date")
	}

	input := domain.UpdateTaskInput{
		ID:             req.GetId(),
		OwnerID:        caller.UserID,
		BusinessID:     businessID,
		Title:          body.GetTitle(),
		Description:    body.GetDescription(),
		Status:         body.GetStatus(),
		Priority:       body.GetPriority(),
		AssigneeID:     body.AssigneeId,
		DueDate:        dueDate,
		StartDate:      startDate,
		Labels:         body.GetLabels(),
		Archived:       body.Archived,
		EffortMinutes:  body.EffortMinutes,
		EstimatedHours: body.EstimatedHours,
		FocusPin:       body.FocusPin,
	}

	if err := a.Service.UpdateTask(ctx, input, caller.UserID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update task")
	}

	updated, fetchErr := a.Service.GetTaskDetail(ctx, input.ID)
	if fetchErr != nil {
		return &flowtaskv2.UpdateTaskResponse{Message: "Task updated successfully"}, nil
	}

	publishTaskEventByProject("task.updated", updated.Task.ProjectID, input.ID, caller.UserID)
	publishBusinessEvent("business.task_updated", businessID, input.ID, caller.UserID)
	return &flowtaskv2.UpdateTaskResponse{Message: "Task updated successfully"}, nil
}

func (a *TaskAPIV2) UpdateTaskStatus(ctx context.Context, req *flowtaskv2.UpdateTaskStatusRequest) (*flowtaskv2.UpdateTaskStatusResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:edit"); err != nil {
		return nil, err
	}
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if roleErr := a.Service.RequireTaskRole(ctx, req.GetId(), caller.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	body := req.GetBody()
	if body == nil || !domain.IsValidTaskStatus(body.GetStatus()) {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid task status")
	}

	input := domain.UpdateTaskStatusInput{
		ID:         req.GetId(),
		OwnerID:    caller.UserID,
		BusinessID: businessID,
		Status:     body.GetStatus(),
	}
	if err := a.Service.UpdateTaskStatus(ctx, input, caller.UserID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update task status")
	}

	if updated, fetchErr := a.Service.GetTaskDetail(ctx, req.GetId()); fetchErr == nil {
		publishTaskEventByProject("task.status_changed", updated.Task.ProjectID, req.GetId(), caller.UserID)
		publishBusinessEvent("business.task_updated", businessID, req.GetId(), caller.UserID)
	}
	return &flowtaskv2.UpdateTaskStatusResponse{Message: "Task status updated"}, nil
}

func (a *TaskAPIV2) DeleteTask(ctx context.Context, req *flowtaskv2.DeleteTaskRequest) (*flowtaskv2.DeleteTaskResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:edit"); err != nil {
		return nil, err
	}
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if roleErr := a.Service.RequireTaskRole(ctx, req.GetId(), caller.UserID, businessID, domain.FlowRoleAdmin); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	var projectID int64
	if existing, fetchErr := a.Service.GetTaskDetail(ctx, req.GetId()); fetchErr == nil {
		projectID = existing.Task.ProjectID
	}
	if err := a.Service.DeleteTask(ctx, req.GetId(), businessID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete task")
	}
	publishTaskEventByProject("task.deleted", projectID, req.GetId(), caller.UserID)
	return &flowtaskv2.DeleteTaskResponse{Message: "Task deleted successfully"}, nil
}

func (a *TaskAPIV2) CreateSubtask(ctx context.Context, req *flowtaskv2.CreateSubtaskRequest) (*flowtaskv2.CreateSubtaskResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:create"); err != nil {
		return nil, err
	}
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	// Handing the subtask to someone else is the same elevated operation as on
	// create/update; without this a member holding only flow:tasks:create could
	// assign a colleague and grant them project membership.
	if err := requireTaskAssignmentPermission(ctx, req.AssigneeId, caller.UserID); err != nil {
		return nil, err
	}

	parentID := req.GetParentTaskId()
	parent, err := a.Service.GetTaskDetail(ctx, parentID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusNotFound, "Parent task not found")
	}
	// v2 derives everything from the trusted parent row; no client echo of
	// project or workspace ids to mismatch.
	parentProjectID := parent.Task.ProjectID

	if roleErr := a.Service.RequireTaskRole(ctx, parentID, caller.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	dueDate, err := parseTaskDate(req.GetDueDate())
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid due_date")
	}
	startDate, err := parseTaskDate(req.GetStartDate())
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid start_date")
	}

	task, createErr := a.Service.CreateSubtask(ctx, domain.TaskInput{
		ProjectID:      parentProjectID,
		WorkspaceID:    parent.Task.WorkspaceID,
		OwnerID:        caller.UserID,
		BusinessID:     businessID,
		ParentTaskID:   &parentID,
		Title:          req.GetTitle(),
		Description:    req.GetDescription(),
		Priority:       req.GetPriority(),
		AssigneeID:     req.AssigneeId,
		ReporterID:     caller.UserID,
		DueDate:        dueDate,
		StartDate:      startDate,
		EstimatedHours: req.EstimatedHours,
	})
	if createErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create subtask")
	}
	return &flowtaskv2.CreateSubtaskResponse{TaskId: task.ID, Message: "Subtask created successfully"}, nil
}

func (a *TaskAPIV2) GetSubtasks(ctx context.Context, req *flowtaskv2.GetSubtasksRequest) (*flowtaskv2.GetSubtasksResponse, error) {
	caller, _, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if !a.Service.IsTaskMember(ctx, caller.UserID, req.GetTaskId()) {
		return nil, onekiterrors.PermissionDenied()
	}

	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	subtasks, total, err := a.Service.GetSubtasks(ctx, req.GetTaskId(), int64(limit), offset)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get subtasks")
	}
	return &flowtaskv2.GetSubtasksResponse{
		Subtasks:   subtasksToCompactV2(subtasks),
		Pagination: buildPagination(page, limit, total, len(subtasks)),
	}, nil
}

func (a *TaskAPIV2) AddComment(ctx context.Context, req *flowtaskv2.AddCommentRequest) (*flowtaskv2.AddCommentResponse, error) {
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if !a.Service.IsTaskMember(ctx, caller.UserID, req.GetTaskId()) {
		return nil, onekiterrors.PermissionDenied()
	}
	if roleErr := a.Service.RequireTaskRole(ctx, req.GetTaskId(), caller.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	comment, err := a.Service.AddComment(ctx, req.GetTaskId(), caller.UserID, req.GetContent())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to add comment")
	}
	publishCommentEvent("comment.added", req.GetTaskId(), caller.UserID)
	return &flowtaskv2.AddCommentResponse{Comment: toCommentV2(comment)}, nil
}

func (a *TaskAPIV2) DeleteComment(ctx context.Context, req *flowtaskv2.DeleteCommentRequest) (*flowtaskv2.DeleteCommentResponse, error) {
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if roleErr := a.Service.RequireCommentRole(ctx, req.GetId(), caller.UserID, businessID, domain.FlowRoleAdmin); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if err := a.Service.DeleteComment(ctx, req.GetId(), caller.UserID); err != nil {
		// The delete is scoped to the comment's author, so a comment written by
		// someone else reports ErrUnauthorized: the caller may administer the
		// project but not this row. That is a permission answer, not an outage.
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete comment")
	}
	return &flowtaskv2.DeleteCommentResponse{Message: "Comment deleted successfully"}, nil
}

func (a *TaskAPIV2) GetTaskActivity(ctx context.Context, req *flowtaskv2.GetTaskActivityRequest) (*flowtaskv2.GetTaskActivityResponse, error) {
	caller, _, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if !a.Service.IsTaskMember(ctx, caller.UserID, req.GetTaskId()) {
		return nil, onekiterrors.PermissionDenied()
	}
	activities, err := a.Service.GetTaskActivity(ctx, req.GetTaskId())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get task activity")
	}
	out := make([]*flowtaskv2.TaskActivity, 0, len(activities))
	for _, act := range activities {
		out = append(out, toActivityV2(act))
	}
	return &flowtaskv2.GetTaskActivityResponse{Activity: out}, nil
}

func (a *TaskAPIV2) CreateAttachment(ctx context.Context, req *flowtaskv2.CreateAttachmentRequest) (*flowtaskv2.CreateAttachmentResponse, error) {
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if roleErr := a.Service.RequireTaskRole(ctx, req.GetTaskId(), caller.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	id, createErr := a.Service.CreateAttachment(ctx, domain.TaskAttachmentInput{
		TaskID: req.GetTaskId(), OwnerID: caller.UserID, BusinessID: businessID,
		Name: req.GetName(), URL: req.GetUrl(), Size: req.GetSize(), UploadedBy: caller.UserID,
	})
	if createErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create attachment")
	}
	return &flowtaskv2.CreateAttachmentResponse{AttachmentId: id, Message: "Attachment created successfully"}, nil
}

func (a *TaskAPIV2) GetAttachments(ctx context.Context, req *flowtaskv2.GetAttachmentsRequest) (*flowtaskv2.GetAttachmentsResponse, error) {
	caller, _, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if !a.Service.IsTaskMember(ctx, caller.UserID, req.GetTaskId()) {
		return nil, onekiterrors.PermissionDenied()
	}
	attachments, err := a.Service.GetAttachments(ctx, req.GetTaskId())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get attachments")
	}
	out := make([]*flowtaskv2.Attachment, 0, len(attachments))
	for _, att := range attachments {
		out = append(out, toAttachmentV2(att))
	}
	return &flowtaskv2.GetAttachmentsResponse{Attachments: out}, nil
}

func (a *TaskAPIV2) DeleteAttachment(ctx context.Context, req *flowtaskv2.DeleteAttachmentRequest) (*flowtaskv2.DeleteAttachmentResponse, error) {
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if roleErr := a.Service.RequireAttachmentRole(ctx, req.GetId(), caller.UserID, businessID, domain.FlowRoleAdmin); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if err := a.Service.DeleteAttachment(ctx, req.GetId(), businessID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete attachment")
	}
	return &flowtaskv2.DeleteAttachmentResponse{Message: "Attachment deleted successfully"}, nil
}

func (a *TaskAPIV2) CreateTimeLog(ctx context.Context, req *flowtaskv2.CreateTimeLogRequest) (*flowtaskv2.CreateTimeLogResponse, error) {
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	taskID := req.GetTaskId()
	if taskID == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "task_id is required")
	}
	if roleErr := a.Service.RequireTaskRole(ctx, taskID, caller.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	id, err := a.Service.CreateTimeLog(ctx, domain.TimeLogInput{
		OwnerID: caller.UserID, BusinessID: businessID, TaskID: taskID,
		UserID: caller.UserID, DurationMinutes: int(req.GetDurationMinutes()),
		Description: req.GetDescription(),
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create time log")
	}
	return &flowtaskv2.CreateTimeLogResponse{TimeLogId: id, Message: "Time log created successfully"}, nil
}

func (a *TaskAPIV2) GetTimeLogs(ctx context.Context, req *flowtaskv2.GetTimeLogsRequest) (*flowtaskv2.GetTimeLogsResponse, error) {
	caller, _, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if !a.Service.IsTaskMember(ctx, caller.UserID, req.GetTaskId()) {
		return nil, onekiterrors.PermissionDenied()
	}
	logs, err := a.Service.GetTimeLogs(ctx, req.GetTaskId())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get time logs")
	}
	out := make([]*flowtaskv2.TimeLog, 0, len(logs))
	for _, l := range logs {
		out = append(out, toTimeLogV2(l))
	}
	return &flowtaskv2.GetTimeLogsResponse{TimeLogs: out}, nil
}

func (a *TaskAPIV2) DeleteTimeLog(ctx context.Context, req *flowtaskv2.DeleteTimeLogRequest) (*flowtaskv2.DeleteTimeLogResponse, error) {
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if roleErr := a.Service.RequireTimeLogRole(ctx, req.GetId(), caller.UserID, businessID, domain.FlowRoleAdmin); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if err := a.Service.DeleteTimeLogFlat(ctx, req.GetId(), caller.UserID); err != nil {
		// Same shape as DeleteComment: a log written by someone else is a
		// permission answer from the row-scoped delete, not an outage.
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete time log")
	}
	return &flowtaskv2.DeleteTimeLogResponse{Message: "Time log deleted successfully"}, nil
}

func (a *TaskAPIV2) AddTaskDependency(ctx context.Context, req *flowtaskv2.AddTaskDependencyRequest) (*flowtaskv2.AddTaskDependencyResponse, error) {
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if roleErr := a.Service.RequireTaskRole(ctx, req.GetTaskId(), caller.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	body := req.GetBody()
	if body == nil || body.GetDependsOnTaskId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "depends_on_task_id is required")
	}
	if err := a.Service.AddTaskDependency(ctx, req.GetTaskId(), body.GetDependsOnTaskId(), businessID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to add dependency")
	}
	return &flowtaskv2.AddTaskDependencyResponse{Message: "Dependency added"}, nil
}

func (a *TaskAPIV2) RemoveTaskDependency(ctx context.Context, req *flowtaskv2.RemoveTaskDependencyRequest) (*flowtaskv2.RemoveTaskDependencyResponse, error) {
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if roleErr := a.Service.RequireTaskRole(ctx, req.GetTaskId(), caller.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if err := a.Service.RemoveTaskDependency(ctx, req.GetTaskId(), req.GetDependsOnTaskId(), businessID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to remove dependency")
	}
	return &flowtaskv2.RemoveTaskDependencyResponse{Message: "Dependency removed"}, nil
}
