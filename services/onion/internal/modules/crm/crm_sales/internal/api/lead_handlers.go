package api

import (
	"context"
	"errors"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	"net/http"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/application"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/money"
	"github.com/voxire/onion/internal/platform/onekiterrors"
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
		Pipeline []pipelineSummaryPayload `json:"pipeline"`
	}
)

func (h *GetPipelineHandler) Serve(ctx context.Context, _ *GetPipelineRequest) (GetPipelineResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetPipelineResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireCRMFeaturePermission(ctx, "crm:customers:view"); err != nil {
		return GetPipelineResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetPipelineResponse{}, onekiterrors.PermissionDenied()
	}

	summaries, err := h.Service.GetPipeline(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to get pipeline", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return GetPipelineResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get pipeline")
	}

	return GetPipelineResponse{Pipeline: payloadPipelineSummaryList(summaries)}, nil
}

// ---------------------------------------------------------------------------
// GET /crm/sales/leads
// ---------------------------------------------------------------------------

func ListLeadsRoutes() []string { return []string{"/crm/sales/leads"} }

type ListLeadsHandler struct{ Service *application.Service }

type ListLeadsRequest struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ListLeadsResponse struct {
	Leads  []leadPayload `json:"leads"`
	Total  int64         `json:"total"`
	Limit  int32         `json:"limit"`
	Offset int32         `json:"offset"`
}

func (h *ListLeadsHandler) Serve(ctx context.Context, req *ListLeadsRequest) (ListLeadsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListLeadsResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireCRMFeaturePermission(ctx, "crm:customers:view"); err != nil {
		return ListLeadsResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListLeadsResponse{}, onekiterrors.PermissionDenied()
	}

	page, err := h.Service.ListLeads(ctx, businessID, req.Limit, req.Offset)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list leads", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ListLeadsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list leads")
	}

	return ListLeadsResponse{Leads: payloadLeadList(page.Leads), Total: page.Total, Limit: page.Limit, Offset: page.Offset}, nil
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
	MRRPotential *money.Amount    `json:"mrr_potential"`
	Notes        string           `json:"notes"`
}

type CreateLeadResponse struct {
	Lead leadPayload `json:"lead"`
}

func (h *CreateLeadHandler) Validate(_ context.Context, req *CreateLeadRequest) error {
	if strings.TrimSpace(req.CompanyName) == "" {
		return onekiterrors.New(http.StatusBadRequest, "company_name is required")
	}
	return nil
}

func (h *CreateLeadHandler) Serve(ctx context.Context, req *CreateLeadRequest) (CreateLeadResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateLeadResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireCRMFeaturePermission(ctx, "crm:customers:create"); err != nil {
		return CreateLeadResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateLeadResponse{}, onekiterrors.PermissionDenied()
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
		if errors.Is(err, domain.ErrInvalidLeadStage) {
			return CreateLeadResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid lead stage")
		}
		if errors.Is(err, domain.ErrInvalidLeadAssignee) {
			return CreateLeadResponse{}, onekiterrors.New(http.StatusBadRequest, "Assigned user is not a member of this business")
		}
		logger.Ctx(ctx).Error("failed to create lead", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return CreateLeadResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create lead")
	}

	return CreateLeadResponse{Lead: payloadLead(lead)}, nil
}

// ---------------------------------------------------------------------------
// GET /crm/sales/leads/{id}
// ---------------------------------------------------------------------------

func GetLeadRoutes() []string { return []string{"/crm/sales/leads/{id}"} }

type (
	GetLeadHandler  struct{ Service *application.Service }
	GetLeadRequest  struct{}
	GetLeadResponse struct {
		Lead leadPayload `json:"lead"`
	}
)

func (h *GetLeadHandler) Serve(ctx context.Context, _ *GetLeadRequest) (GetLeadResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetLeadResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireCRMFeaturePermission(ctx, "crm:customers:view"); err != nil {
		return GetLeadResponse{}, err
	}

	idStr := serverHttp.PathParam(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return GetLeadResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid lead id")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetLeadResponse{}, onekiterrors.PermissionDenied()
	}

	lead, err := h.Service.GetLead(ctx, id, businessID)
	if err != nil {
		if errors.Is(err, domain.ErrLeadNotFound) {
			return GetLeadResponse{}, onekiterrors.New(http.StatusNotFound, "Lead not found")
		}
		logger.Ctx(ctx).Error("failed to get lead", logger.Any("owner_id", device.User.UserID), logger.Any("lead_id", id), logger.Error(err))
		return GetLeadResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get lead")
	}

	return GetLeadResponse{Lead: payloadLead(lead)}, nil
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
	MRRPotential *money.Amount    `json:"mrr_potential"`
	Notes        string           `json:"notes"`
}

type UpdateLeadResponse struct {
	Lead leadPayload `json:"lead"`
}

func (h *UpdateLeadHandler) Validate(_ context.Context, req *UpdateLeadRequest) error {
	return nil
}

func (h *UpdateLeadHandler) Serve(ctx context.Context, req *UpdateLeadRequest) (UpdateLeadResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateLeadResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireCRMFeaturePermission(ctx, "crm:customers:edit"); err != nil {
		return UpdateLeadResponse{}, err
	}

	idStr := serverHttp.PathParam(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return UpdateLeadResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid lead id")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateLeadResponse{}, onekiterrors.PermissionDenied()
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
		if errors.Is(err, domain.ErrInvalidLeadStage) {
			return UpdateLeadResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid lead stage")
		}
		if errors.Is(err, domain.ErrInvalidLeadAssignee) {
			return UpdateLeadResponse{}, onekiterrors.New(http.StatusBadRequest, "Assigned user is not a member of this business")
		}
		if errors.Is(err, domain.ErrInvalidLeadTransition) {
			return UpdateLeadResponse{}, onekiterrors.New(http.StatusConflict, "Lead stage cannot move backwards or leave a terminal stage")
		}
		if errors.Is(err, domain.ErrLeadNotFound) {
			return UpdateLeadResponse{}, onekiterrors.New(http.StatusNotFound, "Lead not found")
		}
		logger.Ctx(ctx).Error("failed to update lead", logger.Any("owner_id", device.User.UserID), logger.Any("lead_id", id), logger.Error(err))
		return UpdateLeadResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update lead")
	}

	return UpdateLeadResponse{Lead: payloadLead(lead)}, nil
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

func (h *DeleteLeadHandler) Serve(ctx context.Context, _ *DeleteLeadRequest) (DeleteLeadResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteLeadResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireCRMFeaturePermission(ctx, "crm:customers:delete"); err != nil {
		return DeleteLeadResponse{}, err
	}

	idStr := serverHttp.PathParam(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return DeleteLeadResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid lead id")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteLeadResponse{}, onekiterrors.PermissionDenied()
	}

	if err := h.Service.DeleteLead(ctx, id, businessID); err != nil {
		if errors.Is(err, domain.ErrLeadNotFound) {
			return DeleteLeadResponse{}, onekiterrors.New(http.StatusNotFound, "Lead not found")
		}
		logger.Ctx(ctx).Error("failed to delete lead", logger.Any("owner_id", device.User.UserID), logger.Any("lead_id", id), logger.Error(err))
		return DeleteLeadResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to delete lead")
	}

	return DeleteLeadResponse{Success: true}, nil
}
