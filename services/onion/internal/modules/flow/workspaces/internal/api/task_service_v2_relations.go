package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	flowtaskv2 "github.com/voxire/voxie/gen/go/flow/task/v2"
)

// relationError maps a relation failure onto the status a client can act on:
// 400 for a request that can never work, 409 for one the current relations
// rule out, 404 for a relation that is not there.
func relationError(err error, fallback string) error {
	switch {
	case errors.Is(err, domain.ErrUnauthorized):
		return onekiterrors.PermissionDenied()
	case errors.Is(err, domain.ErrRelationInvalidType):
		return onekiterrors.New(http.StatusBadRequest, "type must be one of blocks, relates, duplicates, clones")
	case errors.Is(err, domain.ErrRelationInvalidDirection):
		return onekiterrors.New(http.StatusBadRequest, "direction must be outward or inward")
	case errors.Is(err, domain.ErrRelationSelf):
		return onekiterrors.New(http.StatusBadRequest, "A task cannot be related to itself")
	case errors.Is(err, domain.ErrRelationExists):
		return onekiterrors.New(http.StatusConflict, "These tasks already have this relation")
	case errors.Is(err, domain.ErrRelationCycle):
		return onekiterrors.New(http.StatusConflict, "This would make the tasks block each other")
	case errors.Is(err, domain.ErrRelationNotFound):
		return onekiterrors.New(http.StatusNotFound, "Relation not found")
	}
	return onekiterrors.New(http.StatusInternalServerError, fallback)
}

// publishRelationChange tells both ends' subscribers their task changed, with
// the events a task edit already sends.
func publishRelationChange(change domain.TaskRelationChange, businessID, actorID int64) {
	for _, task := range change.Tasks {
		if task.ID == 0 {
			continue
		}
		publishTaskEventByProject("task.updated", task.ProjectID, task.ID, actorID)
		publishBusinessEvent("business.task_updated", businessID, task.ID, actorID)
	}
}

func toRelationGroupsV2(groups []domain.TaskRelationGroup) []*flowtaskv2.TaskRelationGroup {
	out := make([]*flowtaskv2.TaskRelationGroup, 0, len(groups))
	for _, group := range groups {
		tasks := make([]*flowtaskv2.RelatedTask, 0, len(group.Relations))
		for _, relation := range group.Relations {
			related := &flowtaskv2.RelatedTask{
				RelationId:       relation.ID,
				TaskId:           relation.Task.ID,
				Identifier:       relation.Task.Identifier,
				Title:            relation.Task.Title,
				ProjectId:        relation.Task.ProjectID,
				ProjectName:      relation.Task.ProjectName,
				Status:           relation.Task.Status,
				StatusName:       relation.Task.StatusName,
				StatusColor:      relation.Task.StatusColor,
				StatusBucket:     relation.Task.StatusBucket,
				AssigneeName:     relation.Task.AssigneeName,
				AssigneeInitials: relation.Task.AssigneeInitials,
			}
			if relation.Task.AssigneeID != nil {
				assigneeID := *relation.Task.AssigneeID
				related.AssigneeId = &assigneeID
			}
			tasks = append(tasks, related)
		}
		out = append(out, &flowtaskv2.TaskRelationGroup{
			Type:      group.Type,
			Direction: group.Direction,
			Relation:  group.Relation,
			Tasks:     tasks,
		})
	}
	return out
}

func (a *TaskAPIV2) AddTaskRelation(ctx context.Context, req *flowtaskv2.AddTaskRelationRequest) (*flowtaskv2.AddTaskRelationResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:edit"); err != nil {
		return nil, err
	}
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if roleErr := a.Service.RequireTaskRole(ctx, req.GetTaskId(), caller.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	body := req.GetBody()
	if body == nil || body.GetOtherTaskId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "other_task_id is required")
	}
	change, err := a.Service.AddTaskRelation(ctx, req.GetTaskId(), body.GetOtherTaskId(), caller.UserID, businessID,
		body.GetType(), body.GetDirection())
	if err != nil {
		return nil, relationError(err, "Failed to add relation")
	}
	publishRelationChange(change, businessID, caller.UserID)
	return &flowtaskv2.AddTaskRelationResponse{RelationId: change.RelationID, Message: "Relation added"}, nil
}

func (a *TaskAPIV2) ListTaskRelations(ctx context.Context, req *flowtaskv2.ListTaskRelationsRequest) (*flowtaskv2.ListTaskRelationsResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:view"); err != nil {
		return nil, err
	}
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if !a.Service.CanReadTask(ctx, caller.UserID, businessID, req.GetTaskId()) {
		return nil, onekiterrors.PermissionDenied()
	}
	groups, err := a.Service.ListTaskRelations(ctx, req.GetTaskId(), caller.UserID, businessID)
	if err != nil {
		return nil, relationError(err, "Failed to list relations")
	}
	return &flowtaskv2.ListTaskRelationsResponse{Groups: toRelationGroupsV2(groups)}, nil
}

func (a *TaskAPIV2) RemoveTaskRelation(ctx context.Context, req *flowtaskv2.RemoveTaskRelationRequest) (*flowtaskv2.RemoveTaskRelationResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:edit"); err != nil {
		return nil, err
	}
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if roleErr := a.Service.RequireTaskRole(ctx, req.GetTaskId(), caller.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	change, err := a.Service.RemoveTaskRelation(ctx, req.GetTaskId(), req.GetRelationId(), businessID)
	if err != nil {
		return nil, relationError(err, "Failed to remove relation")
	}
	publishRelationChange(change, businessID, caller.UserID)
	return &flowtaskv2.RemoveTaskRelationResponse{Message: "Relation removed"}, nil
}
