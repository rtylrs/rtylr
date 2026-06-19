package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/crm_sales/internal/application"
	"github.com/voxire/onion/internal/modules/crm_sales/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
)

// ---------------------------------------------------------------------------
// GET /crm/sales/pipeline
// ---------------------------------------------------------------------------

func GetPipelineRoutes() []string { return []string{"/crm/sales/pipeline"} }

type (
	GetPipelineHandler  struct{ Service *application.Service }
	GetPipelineRequest  struct{}
	GetPipelineResponse struct {
		Pipeline []domain.PipelineSummary `json:"pipeline"`
	}
)

func (h *GetPipelineHandler) Serve(ctx context.Context, _ *GetPipelineRequest) (GetPipelineResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetPipelineResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetPipelineResponse{}, errorModels.PermissionDeniedServerError()
	}

	summaries, err := h.Service.GetPipeline(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to get pipeline", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return GetPipelineResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to get pipeline"}
	}

	return GetPipelineResponse{Pipeline: summaries}, nil
}

// ---------------------------------------------------------------------------
// GET /crm/sales/leads
// ---------------------------------------------------------------------------

func ListLeadsRoutes() []string { return []string{"/crm/sales/leads"} }

type (
	ListLeadsHandler  struct{ Service *application.Service }
	ListLeadsRequest  struct{}
	ListLeadsResponse struct {
		Leads []domain.Lead `json:"leads"`
	}
)

func (h *ListLeadsHandler) Serve(ctx context.Context, _ *ListLeadsRequest) (ListLeadsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListLeadsResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListLeadsResponse{}, errorModels.PermissionDeniedServerError()
	}

	leads, err := h.Service.ListLeads(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list leads", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ListLeadsResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to list leads"}
	}

	return ListLeadsResponse{Leads: leads}, nil
}

// ---------------------------------------------------------------------------
// POST /crm/sales/leads
// ---------------------------------------------------------------------------

func CreateLeadRoutes() []string { return []string{"/crm/sales/leads"} }

type CreateLeadHandler struct{ Service *application.Service }

type CreateLeadRequest struct {
	CompanyName  string           `json:"company_name"`
	Industry     string           `json:"industry"`
	Size         string           `json:"size"`
	Source       string           `json:"source"`
	Stage        domain.LeadStage `json:"stage"`
	AssignedTo   *int64           `json:"assigned_to"`
	MRRPotential *float64         `json:"mrr_potential"`
	Notes        string           `json:"notes"`
}

type CreateLeadResponse struct {
	Lead domain.Lead `json:"lead"`
}

func (h *CreateLeadHandler) Validate(_ context.Context, req *CreateLeadRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.CompanyName) == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "company_name is required"}
	}
	return nil
}

func (h *CreateLeadHandler) Serve(ctx context.Context, req *CreateLeadRequest) (CreateLeadResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateLeadResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateLeadResponse{}, errorModels.PermissionDeniedServerError()
	}

	lead, err := h.Service.CreateLead(ctx, businessID, domain.CreateLeadRequest{
		CompanyName:  req.CompanyName,
		Industry:     req.Industry,
		Size:         req.Size,
		Source:       req.Source,
		Stage:        req.Stage,
		AssignedTo:   req.AssignedTo,
		MRRPotential: req.MRRPotential,
		Notes:        req.Notes,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to create lead", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return CreateLeadResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to create lead"}
	}

	return CreateLeadResponse{Lead: lead}, nil
}

// ---------------------------------------------------------------------------
// GET /crm/sales/leads/{id}
// ---------------------------------------------------------------------------

func GetLeadRoutes() []string { return []string{"/crm/sales/leads/{id}"} }

type (
	GetLeadHandler  struct{ Service *application.Service }
	GetLeadRequest  struct{}
	GetLeadResponse struct {
		Lead domain.Lead `json:"lead"`
	}
)

func (h *GetLeadHandler) Serve(ctx context.Context, _ *GetLeadRequest) (GetLeadResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetLeadResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return GetLeadResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid lead id"}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetLeadResponse{}, errorModels.PermissionDeniedServerError()
	}

	lead, err := h.Service.GetLead(ctx, id, businessID)
	if err != nil {
		if errors.Is(err, domain.ErrLeadNotFound) {
			return GetLeadResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Lead not found"}
		}
		logger.Ctx(ctx).Error("failed to get lead", logger.Any("owner_id", device.User.UserID), logger.Any("lead_id", id), logger.Error(err))
		return GetLeadResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to get lead"}
	}

	return GetLeadResponse{Lead: lead}, nil
}

// ---------------------------------------------------------------------------
// PUT /crm/sales/leads/{id}
// ---------------------------------------------------------------------------

func UpdateLeadRoutes() []string { return []string{"/crm/sales/leads/{id}"} }

type UpdateLeadHandler struct{ Service *application.Service }

type UpdateLeadRequest struct {
	CompanyName  string           `json:"company_name"`
	Industry     string           `json:"industry"`
	Size         string           `json:"size"`
	Source       string           `json:"source"`
	Stage        domain.LeadStage `json:"stage"`
	AssignedTo   *int64           `json:"assigned_to"`
	MRRPotential *float64         `json:"mrr_potential"`
	Notes        string           `json:"notes"`
}

type UpdateLeadResponse struct {
	Lead domain.Lead `json:"lead"`
}

func (h *UpdateLeadHandler) Validate(_ context.Context, req *UpdateLeadRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.CompanyName) == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "company_name is required"}
	}
	return nil
}

func (h *UpdateLeadHandler) Serve(ctx context.Context, req *UpdateLeadRequest) (UpdateLeadResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateLeadResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return UpdateLeadResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid lead id"}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateLeadResponse{}, errorModels.PermissionDeniedServerError()
	}

	lead, err := h.Service.UpdateLead(ctx, id, businessID, domain.UpdateLeadRequest{
		CompanyName:  req.CompanyName,
		Industry:     req.Industry,
		Size:         req.Size,
		Source:       req.Source,
		Stage:        req.Stage,
		AssignedTo:   req.AssignedTo,
		MRRPotential: req.MRRPotential,
		Notes:        req.Notes,
	})
	if err != nil {
		if errors.Is(err, domain.ErrLeadNotFound) {
			return UpdateLeadResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Lead not found"}
		}
		logger.Ctx(ctx).Error("failed to update lead", logger.Any("owner_id", device.User.UserID), logger.Any("lead_id", id), logger.Error(err))
		return UpdateLeadResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to update lead"}
	}

	return UpdateLeadResponse{Lead: lead}, nil
}

// ---------------------------------------------------------------------------
// DELETE /crm/sales/leads/{id}
// ---------------------------------------------------------------------------

func DeleteLeadRoutes() []string { return []string{"/crm/sales/leads/{id}"} }

type (
	DeleteLeadHandler  struct{ Service *application.Service }
	DeleteLeadRequest  struct{}
	DeleteLeadResponse struct {
		Success bool `json:"success"`
	}
)

func (h *DeleteLeadHandler) Serve(ctx context.Context, _ *DeleteLeadRequest) (DeleteLeadResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteLeadResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return DeleteLeadResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid lead id"}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteLeadResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.DeleteLead(ctx, id, businessID); err != nil {
		if errors.Is(err, domain.ErrLeadNotFound) {
			return DeleteLeadResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Lead not found"}
		}
		logger.Ctx(ctx).Error("failed to delete lead", logger.Any("owner_id", device.User.UserID), logger.Any("lead_id", id), logger.Error(err))
		return DeleteLeadResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to delete lead"}
	}

	return DeleteLeadResponse{Success: true}, nil
}
