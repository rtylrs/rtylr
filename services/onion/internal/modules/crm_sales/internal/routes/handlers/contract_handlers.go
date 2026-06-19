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
// GET /crm/sales/contracts
// ---------------------------------------------------------------------------

func ListContractsRoutes() []string { return []string{"/crm/sales/contracts"} }

type (
	ListContractsHandler  struct{ Service *application.Service }
	ListContractsRequest  struct{}
	ListContractsResponse struct {
		Contracts []domain.Contract `json:"contracts"`
	}
)

func (h *ListContractsHandler) Serve(ctx context.Context, _ *ListContractsRequest) (ListContractsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListContractsResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListContractsResponse{}, errorModels.PermissionDeniedServerError()
	}

	contracts, err := h.Service.ListContracts(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list contracts", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ListContractsResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to list contracts"}
	}

	return ListContractsResponse{Contracts: contracts}, nil
}

// ---------------------------------------------------------------------------
// POST /crm/sales/contracts
// ---------------------------------------------------------------------------

func CreateContractRoutes() []string { return []string{"/crm/sales/contracts"} }

type CreateContractHandler struct{ Service *application.Service }

type CreateContractRequest struct {
	LeadID       int64   `json:"lead_id"`
	Plan         string  `json:"plan"`
	MRR          float64 `json:"mrr"`
	Seats        int     `json:"seats"`
	SignedAt     *string `json:"signed_at"`
	SignedByName string  `json:"signed_by_name"`
	ClosedBy     *int64  `json:"closed_by"`
	Notes        string  `json:"notes"`
}

type CreateContractResponse struct {
	Contract domain.Contract `json:"contract"`
}

func (h *CreateContractHandler) Validate(_ context.Context, req *CreateContractRequest) *errorModels.ServerError {
	if req.LeadID == 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "lead_id is required"}
	}
	if req.MRR < 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "mrr must not be negative"}
	}
	return nil
}

func (h *CreateContractHandler) Serve(ctx context.Context, req *CreateContractRequest) (CreateContractResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateContractResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateContractResponse{}, errorModels.PermissionDeniedServerError()
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
		logger.Ctx(ctx).Error("failed to create contract", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return CreateContractResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to create contract"}
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

func (h *GetContractHandler) Serve(ctx context.Context, _ *GetContractRequest) (GetContractResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetContractResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return GetContractResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid contract id"}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetContractResponse{}, errorModels.PermissionDeniedServerError()
	}

	contract, err := h.Service.GetContract(ctx, id, businessID)
	if err != nil {
		if errors.Is(err, domain.ErrContractNotFound) {
			return GetContractResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Contract not found"}
		}
		logger.Ctx(ctx).Error("failed to get contract", logger.Any("owner_id", device.User.UserID), logger.Any("contract_id", id), logger.Error(err))
		return GetContractResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to get contract"}
	}

	return GetContractResponse{Contract: contract}, nil
}

// ---------------------------------------------------------------------------
// PUT /crm/sales/contracts/{id}
// ---------------------------------------------------------------------------

func UpdateContractRoutes() []string { return []string{"/crm/sales/contracts/{id}"} }

type UpdateContractHandler struct{ Service *application.Service }

type UpdateContractRequest struct {
	Plan         string  `json:"plan"`
	MRR          float64 `json:"mrr"`
	Seats        int     `json:"seats"`
	SignedAt     *string `json:"signed_at"`
	SignedByName string  `json:"signed_by_name"`
	ClosedBy     *int64  `json:"closed_by"`
	Notes        string  `json:"notes"`
}

type UpdateContractResponse struct {
	Contract domain.Contract `json:"contract"`
}

func (h *UpdateContractHandler) Validate(_ context.Context, req *UpdateContractRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.Plan) == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "plan is required"}
	}
	if req.MRR < 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "mrr must not be negative"}
	}
	return nil
}

func (h *UpdateContractHandler) Serve(ctx context.Context, req *UpdateContractRequest) (UpdateContractResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateContractResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return UpdateContractResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid contract id"}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateContractResponse{}, errorModels.PermissionDeniedServerError()
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
			return UpdateContractResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Contract not found"}
		}
		logger.Ctx(ctx).Error("failed to update contract", logger.Any("owner_id", device.User.UserID), logger.Any("contract_id", id), logger.Error(err))
		return UpdateContractResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to update contract"}
	}

	return UpdateContractResponse{Contract: contract}, nil
}
