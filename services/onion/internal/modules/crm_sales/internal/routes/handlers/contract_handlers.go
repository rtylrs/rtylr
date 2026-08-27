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
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/money"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// ---------------------------------------------------------------------------
// GET /crm/sales/contracts
// ---------------------------------------------------------------------------

func ListContractsRoutes() []string { return []string{"/crm/sales/contracts"} }

type ListContractsHandler struct{ Service *application.Service }

type ListContractsRequest struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ListContractsResponse struct {
	Contracts []domain.Contract `json:"contracts"`
	Total     int64             `json:"total"`
	Limit     int32             `json:"limit"`
	Offset    int32             `json:"offset"`
}

func (h *ListContractsHandler) Serve(ctx context.Context, req *ListContractsRequest) (ListContractsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListContractsResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireCRMFeaturePermission(ctx, "crm:customers:view"); err != nil {
		return ListContractsResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListContractsResponse{}, onekiterrors.PermissionDenied()
	}

	const maxContractPageSize int32 = 100
	if req.Limit <= 0 || req.Limit > maxContractPageSize {
		req.Limit = maxContractPageSize
	}
	if req.Offset < 0 {
		req.Offset = 0
	}
	page, err := h.Service.ListContracts(ctx, businessID, req.Limit, req.Offset)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list contracts", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ListContractsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list contracts")
	}

	return ListContractsResponse{Contracts: page.Contracts, Total: page.Total, Limit: page.Limit, Offset: page.Offset}, nil
}

// ---------------------------------------------------------------------------
// POST /crm/sales/contracts
// ---------------------------------------------------------------------------

func CreateContractRoutes() []string { return []string{"/crm/sales/contracts"} }

type CreateContractHandler struct{ Service *application.Service }

type CreateContractRequest struct {
	LeadID       int64        `json:"lead_id"`
	Plan         string       `json:"plan"`
	MRR          money.Amount `json:"mrr"`
	Seats        int          `json:"seats"`
	SignedAt     *string      `json:"signed_at"`
	SignedByName string       `json:"signed_by_name"`
	ClosedBy     *int64       `json:"closed_by"`
	Notes        string       `json:"notes"`
}

type CreateContractResponse struct {
	Contract domain.Contract `json:"contract"`
}

func (h *CreateContractHandler) Validate(_ context.Context, req *CreateContractRequest) error {
	if req.LeadID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "lead_id is required")
	}
	if req.MRR < 0 {
		return onekiterrors.New(http.StatusBadRequest, "mrr must not be negative")
	}
	return nil
}

func (h *CreateContractHandler) Serve(ctx context.Context, req *CreateContractRequest) (CreateContractResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateContractResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireCRMFeaturePermission(ctx, "crm:customers:create"); err != nil {
		return CreateContractResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateContractResponse{}, onekiterrors.PermissionDenied()
	}

	contract, err := h.Service.CreateContract(ctx, businessID, domain.CreateContractRequest{
		LeadID:       req.LeadID,
		Plan:         req.Plan,
		MRR:          req.MRR,
		Seats:        req.Seats,
		SignedAt:     req.SignedAt,
		SignedByName: req.SignedByName,
		ClosedBy:     req.ClosedBy,
		Notes:        req.Notes,
	})
	if err != nil {
		if errors.Is(err, domain.ErrLeadNotFound) {
			return CreateContractResponse{}, onekiterrors.New(http.StatusNotFound, "Lead not found")
		}
		if errors.Is(err, domain.ErrInvalidLeadTransition) {
			return CreateContractResponse{}, onekiterrors.New(http.StatusConflict, "Lead cannot be contracted from its current stage")
		}
		logger.Ctx(ctx).Error("failed to create contract", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return CreateContractResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create contract")
	}

	return CreateContractResponse{Contract: contract}, nil
}

// ---------------------------------------------------------------------------
// GET /crm/sales/contracts/{id}
// ---------------------------------------------------------------------------

func GetContractRoutes() []string { return []string{"/crm/sales/contracts/{id}"} }

type (
	GetContractHandler  struct{ Service *application.Service }
	GetContractRequest  struct{}
	GetContractResponse struct {
		Contract domain.Contract `json:"contract"`
	}
)

func (h *GetContractHandler) Serve(ctx context.Context, _ *GetContractRequest) (GetContractResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetContractResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireCRMFeaturePermission(ctx, "crm:customers:view"); err != nil {
		return GetContractResponse{}, err
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return GetContractResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid contract id")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetContractResponse{}, onekiterrors.PermissionDenied()
	}

	contract, err := h.Service.GetContract(ctx, id, businessID)
	if err != nil {
		if errors.Is(err, domain.ErrContractNotFound) {
			return GetContractResponse{}, onekiterrors.New(http.StatusNotFound, "Contract not found")
		}
		logger.Ctx(ctx).Error("failed to get contract", logger.Any("owner_id", device.User.UserID), logger.Any("contract_id", id), logger.Error(err))
		return GetContractResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get contract")
	}

	return GetContractResponse{Contract: contract}, nil
}

// ---------------------------------------------------------------------------
// PUT /crm/sales/contracts/{id}
// ---------------------------------------------------------------------------

func UpdateContractRoutes() []string { return []string{"/crm/sales/contracts/{id}"} }

type UpdateContractHandler struct{ Service *application.Service }

type UpdateContractRequest struct {
	Plan         string       `json:"plan"`
	MRR          money.Amount `json:"mrr"`
	Seats        int          `json:"seats"`
	SignedAt     *string      `json:"signed_at"`
	SignedByName string       `json:"signed_by_name"`
	ClosedBy     *int64       `json:"closed_by"`
	Notes        string       `json:"notes"`
}

type UpdateContractResponse struct {
	Contract domain.Contract `json:"contract"`
}

func (h *UpdateContractHandler) Validate(_ context.Context, req *UpdateContractRequest) error {
	if strings.TrimSpace(req.Plan) == "" {
		return onekiterrors.New(http.StatusBadRequest, "plan is required")
	}
	if req.MRR < 0 {
		return onekiterrors.New(http.StatusBadRequest, "mrr must not be negative")
	}
	return nil
}

func (h *UpdateContractHandler) Serve(ctx context.Context, req *UpdateContractRequest) (UpdateContractResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateContractResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireCRMFeaturePermission(ctx, "crm:customers:edit"); err != nil {
		return UpdateContractResponse{}, err
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return UpdateContractResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid contract id")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateContractResponse{}, onekiterrors.PermissionDenied()
	}

	contract, err := h.Service.UpdateContract(ctx, id, businessID, domain.UpdateContractRequest{
		Plan:         req.Plan,
		MRR:          req.MRR,
		Seats:        req.Seats,
		SignedAt:     req.SignedAt,
		SignedByName: req.SignedByName,
		ClosedBy:     req.ClosedBy,
		Notes:        req.Notes,
	})
	if err != nil {
		if errors.Is(err, domain.ErrContractNotFound) {
			return UpdateContractResponse{}, onekiterrors.New(http.StatusNotFound, "Contract not found")
		}
		logger.Ctx(ctx).Error("failed to update contract", logger.Any("owner_id", device.User.UserID), logger.Any("contract_id", id), logger.Error(err))
		return UpdateContractResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update contract")
	}

	return UpdateContractResponse{Contract: contract}, nil
}
