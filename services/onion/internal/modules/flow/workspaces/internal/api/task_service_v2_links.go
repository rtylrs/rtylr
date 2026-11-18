package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	flowtaskv2 "github.com/voxire/voxie/gen/go/flow/task/v2"
)

func toTaskLinkV2(l domain.TaskLink) *flowtaskv2.TaskLink {
	return &flowtaskv2.TaskLink{
		Id: l.ID, TaskId: l.TaskID, Url: l.URL, Kind: l.Kind, Title: l.Title,
		Ref: l.Ref, State: l.State, AddedBy: l.AddedBy,
		CreatedAt: l.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

func (a *TaskAPIV2) CreateTaskLink(ctx context.Context, req *flowtaskv2.CreateTaskLinkRequest) (*flowtaskv2.CreateTaskLinkResponse, error) {
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if roleErr := a.Service.RequireTaskRole(ctx, req.GetTaskId(), caller.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	link, err := a.Service.CreateTaskLink(ctx, req.GetTaskId(), businessID, caller.UserID, req.GetUrl())
	switch {
	case errors.Is(err, domain.ErrInvalidTaskLink):
		return nil, onekiterrors.New(http.StatusBadRequest, domain.ErrInvalidTaskLink.Error())
	case errors.Is(err, domain.ErrUnauthorized):
		return nil, onekiterrors.PermissionDenied()
	case err != nil:
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to add link")
	}
	return &flowtaskv2.CreateTaskLinkResponse{Link: toTaskLinkV2(link)}, nil
}

func (a *TaskAPIV2) ListTaskLinks(ctx context.Context, req *flowtaskv2.ListTaskLinksRequest) (*flowtaskv2.ListTaskLinksResponse, error) {
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if !a.Service.CanReadTask(ctx, caller.UserID, businessID, req.GetTaskId()) {
		return nil, onekiterrors.PermissionDenied()
	}
	links, err := a.Service.ListTaskLinks(ctx, req.GetTaskId(), businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get links")
	}
	out := make([]*flowtaskv2.TaskLink, 0, len(links))
	for _, l := range links {
		out = append(out, toTaskLinkV2(l))
	}
	return &flowtaskv2.ListTaskLinksResponse{Links: out}, nil
}

// DeleteTaskLink lets anyone who can edit the task remove any of its links.
func (a *TaskAPIV2) DeleteTaskLink(ctx context.Context, req *flowtaskv2.DeleteTaskLinkRequest) (*flowtaskv2.DeleteTaskLinkResponse, error) {
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	taskID, err := a.Service.TaskLinkTaskID(ctx, req.GetId(), businessID)
	if err != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if roleErr := a.Service.RequireTaskRole(ctx, taskID, caller.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if err := a.Service.DeleteTaskLink(ctx, req.GetId(), businessID, caller.UserID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to remove link")
	}
	return &flowtaskv2.DeleteTaskLinkResponse{Message: "Link removed"}, nil
}
