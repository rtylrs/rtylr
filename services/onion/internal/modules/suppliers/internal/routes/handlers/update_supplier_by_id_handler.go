package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/suppliers/internal/application"
	"github.com/voxire/onion/internal/modules/suppliers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// ─── Update Supplier by Path ID ───────────────────────────────────────────────
//
// PUT /erp/suppliers/{id} — REST path variant that coexists with the existing
// PUT /erp/suppliers (body-ID) route. Reuses Service.UpdateSupplier.

func UpdateSupplierByIDRoutes() []string {
	return []string{"/erp/suppliers/{id}"}
}

type UpdateSupplierByIDHandler struct {
	Service *application.Service
}

type UpdateSupplierByIDRequest struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Address      string `json:"address"`
	City         string `json:"city"`
	Country      string `json:"country"`
	TaxID        string `json:"tax_id"`
	PaymentTerms string `json:"payment_terms"`
	Notes        string `json:"notes"`
	IsActive     bool   `json:"is_active"`
}

type UpdateSupplierByIDResponse struct {
	Message string `json:"message"`
}

func (h *UpdateSupplierByIDHandler) Validate(_ context.Context, req *UpdateSupplierByIDRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Supplier name is required",
		}
	}
	return nil
}

func (h *UpdateSupplierByIDHandler) Serve(ctx context.Context, req *UpdateSupplierByIDRequest) (UpdateSupplierByIDResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateSupplierByIDResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	idRaw := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil || id <= 0 {
		return UpdateSupplierByIDResponse{}, &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Invalid supplier id",
		}
	}

	input := domain.SupplierInput{
		Name:         req.Name,
		Email:        req.Email,
		Phone:        req.Phone,
		Address:      req.Address,
		City:         req.City,
		Country:      req.Country,
		TaxID:        req.TaxID,
		PaymentTerms: req.PaymentTerms,
		Notes:        req.Notes,
		IsActive:     req.IsActive,
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return UpdateSupplierByIDResponse{}, errorModels.PermissionDeniedServerError()
	}
	if err := h.Service.UpdateSupplier(ctx, businessID, id, input); err != nil {
		if errors.Is(err, domain.ErrSupplierNotFound) {
			return UpdateSupplierByIDResponse{}, &errorModels.ServerError{
				Code:     http.StatusNotFound,
				HTTPCode: http.StatusNotFound,
				Message:  "Supplier not found",
			}
		}
		return UpdateSupplierByIDResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to update supplier",
		}
	}

	return UpdateSupplierByIDResponse{Message: "Supplier updated successfully"}, nil
}

// ─── Delete Supplier by Path ID ───────────────────────────────────────────────
//
// DELETE /erp/suppliers/{id} — new REST path variant. Reuses Service.DeleteSupplier.

func DeleteSupplierByIDRoutes() []string {
	return []string{"/erp/suppliers/{id}"}
}

type DeleteSupplierByIDHandler struct {
	Service *application.Service
}

type DeleteSupplierByIDRequest struct{}

type DeleteSupplierByIDResponse struct {
	Message string `json:"message"`
}

func (h *DeleteSupplierByIDHandler) Serve(ctx context.Context, _ *DeleteSupplierByIDRequest) (DeleteSupplierByIDResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return DeleteSupplierByIDResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	idRaw := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil || id <= 0 {
		return DeleteSupplierByIDResponse{}, &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Invalid supplier id",
		}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return DeleteSupplierByIDResponse{}, errorModels.PermissionDeniedServerError()
	}
	if err := h.Service.DeleteSupplier(ctx, businessID, id); err != nil {
		if errors.Is(err, domain.ErrSupplierNotFound) {
			return DeleteSupplierByIDResponse{}, &errorModels.ServerError{
				Code:     http.StatusNotFound,
				HTTPCode: http.StatusNotFound,
				Message:  "Supplier not found",
			}
		}
		return DeleteSupplierByIDResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to delete supplier",
		}
	}

	return DeleteSupplierByIDResponse{Message: "Supplier deleted successfully"}, nil
}
