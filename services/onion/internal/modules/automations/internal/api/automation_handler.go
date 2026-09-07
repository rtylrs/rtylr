package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/automations/internal/application"
	"github.com/voxire/onion/internal/modules/automations/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// automationsCacheScope namespaces the per-owner automations list cache.
// crm_automations is written only within this module, so a short TTL plus
// invalidation on every create/update/delete keeps reads consistent.
const automationsCacheScope = "automations"

func invalidateAutomationsCache(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(automationsCacheScope, businessID))
}

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

func (h *ListAutomationsHandler) Serve(ctx context.Context, _ *ListAutomationsRequest) (ListAutomationsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListAutomationsResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireAutomationManager(ctx); err != nil {
		return ListAutomationsResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListAutomationsResponse{}, onekiterrors.PermissionDenied()
	}
	res, cacheErr := cacher.Of[ListAutomationsResponse]().GetOrComputeSingle(
		ctx,
		cacheKeys.Scoped(automationsCacheScope, businessID, "list:http"),
		func(ctx context.Context) (ListAutomationsResponse, error) {
			return h.computeList(ctx, businessID)
		},
		cacher.WithExpiry(ctx, 3*time.Minute),
	)
	if cacheErr == nil {
		return res, nil
	}

	res, computeErr := h.computeList(ctx, businessID)
	if computeErr != nil {
		return ListAutomationsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list automations")
	}
	return res, nil
}

func (h *ListAutomationsHandler) computeList(ctx context.Context, businessID int64) (ListAutomationsResponse, error) {
	items, err := h.Service.List(ctx, businessID)
	if err != nil {
		return ListAutomationsResponse{}, err
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

func (h *CreateAutomationHandler) Validate(_ context.Context, req *CreateAutomationRequest) error {
	if req.Name == "" {
		return onekiterrors.New(http.StatusBadRequest, "Automation name is required")
	}

	if req.TriggerType == "" {
		return onekiterrors.New(http.StatusBadRequest, "Trigger type is required")
	}
	if err := domain.ValidateAutomationDefinition(req.TriggerType, req.Actions); err != nil {
		return onekiterrors.New(http.StatusBadRequest, err.Error())
	}

	return nil
}

func (h *CreateAutomationHandler) Serve(ctx context.Context, req *CreateAutomationRequest) (CreateAutomationResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateAutomationResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireAutomationManager(ctx); err != nil {
		return CreateAutomationResponse{}, err
	}

	triggerConfig := []byte(req.TriggerConfig)
	if triggerConfig == nil {
		triggerConfig = []byte("{}")
	}

	actions, actionErr := domain.NormalizeAutomationActions(req.Actions)
	if actionErr != nil {
		return CreateAutomationResponse{}, onekiterrors.New(http.StatusBadRequest, actionErr.Error())
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateAutomationResponse{}, onekiterrors.PermissionDenied()
	}
	id, err := h.Service.Create(ctx, domain.AutomationInput{
		OwnerID:       device.User.UserID,
		BusinessID:    businessID,
		Name:          req.Name,
		TriggerType:   req.TriggerType,
		TriggerConfig: triggerConfig,
		Actions:       actions,
		IsActive:      req.IsActive,
	})
	if err != nil {
		return CreateAutomationResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create automation")
	}

	invalidateAutomationsCache(ctx, businessID)

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

func (h *UpdateAutomationHandler) Validate(_ context.Context, req *UpdateAutomationRequest) error {
	if req.Name == "" {
		return onekiterrors.New(http.StatusBadRequest, "Automation name is required")
	}
	if req.TriggerType == "" {
		return onekiterrors.New(http.StatusBadRequest, "Trigger type is required")
	}
	if err := domain.ValidateAutomationDefinition(req.TriggerType, req.Actions); err != nil {
		return onekiterrors.New(http.StatusBadRequest, err.Error())
	}

	return nil
}

func (h *UpdateAutomationHandler) Serve(ctx context.Context, req *UpdateAutomationRequest) (UpdateAutomationResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateAutomationResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireAutomationManager(ctx); err != nil {
		return UpdateAutomationResponse{}, err
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil || id <= 0 {
		return UpdateAutomationResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid automation ID")
	}

	triggerConfig := []byte(req.TriggerConfig)
	if triggerConfig == nil {
		triggerConfig = []byte("{}")
	}

	actions, actionErr := domain.NormalizeAutomationActions(req.Actions)
	if actionErr != nil {
		return UpdateAutomationResponse{}, onekiterrors.New(http.StatusBadRequest, actionErr.Error())
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateAutomationResponse{}, onekiterrors.PermissionDenied()
	}
	if err := h.Service.Update(ctx, domain.AutomationUpdateInput{
		ID:            id,
		OwnerID:       device.User.UserID,
		BusinessID:    businessID,
		Name:          req.Name,
		TriggerType:   req.TriggerType,
		TriggerConfig: triggerConfig,
		Actions:       actions,
		IsActive:      req.IsActive,
	}); err != nil {
		if errors.Is(err, domain.ErrAutomationNotFound) {
			return UpdateAutomationResponse{}, onekiterrors.New(http.StatusNotFound, "Automation not found")
		}

		return UpdateAutomationResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update automation")
	}

	invalidateAutomationsCache(ctx, businessID)

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

func (h *DeleteAutomationHandler) Serve(ctx context.Context, _ *DeleteAutomationRequest) (DeleteAutomationResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteAutomationResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireAutomationManager(ctx); err != nil {
		return DeleteAutomationResponse{}, err
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil || id <= 0 {
		return DeleteAutomationResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid automation ID")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteAutomationResponse{}, onekiterrors.PermissionDenied()
	}
	if err := h.Service.Delete(ctx, businessID, id); err != nil {
		if errors.Is(err, domain.ErrAutomationNotFound) {
			return DeleteAutomationResponse{}, onekiterrors.New(http.StatusNotFound, "Automation not found")
		}

		return DeleteAutomationResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to delete automation")
	}

	invalidateAutomationsCache(ctx, businessID)

	return DeleteAutomationResponse{Message: "Automation deleted successfully"}, nil
}
