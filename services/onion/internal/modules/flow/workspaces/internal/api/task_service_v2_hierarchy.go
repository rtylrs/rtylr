package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	flowtaskv2 "github.com/voxire/voxie/gen/go/flow/task/v2"
)

// hierarchyError maps a type or parent rule violation onto a client error
// carrying the rule that was broken. Only the sentinel's own fixed message is
// sent, never the wrapped error. It returns nil for anything else, so the
// caller keeps its own generic failure for real faults.
func hierarchyError(err error) error {
	if errors.Is(err, domain.ErrTaskParentNotFound) {
		return onekiterrors.New(http.StatusNotFound, domain.ErrTaskParentNotFound.Error())
	}
	if rule := domain.HierarchyRuleError(err); rule != nil {
		return onekiterrors.New(http.StatusBadRequest, rule.Error())
	}
	return nil
}

// parseTaskTypeFilter reads the comma-separated task_type query value.
func parseTaskTypeFilter(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	types := make([]string, 0, len(parts))
	for _, part := range parts {
		taskType := strings.TrimSpace(part)
		if !domain.IsValidTaskType(taskType) {
			return nil, domain.ErrInvalidTaskType
		}
		types = append(types, taskType)
	}
	return types, nil
}

func (a *TaskAPIV2) SetTaskParent(ctx context.Context, req *flowtaskv2.SetTaskParentRequest) (*flowtaskv2.SetTaskParentResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:edit"); err != nil {
		return nil, err
	}
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	// The parent has to share the task's project, so the role held on the task
	// is the role held on the parent too.
	if roleErr := a.Service.RequireTaskRole(ctx, req.GetId(), caller.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	var parentID *int64
	if body := req.GetBody(); body != nil {
		parentID = body.ParentId
	}
	task, err := a.Service.SetTaskParent(ctx, req.GetId(), caller.UserID, businessID, parentID)
	if err != nil {
		if mapped := hierarchyError(err); mapped != nil {
			return nil, mapped
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to move task")
	}

	publishTaskEventByProject("task.updated", task.ProjectID, task.ID, caller.UserID)
	publishBusinessEvent("business.task_updated", businessID, task.ID, caller.UserID)
	return &flowtaskv2.SetTaskParentResponse{Task: toTaskV2(task)}, nil
}

func (a *TaskAPIV2) ListEpicRollups(ctx context.Context, req *flowtaskv2.ListEpicRollupsRequest) (*flowtaskv2.ListEpicRollupsResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:view"); err != nil {
		return nil, err
	}
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if roleErr := a.Service.RequireProjectRole(ctx, req.GetProjectId(), caller.UserID, businessID, domain.FlowRoleViewer); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	rollups, err := a.Service.ListEpicRollups(ctx, req.GetProjectId(), businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list epics")
	}
	out := make([]*flowtaskv2.EpicRollup, 0, len(rollups))
	for _, r := range rollups {
		out = append(out, &flowtaskv2.EpicRollup{
			EpicId:              r.EpicID,
			Identifier:          r.Identifier,
			Title:               r.Title,
			Color:               r.Color,
			Status:              r.Status,
			TotalTasks:          clampInt32(r.TotalTasks),
			BacklogTasks:        clampInt32(r.BacklogTasks),
			ActiveTasks:         clampInt32(r.ActiveTasks),
			DoneTasks:           clampInt32(r.DoneTasks),
			CancelledTasks:      clampInt32(r.CancelledTasks),
			EstimateMinutes:     clampInt32(r.EstimateMinutes),
			DoneEstimateMinutes: clampInt32(r.DoneEstimateMinutes),
			LoggedMinutes:       clampInt32(r.LoggedMinutes),
		})
	}
	return &flowtaskv2.ListEpicRollupsResponse{Epics: out}, nil
}

func clampInt32(v int64) int32 {
	const maxInt32 = int64(1<<31 - 1)
	if v > maxInt32 {
		return int32(maxInt32)
	}
	if v < 0 {
		return 0
	}
	return int32(v) // #nosec G115 -- Bounded to the int32 range above.
}
