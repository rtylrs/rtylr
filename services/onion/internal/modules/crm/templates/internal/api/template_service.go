package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/crm/templates/internal/application"
	"github.com/voxire/onion/internal/modules/crm/templates/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	crmtemplatev1 "github.com/voxire/voxie/gen/go/crm/template/v1"
)

type TemplateAPI struct {
	service *application.Service
}

func NewTemplateAPI(service *application.Service) *TemplateAPI {
	return &TemplateAPI{service: service}
}

func (a *TemplateAPI) ListTemplates(ctx context.Context, _ *crmtemplatev1.ListTemplatesRequest) (*crmtemplatev1.ListTemplatesResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	items, err := a.service.List(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list templates")
	}

	templates := make([]*crmtemplatev1.Template, len(items))
	for i, item := range items {
		templates[i] = newTemplate(item)
	}

	return &crmtemplatev1.ListTemplatesResponse{Templates: templates}, nil
}

func (a *TemplateAPI) CreateTemplate(ctx context.Context, req *crmtemplatev1.CreateTemplateRequest) (*crmtemplatev1.CreateTemplateResponse, error) {
	if err := requireTemplateManager(ctx); err != nil {
		return nil, err
	}

	if err := validateTemplateFields(req.GetName(), req.GetType(), req.GetBody()); err != nil {
		return nil, err
	}

	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	variables, err := variablesToBytes(req.GetVariables())
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Template variables must be valid JSON")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	id, err := a.service.Create(ctx, domain.TemplateInput{
		OwnerID:    device.User.UserID,
		BusinessID: businessID,
		Name:       req.GetName(),
		Type:       req.GetType(),
		Subject:    req.GetSubject(),
		Body:       req.GetBody(),
		Variables:  variables,
		IsActive:   req.GetIsActive(),
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create template")
	}

	return &crmtemplatev1.CreateTemplateResponse{TemplateId: id, Message: "Template created successfully"}, nil
}

func (a *TemplateAPI) UpdateTemplate(ctx context.Context, req *crmtemplatev1.UpdateTemplateRequest) (*crmtemplatev1.UpdateTemplateResponse, error) {
	if err := requireTemplateManager(ctx); err != nil {
		return nil, err
	}

	if req.GetId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Template ID is required")
	}

	body := req.GetBody()
	if body == nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Template name is required")
	}
	// Same field validation as create: updates previously skipped the
	// type/body checks and could persist unusable templates.
	if err := validateTemplateFields(body.GetName(), body.GetType(), body.GetBody()); err != nil {
		return nil, err
	}

	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	variables, err := variablesToBytes(body.GetVariables())
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Template variables must be valid JSON")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	err = a.service.Update(ctx, domain.TemplateUpdateInput{
		ID:         req.GetId(),
		OwnerID:    device.User.UserID,
		BusinessID: businessID,
		Name:       body.GetName(),
		Type:       body.GetType(),
		Subject:    body.GetSubject(),
		Body:       body.GetBody(),
		Variables:  variables,
		IsActive:   body.GetIsActive(),
	})
	if err != nil {
		if errors.Is(err, domain.ErrTemplateNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Template not found")
		}

		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update template")
	}

	return &crmtemplatev1.UpdateTemplateResponse{Message: "Template updated successfully"}, nil
}

func (a *TemplateAPI) DeleteTemplate(ctx context.Context, req *crmtemplatev1.DeleteTemplateRequest) (*crmtemplatev1.DeleteTemplateResponse, error) {
	if err := requireTemplateManager(ctx); err != nil {
		return nil, err
	}

	if req.GetId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Template ID is required")
	}

	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.service.Delete(ctx, businessID, req.GetId()); err != nil {
		if errors.Is(err, domain.ErrTemplateNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Template not found")
		}

		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete template")
	}

	return &crmtemplatev1.DeleteTemplateResponse{Message: "Template deleted successfully"}, nil
}

func validateTemplateFields(name, templateType, body string) error {
	if strings.TrimSpace(name) == "" {
		return onekiterrors.New(http.StatusBadRequest, "Template name is required")
	}
	if strings.TrimSpace(templateType) == "" {
		return onekiterrors.New(http.StatusBadRequest, "Template type is required")
	}
	if strings.TrimSpace(body) == "" {
		return onekiterrors.New(http.StatusBadRequest, "Template body is required")
	}
	return nil
}

// requireTemplateManager mirrors the HTTP layer's requireCRMManager gate:
// creating/updating/deleting shared tenant templates must not be available to
// every authenticated member role.
func requireTemplateManager(ctx context.Context) error {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return onekiterrors.PermissionDenied()
	}
	if authPkg.IsBusinessAdmin(device.User) || authPkg.HasAppPermission(device.User, authorization.AppCRM, "crm:templates:manage") {
		return nil
	}
	return onekiterrors.PermissionDenied()
}

func newTemplate(template domain.Template) *crmtemplatev1.Template {
	return &crmtemplatev1.Template{
		Id:        template.ID,
		OwnerId:   template.OwnerID,
		Name:      template.Name,
		Type:      template.Type,
		Subject:   template.Subject,
		Body:      template.Body,
		Variables: bytesToVariables(template.Variables),
		IsActive:  template.IsActive,
		CreatedAt: template.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: template.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func variablesToBytes(value string) ([]byte, error) {
	if value == "" {
		return []byte("[]"), nil
	}
	var probe any
	if err := json.Unmarshal([]byte(value), &probe); err != nil {
		return nil, err
	}
	return []byte(value), nil
}

func bytesToVariables(raw []byte) string {
	if len(raw) == 0 {
		return "[]"
	}
	return string(raw)
}
