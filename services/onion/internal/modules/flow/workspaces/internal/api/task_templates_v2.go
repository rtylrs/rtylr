package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/platform/onekiterrors"
	flowtaskv2 "github.com/voxire/voxie/gen/go/flow/task/v2"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

func toTaskTemplateV2(t domain.TaskTemplate) *flowtaskv2.TaskTemplate {
	return &flowtaskv2.TaskTemplate{
		Id:        t.ID,
		ProjectId: t.ProjectID,
		TaskType:  t.TaskType,
		Body:      t.Body,
		UpdatedAt: t.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

func taskTemplateError(ctx context.Context, operation string, err error) error {
	switch {
	case errors.Is(err, domain.ErrUnauthorized):
		return onekiterrors.PermissionDenied()
	case errors.Is(err, domain.ErrInvalidTaskType):
		return onekiterrors.New(http.StatusBadRequest, domain.ErrInvalidTaskType.Error())
	case errors.Is(err, domain.ErrTaskTemplateTooLong):
		return onekiterrors.New(http.StatusBadRequest, domain.ErrTaskTemplateTooLong.Error())
	}
	return workspaceInternalError(ctx, operation+" failed", err)
}

func (a *TaskAPIV2) ListTaskTemplates(ctx context.Context, req *flowtaskv2.ListTaskTemplatesRequest) (*flowtaskv2.ListTaskTemplatesResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:view"); err != nil {
		return nil, err
	}
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	templates, err := a.Service.ListTaskTemplates(ctx, businessID, req.GetProjectId(), caller.UserID)
	if err != nil {
		return nil, taskTemplateError(ctx, "ListTaskTemplates", err)
	}
	out := make([]*flowtaskv2.TaskTemplate, 0, len(templates))
	for _, t := range templates {
		out = append(out, toTaskTemplateV2(t))
	}
	return &flowtaskv2.ListTaskTemplatesResponse{Templates: out}, nil
}

func (a *TaskAPIV2) SetTaskTemplate(ctx context.Context, req *flowtaskv2.SetTaskTemplateRequest) (*flowtaskv2.SetTaskTemplateResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:edit"); err != nil {
		return nil, err
	}
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	saved, err := a.Service.SetTaskTemplate(ctx, businessID, req.GetProjectId(), caller.UserID, req.GetTaskType(), req.GetBody())
	if err != nil {
		return nil, taskTemplateError(ctx, "SetTaskTemplate", err)
	}
	out := &flowtaskv2.SetTaskTemplateResponse{}
	if saved != nil {
		out.Template = toTaskTemplateV2(*saved)
	}
	return out, nil
}
