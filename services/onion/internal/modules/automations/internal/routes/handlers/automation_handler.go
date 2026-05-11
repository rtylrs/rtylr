package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/automations/internal/application"
	"github.com/voxire/onion/internal/modules/automations/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func AutomationListRoutes() []string {
	return []string{"/crm/automations"}
}

func AutomationCreateRoutes() []string {
	return []string{"/crm/automations"}
}

func AutomationUpdateRoutes() []string {
	return []string{"/crm/automations/{id}"}
}

func AutomationDeleteRoutes() []string {
	return []string{"/crm/automations/{id}"}
}

type AutomationPayload struct {
	ID            int64           `json:"id"`
	OwnerID       int64           `json:"owner_id"`
	Name          string          `json:"name"`
	TriggerType   string          `json:"trigger_type"`
	TriggerConfig json.RawMessage `json:"trigger_config"`
	Actions       json.RawMessage `json:"actions"`
	IsActive      bool            `json:"is_active"`
	LastRunAt     *string         `json:"last_run_at"`
	RunCount      int             `json:"run_count"`
	CreatedAt     string          `json:"created_at"`
	UpdatedAt     string          `json:"updated_at"`
}

func newAutomationPayload(a domain.Automation) AutomationPayload {
	p := AutomationPayload{
		ID:            a.ID,
		OwnerID:       a.OwnerID,
		Name:          a.Name,
		TriggerType:   a.TriggerType,
		TriggerConfig: json.RawMessage(a.TriggerConfig),
		Actions:       json.RawMessage(a.Actions),
		IsActive:      a.IsActive,
		RunCount:      a.RunCount,
		CreatedAt:     a.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     a.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if a.LastRunAt != nil {
		s := a.LastRunAt.Format("2006-01-02T15:04:05Z")
		p.LastRunAt = &s
	}

	return p
}

// -- LIST --------------------------------------------------------------------

type ListAutomationsHandler struct {
	Service *application.Service
}

type ListAutomationsRequest struct{}

type ListAutomationsResponse struct {
	Automations []AutomationPayload `json:"automations"`
}

func (h *ListAutomationsHandler) Serve(ctx context.Context, _ *ListAutomationsRequest) (ListAutomationsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListAutomationsResponse{}, errorModels.PermissionDeniedServerError()
	}

	items, err := h.Service.List(ctx, device.User.UserID)
	if err != nil {
		return ListAutomationsResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list automations",
		}
	}

	payloads := make([]AutomationPayload, len(items))
	for i, a := range items {
		payloads[i] = newAutomationPayload(a)
	}

	return ListAutomationsResponse{Automations: payloads}, nil
}

// -- CREATE ------------------------------------------------------------------

type CreateAutomationHandler struct {
	Service *application.Service
}

type CreateAutomationRequest struct {
	Name          string          `json:"name"`
	TriggerType   string          `json:"trigger_type"`
	TriggerConfig json.RawMessage `json:"trigger_config"`
	Actions       json.RawMessage `json:"actions"`
	IsActive      bool            `json:"is_active"`
}

type CreateAutomationResponse struct {
	AutomationID int64  `json:"automation_id"`
	Message      string `json:"message"`
}

func (h *CreateAutomationHandler) Validate(_ context.Context, req *CreateAutomationRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Automation name is required",
		}
	}

	if req.TriggerType == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Trigger type is required",
		}
	}

	return nil
}

func (h *CreateAutomationHandler) Serve(ctx context.Context, req *CreateAutomationRequest) (CreateAutomationResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateAutomationResponse{}, errorModels.PermissionDeniedServerError()
	}

	triggerConfig := []byte(req.TriggerConfig)
	if triggerConfig == nil {
		triggerConfig = []byte("{}")
	}

	actions := []byte(req.Actions)
	if actions == nil {
		actions = []byte("[]")
	}

	id, err := h.Service.Create(ctx, domain.AutomationInput{
		OwnerID:       device.User.UserID,
		Name:          req.Name,
		TriggerType:   req.TriggerType,
		TriggerConfig: triggerConfig,
		Actions:       actions,
		IsActive:      req.IsActive,
	})
	if err != nil {
		return CreateAutomationResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create automation",
		}
	}

	return CreateAutomationResponse{AutomationID: id, Message: "Automation created successfully"}, nil
}

// -- UPDATE ------------------------------------------------------------------

type UpdateAutomationHandler struct {
	Service *application.Service
}

type UpdateAutomationRequest struct {
	Name          string          `json:"name"`
	TriggerType   string          `json:"trigger_type"`
	TriggerConfig json.RawMessage `json:"trigger_config"`
	Actions       json.RawMessage `json:"actions"`
	IsActive      bool            `json:"is_active"`
}

type UpdateAutomationResponse struct {
	Message string `json:"message"`
}

func (h *UpdateAutomationHandler) Validate(_ context.Context, req *UpdateAutomationRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Automation name is required",
		}
	}

	return nil
}

func (h *UpdateAutomationHandler) Serve(ctx context.Context, req *UpdateAutomationRequest) (UpdateAutomationResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateAutomationResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil || id <= 0 {
		return UpdateAutomationResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid automation ID",
		}
	}

	triggerConfig := []byte(req.TriggerConfig)
	if triggerConfig == nil {
		triggerConfig = []byte("{}")
	}

	actions := []byte(req.Actions)
	if actions == nil {
		actions = []byte("[]")
	}

	if err := h.Service.Update(ctx, domain.AutomationUpdateInput{
		ID:            id,
		OwnerID:       device.User.UserID,
		Name:          req.Name,
		TriggerType:   req.TriggerType,
		TriggerConfig: triggerConfig,
		Actions:       actions,
		IsActive:      req.IsActive,
	}); err != nil {
		if errors.Is(err, domain.ErrAutomationNotFound) {
			return UpdateAutomationResponse{}, &errorModels.ServerError{
				Code:    http.StatusNotFound,
				Message: "Automation not found",
			}
		}

		return UpdateAutomationResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update automation",
		}
	}

	return UpdateAutomationResponse{Message: "Automation updated successfully"}, nil
}

// -- DELETE ------------------------------------------------------------------

type DeleteAutomationHandler struct {
	Service *application.Service
}

type DeleteAutomationRequest struct{}

type DeleteAutomationResponse struct {
	Message string `json:"message"`
}

func (h *DeleteAutomationHandler) Serve(ctx context.Context, _ *DeleteAutomationRequest) (DeleteAutomationResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteAutomationResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil || id <= 0 {
		return DeleteAutomationResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid automation ID",
		}
	}

	if err := h.Service.Delete(ctx, device.User.UserID, id); err != nil {
		if errors.Is(err, domain.ErrAutomationNotFound) {
			return DeleteAutomationResponse{}, &errorModels.ServerError{
				Code:    http.StatusNotFound,
				Message: "Automation not found",
			}
		}

		return DeleteAutomationResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete automation",
		}
	}

	return DeleteAutomationResponse{Message: "Automation deleted successfully"}, nil
}
