package api

import (
	"context"
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
		Id:                t.ID,
		ProjectId:         t.ProjectID,
		OwnerId:           t.OwnerID,
		Title:             t.Title,
		Description:       t.Description,
		Status:            t.Status,
		Priority:          t.Priority,
		ReporterId:        t.ReporterID,
		SortOrder:         int64(t.SortOrder),
		Labels:            t.Labels,
		CreatedAt:         t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:         t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Identifier:        t.Identifier,
		ProjectName:       t.ProjectName,
		ProjectColor:      t.ProjectColor,
		AssigneeName:      t.AssigneeName,
		AssigneeInitials:  t.AssigneeInitials,
		SubtaskCount:      t.SubtaskCount,
		CompletedSubtasks: t.CompletedSubtasks,
		CommentCount:      t.CommentCount,
		AttachmentCount:   t.AttachmentCount,
		StatusName:        t.StatusName,
		StatusColor:       t.StatusColor,
		Dependencies:      dependencies,
		WaitingOn:         t.WaitingOn,
		DecisionNote:      t.DecisionNote,
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
