package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/wsbus"
	flowcyclev2 "github.com/voxire/voxie/gen/go/flow/cycle/v2"
)

// Cycle planning (T-570, T-336). Reads need the viewer tier on the cycle's
// project, writes the member tier; a task move needs the member tier on the
// task. Every lookup is scoped by the caller's business, and a cycle of another
// business answers 404 rather than 403 so ids do not leak.

// cyclePlanningError maps a planning error onto the HTTP status the client
// can act on. Anything unknown is a 500.
func cyclePlanningError(ctx context.Context, operation string, err error) error {
	var decision *domain.CycleDecisionError
	switch {
	case errors.As(err, &decision):
		return onekiterrors.New(http.StatusBadRequest, decision.Reason)
	case errors.Is(err, domain.ErrCycleNotFound), errors.Is(err, domain.ErrProjectNotFound):
		return onekiterrors.New(http.StatusNotFound, "Cycle not found")
	case errors.Is(err, domain.ErrCycleTaskNotFound):
		return onekiterrors.New(http.StatusNotFound, "Task not found in this project")
	case errors.Is(err, domain.ErrCycleAlreadyActive):
		return onekiterrors.New(http.StatusConflict, "This project already has an active cycle. Complete it before starting another.")
	case errors.Is(err, domain.ErrCycleNotUpcoming):
		return onekiterrors.New(http.StatusConflict, "Only an upcoming cycle can be started")
	case errors.Is(err, domain.ErrCycleCompleted):
		return onekiterrors.New(http.StatusConflict, "The cycle is completed")
	case errors.Is(err, domain.ErrCycleRankAnchor):
		return onekiterrors.New(http.StatusBadRequest, "The anchor task is not in the target list")
	case errors.Is(err, domain.ErrInvalidCycleMove):
		return onekiterrors.New(http.StatusBadRequest, "Give at most one of before_task_id and after_task_id")
	case errors.Is(err, domain.ErrInvalidCycleStatus):
		return onekiterrors.New(http.StatusBadRequest, "status must be active or completed")
	case errors.Is(err, domain.ErrInvalidCycleRequest):
		return onekiterrors.New(http.StatusBadRequest, "Invalid cycle request")
	case errors.Is(err, domain.ErrUnauthorized):
		return onekiterrors.PermissionDenied()
	}
	return workspaceInternalError(ctx, operation+" failed", err)
}

// cycleCaller is who is acting on which cycle's project.
type cycleCaller struct {
	UserID     int64
	BusinessID int64
	ProjectID  int64
}

// authorizeCycle resolves the caller and the cycle's project and checks the
// caller holds minRole on it.
func (a *CycleAPIV2) authorizeCycle(ctx context.Context, cycleID int64, minRole string) (cycleCaller, error) {
	caller, businessID, err := resolveV2(ctx)
	if err != nil {
		return cycleCaller{}, err
	}
	projectID, err := a.Service.CycleProjectID(ctx, businessID, cycleID)
	if err != nil {
		return cycleCaller{}, cyclePlanningError(ctx, "Load cycle", err)
	}
	if roleErr := a.Service.RequireProjectRole(ctx, projectID, caller.UserID, businessID, minRole); roleErr != nil {
		return cycleCaller{}, onekiterrors.PermissionDenied()
	}
	return cycleCaller{UserID: caller.UserID, BusinessID: businessID, ProjectID: projectID}, nil
}

// publishCycleEvent tells the project's open screens that its cycles changed,
// so the Backlog, Cycles page and task filters refetch.
func publishCycleEvent(eventType string, c cycleCaller, cycleID int64) {
	wsbus.Publish(wsbus.Event{
		Topic:    fmt.Sprintf("project:%d", c.ProjectID),
		Type:     eventType,
		EntityID: cycleID,
		ActorID:  c.UserID,
	})
}

func parseCycleDate(value, field string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, onekiterrors.New(http.StatusBadRequest, "Invalid "+field)
	}
	return parsed, nil
}

func (a *CycleAPIV2) GetCycle(ctx context.Context, req *flowcyclev2.GetCycleRequest) (*flowcyclev2.GetCycleResponse, error) {
	c, err := a.authorizeCycle(ctx, req.GetId(), domain.FlowRoleViewer)
	if err != nil {
		return nil, err
	}
	businessID := c.BusinessID
	cycle, err := a.Service.GetCycleV2(ctx, businessID, req.GetId())
	if err != nil {
		return nil, cyclePlanningError(ctx, "GetCycle", err)
	}
	return &flowcyclev2.GetCycleResponse{Cycle: toContractCycleV2(cycle)}, nil
}

// DeleteCycle needs the admin tier on the cycle's project, like deleting a
// milestone. The cycle's tasks stay in the project, unplanned.
func (a *CycleAPIV2) DeleteCycle(ctx context.Context, req *flowcyclev2.DeleteCycleRequest) (*flowcyclev2.DeleteCycleResponse, error) {
	c, err := a.authorizeCycle(ctx, req.GetId(), domain.FlowRoleAdmin)
	if err != nil {
		return nil, err
	}
	if err := a.Service.DeleteCycle(ctx, c.BusinessID, req.GetId()); err != nil {
		return nil, cyclePlanningError(ctx, "DeleteCycle", err)
	}
	publishCycleEvent("cycle.deleted", c, req.GetId())
	return &flowcyclev2.DeleteCycleResponse{}, nil
}

// UpdateCycle edits name, dates and goal. A status change goes through the
// same rules as StartCycle and CompleteCycle, so this older entry point cannot
// open a second active cycle or strand unfinished tasks.
func (a *CycleAPIV2) UpdateCycle(ctx context.Context, req *flowcyclev2.UpdateCycleRequest) (*flowcyclev2.UpdateCycleResponse, error) {
	start, err := parseCycleDate(req.GetStartDate(), "start_date")
	if err != nil {
		return nil, err
	}
	end, err := parseCycleDate(req.GetEndDate(), "end_date")
	if err != nil {
		return nil, err
	}
	c, err := a.authorizeCycle(ctx, req.GetId(), domain.FlowRoleMember)
	if err != nil {
		return nil, err
	}
	actorID, businessID := c.UserID, c.BusinessID
	existing, err := a.Service.GetCycleV2(ctx, businessID, req.GetId())
	if err != nil {
		return nil, cyclePlanningError(ctx, "UpdateCycle", err)
	}
	effectiveStart, effectiveEnd := existing.StartDate, existing.EndDate
	if !start.IsZero() {
		effectiveStart = start
	}
	if !end.IsZero() {
		effectiveEnd = end
	}
	if !effectiveStart.IsZero() && !effectiveEnd.IsZero() && effectiveEnd.Before(effectiveStart) {
		return nil, onekiterrors.New(http.StatusBadRequest, "end_date cannot be before start_date")
	}
	name := strings.TrimSpace(req.GetName())
	if err := a.Service.UpdateCycleV2(ctx, domain.UpdateCycleV2Input{
		ID: req.GetId(), BusinessID: businessID, Name: name,
		StartDate: start, EndDate: end, Goal: req.Goal,
	}); err != nil {
		return nil, cyclePlanningError(ctx, "UpdateCycle", err)
	}

	switch req.GetStatus() {
	case "", existing.Status:
	case domain.CycleStatusActive:
		if _, err := a.Service.StartCycle(ctx, domain.StartCycleInput{ID: req.GetId(), BusinessID: businessID}); err != nil {
			return nil, cyclePlanningError(ctx, "UpdateCycle", err)
		}
	case domain.CycleStatusCompleted:
		if _, err := a.Service.CompleteCycle(ctx, domain.CompleteCycleInput{ID: req.GetId(), BusinessID: businessID, ActorID: actorID}); err != nil {
			return nil, cyclePlanningError(ctx, "UpdateCycle", err)
		}
	default:
		return nil, cyclePlanningError(ctx, "UpdateCycle", domain.ErrInvalidCycleStatus)
	}
	publishCycleEvent("cycle.updated", c, req.GetId())
	return &flowcyclev2.UpdateCycleResponse{}, nil
}

func (a *CycleAPIV2) StartCycle(ctx context.Context, req *flowcyclev2.StartCycleRequest) (*flowcyclev2.StartCycleResponse, error) {
	c, err := a.authorizeCycle(ctx, req.GetId(), domain.FlowRoleMember)
	if err != nil {
		return nil, err
	}
	businessID := c.BusinessID
	start, err := parseCycleDate(req.GetStartDate(), "start_date")
	if err != nil {
		return nil, err
	}
	weeks := req.GetLengthWeeks()
	if weeks < 0 || weeks > 4 {
		return nil, onekiterrors.New(http.StatusBadRequest, "length_weeks must be between 1 and 4")
	}
	cycle, err := a.Service.StartCycle(ctx, domain.StartCycleInput{
		ID: req.GetId(), BusinessID: businessID, StartDate: start,
		LengthDays: int(weeks) * 7, Goal: req.Goal,
	})
	if err != nil {
		return nil, cyclePlanningError(ctx, "StartCycle", err)
	}
	publishCycleEvent("cycle.started", c, req.GetId())
	return &flowcyclev2.StartCycleResponse{Cycle: toContractCycleV2(cycle)}, nil
}

func (a *CycleAPIV2) CompleteCycle(ctx context.Context, req *flowcyclev2.CompleteCycleRequest) (*flowcyclev2.CompleteCycleResponse, error) {
	c, err := a.authorizeCycle(ctx, req.GetId(), domain.FlowRoleMember)
	if err != nil {
		return nil, err
	}
	actorID, businessID := c.UserID, c.BusinessID
	decisions := make([]domain.CycleTaskDecision, 0, len(req.GetDecisions()))
	for _, d := range req.GetDecisions() {
		if d == nil {
			continue
		}
		decisions = append(decisions, domain.CycleTaskDecision{
			TaskID: d.GetTaskId(), Decision: d.GetDecision(), ToCycleID: d.GetToCycleId(),
		})
	}
	cycle, err := a.Service.CompleteCycle(ctx, domain.CompleteCycleInput{
		ID: req.GetId(), BusinessID: businessID, ActorID: actorID, Decisions: decisions,
	})
	if err != nil {
		return nil, cyclePlanningError(ctx, "CompleteCycle", err)
	}
	publishCycleEvent("cycle.completed", c, req.GetId())
	return &flowcyclev2.CompleteCycleResponse{Cycle: toContractCycleV2(cycle)}, nil
}

func (a *CycleAPIV2) MoveCycleTask(ctx context.Context, req *flowcyclev2.MoveCycleTaskRequest) (*flowcyclev2.MoveCycleTaskResponse, error) {
	// Moving a task between lists is a task edit, the same permission the
	// route policy applies to PUT /cycles/tasks/order.
	if err := requireFlowPermission(ctx, "flow:tasks:edit"); err != nil {
		return nil, err
	}
	caller, businessID, err := resolveV2(ctx)
	if err != nil {
		return nil, err
	}
	taskID := req.GetTaskId()
	if roleErr := a.Service.RequireTaskRole(ctx, taskID, caller.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	moved, err := a.Service.MoveCycleTask(ctx, domain.MoveCycleTaskInput{
		BusinessID: businessID, TaskID: taskID, CycleID: req.GetCycleId(),
		BeforeTaskID: req.GetBeforeTaskId(), AfterTaskID: req.GetAfterTaskId(),
	})
	if err != nil {
		return nil, cyclePlanningError(ctx, "MoveCycleTask", err)
	}
	publishTaskEventByProject("task.updated", moved.ProjectID, taskID, caller.UserID)
	out := &flowcyclev2.MoveCycleTaskResponse{PlanRank: moved.Rank}
	if cycleID := req.GetCycleId(); cycleID != 0 {
		out.CycleId = &cycleID
	}
	return out, nil
}

func (a *CycleAPIV2) GetCycleCapacity(ctx context.Context, req *flowcyclev2.GetCycleCapacityRequest) (*flowcyclev2.GetCycleCapacityResponse, error) {
	c, err := a.authorizeCycle(ctx, req.GetId(), domain.FlowRoleViewer)
	if err != nil {
		return nil, err
	}
	businessID := c.BusinessID
	capacity, err := a.Service.GetCycleCapacity(ctx, businessID, req.GetId())
	if err != nil {
		return nil, cyclePlanningError(ctx, "GetCycleCapacity", err)
	}
	out := &flowcyclev2.GetCycleCapacityResponse{
		LeaveApplied:      capacity.LeaveApplied,
		MinutesPerDay:     clampInt32(int64(capacity.MinutesPerDay)),
		UnassignedMinutes: clampInt32(int64(capacity.UnassignedMinutes)),
		UnassignedCount:   clampInt32(int64(capacity.UnassignedCount)),
	}
	for _, m := range capacity.Members {
		out.Members = append(out.Members, &flowcyclev2.CycleMemberCapacity{
			UserId:           m.UserID,
			Name:             m.Name,
			WorkingDays:      clampInt32(int64(m.WorkingDays)),
			LeaveDays:        clampInt32(int64(m.LeaveDays)),
			AvailableMinutes: clampInt32(int64(m.AvailableMinutes)),
			CommittedMinutes: clampInt32(int64(m.CommittedMinutes)),
			TaskCount:        clampInt32(int64(m.TaskCount)),
		})
	}
	return out, nil
}
