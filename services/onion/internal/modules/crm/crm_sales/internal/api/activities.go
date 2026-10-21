package api

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	crmsalesv1 "github.com/voxire/voxie/gen/go/crm/sales/v1"
)

func (a *API) LogActivity(ctx context.Context, req *crmsalesv1.LogActivityRequest) (*crmsalesv1.LogActivityResponse, error) {
	c, err := a.caller(ctx, permCreate)
	if err != nil {
		return nil, err
	}
	in := domain.ActivityInput{
		LeadID:  positivePtr(req.LeadId),
		DealID:  positivePtr(req.DealId),
		Type:    domain.ActivityType(req.Type),
		Outcome: domain.CallOutcome(req.Outcome),
		Notes:   req.Notes,
	}
	if req.DurationSeconds != nil {
		seconds := int(*req.DurationSeconds)
		in.DurationSeconds = &seconds
	}
	if req.OccurredAt != nil && *req.OccurredAt > 0 {
		at := time.Unix(*req.OccurredAt, 0).UTC()
		in.OccurredAt = &at
	}
	if req.FollowUp != nil {
		followUp, err := taskFieldsFromGen(req.FollowUp)
		if err != nil {
			return nil, err
		}
		in.FollowUp = &followUp
	}
	activity, task, err := a.Service.LogActivity(ctx, c.businessID, c.userID, in)
	if err != nil {
		return nil, fail(ctx, "log activity", err)
	}
	response := &crmsalesv1.LogActivityResponse{Activity: activityToGen(activity)}
	if task != nil {
		response.FollowUp = taskToGen(*task)
	}
	return response, nil
}

func (a *API) ListLeadActivities(ctx context.Context, req *crmsalesv1.ListLeadActivitiesRequest) (*crmsalesv1.ListActivitiesResponse, error) {
	c, err := a.caller(ctx, permView)
	if err != nil {
		return nil, err
	}
	page, err := a.Service.ListLeadActivities(ctx, c.businessID, req.LeadId, req.Cursor, int(req.Limit))
	if err != nil {
		return nil, fail(ctx, "list lead activities", err)
	}
	return activityPageToGen(page), nil
}

func (a *API) ListDealActivities(ctx context.Context, req *crmsalesv1.ListDealActivitiesRequest) (*crmsalesv1.ListActivitiesResponse, error) {
	c, err := a.caller(ctx, permView)
	if err != nil {
		return nil, err
	}
	page, err := a.Service.ListDealActivities(ctx, c.businessID, req.DealId, req.Cursor, int(req.Limit))
	if err != nil {
		return nil, fail(ctx, "list deal activities", err)
	}
	return activityPageToGen(page), nil
}

func (a *API) DeleteActivity(ctx context.Context, req *crmsalesv1.DeleteActivityRequest) (*crmsalesv1.DeleteActivityResponse, error) {
	c, err := a.caller(ctx, permDelete)
	if err != nil {
		return nil, err
	}
	if err := a.Service.DeleteActivity(ctx, c.businessID, req.Id); err != nil {
		return nil, fail(ctx, "delete activity", err)
	}
	return &crmsalesv1.DeleteActivityResponse{Deleted: true}, nil
}

func (a *API) ListTasks(ctx context.Context, req *crmsalesv1.ListSalesTasksRequest) (*crmsalesv1.ListSalesTasksResponse, error) {
	c, err := a.caller(ctx, permView)
	if err != nil {
		return nil, err
	}
	page, err := a.Service.ListTasks(ctx, c.businessID, domain.TaskFilter{
		Scope:      domain.TaskScope(req.Scope),
		AssignedTo: req.AssignedTo,
		Due:        domain.FollowUpWindow(req.Due),
		LeadID:     positivePtr(req.LeadId),
		DealID:     positivePtr(req.DealId),
		Cursor:     req.Cursor,
		Limit:      int(req.Limit),
	}, a.callerDay(req.Today, 0))
	if err != nil {
		return nil, fail(ctx, "list tasks", err)
	}
	tasks := make([]*crmsalesv1.SalesTask, 0, len(page.Tasks))
	for _, task := range page.Tasks {
		tasks = append(tasks, taskToGen(task))
	}
	return &crmsalesv1.ListSalesTasksResponse{Tasks: tasks, NextCursor: page.NextCursor}, nil
}

func (a *API) GetTaskCounts(ctx context.Context, req *crmsalesv1.GetSalesTaskCountsRequest) (*crmsalesv1.GetSalesTaskCountsResponse, error) {
	c, err := a.caller(ctx, permView)
	if err != nil {
		return nil, err
	}
	counts, err := a.Service.CountTasks(ctx, c.businessID, req.AssignedTo, a.callerDay(req.Today, 0))
	if err != nil {
		return nil, fail(ctx, "count tasks", err)
	}
	return &crmsalesv1.GetSalesTaskCountsResponse{Overdue: counts.Overdue, DueToday: counts.DueToday, Upcoming: counts.Upcoming}, nil
}

func (a *API) CreateTask(ctx context.Context, req *crmsalesv1.CreateSalesTaskRequest) (*crmsalesv1.SalesTaskResponse, error) {
	c, err := a.caller(ctx, permCreate)
	if err != nil {
		return nil, err
	}
	in, err := taskFieldsFromGen(req.Task)
	if err != nil {
		return nil, err
	}
	task, err := a.Service.CreateTask(ctx, c.businessID, c.userID, in)
	if err != nil {
		return nil, fail(ctx, "create task", err)
	}
	return &crmsalesv1.SalesTaskResponse{Task: taskToGen(task)}, nil
}

func (a *API) UpdateTask(ctx context.Context, req *crmsalesv1.UpdateSalesTaskRequest) (*crmsalesv1.SalesTaskResponse, error) {
	c, err := a.caller(ctx, permEdit)
	if err != nil {
		return nil, err
	}
	in, err := taskFieldsFromGen(req.Body)
	if err != nil {
		return nil, err
	}
	task, err := a.Service.UpdateTask(ctx, c.businessID, req.Id, in)
	if err != nil {
		return nil, fail(ctx, "update task", err)
	}
	return &crmsalesv1.SalesTaskResponse{Task: taskToGen(task)}, nil
}

func (a *API) CompleteTask(ctx context.Context, req *crmsalesv1.CompleteSalesTaskRequest) (*crmsalesv1.SalesTaskResponse, error) {
	c, err := a.caller(ctx, permEdit)
	if err != nil {
		return nil, err
	}
	done := req.Body != nil && req.Body.Done
	task, err := a.Service.CompleteTask(ctx, c.businessID, req.Id, done)
	if err != nil {
		return nil, fail(ctx, "complete task", err)
	}
	return &crmsalesv1.SalesTaskResponse{Task: taskToGen(task)}, nil
}

func (a *API) DeleteTask(ctx context.Context, req *crmsalesv1.DeleteSalesTaskRequest) (*crmsalesv1.DeleteSalesTaskResponse, error) {
	c, err := a.caller(ctx, permDelete)
	if err != nil {
		return nil, err
	}
	if err := a.Service.DeleteTask(ctx, c.businessID, req.Id); err != nil {
		return nil, fail(ctx, "delete task", err)
	}
	return &crmsalesv1.DeleteSalesTaskResponse{Deleted: true}, nil
}
