package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/suppliers/internal/application"
	"github.com/voxire/onion/internal/modules/suppliers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
	erpsupplierv1 "github.com/voxire/voxie/gen/go/erp/supplier/v1"
)

// ─── List Suppliers ───────────────────────────────────────────────────────────

func ListSuppliersRoutes() []string {
	return []string{"/erp/suppliers"}
}

type ListSuppliersHandler struct {
	Service *application.Service
}

type ListSuppliersRequest struct {
	Page     int64 `json:"page"`
	PageSize int64 `json:"page_size"`
}

type ListSuppliersResponse struct {
	Suppliers []*erpsupplierv1.Supplier `json:"suppliers"`
	HasMore   bool                      `json:"has_more"`
	Page      int64                     `json:"page"`
	PageSize  int64                     `json:"page_size"`
}

func newSupplierPayload(s domain.Supplier) *erpsupplierv1.Supplier {
	return &erpsupplierv1.Supplier{
		Id:           s.ID,
		OwnerId:      s.OwnerID,
		Name:         s.Name,
		Email:        s.Email,
		Phone:        s.Phone,
		Address:      s.Address,
		City:         s.City,
		Country:      s.Country,
		TaxId:        s.TaxID,
		PaymentTerms: s.PaymentTerms,
		Notes:        s.Notes,
		IsActive:     s.IsActive,
		CreatedAt:    s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    s.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *ListSuppliersHandler) Serve(ctx context.Context, req *ListSuppliersRequest) (ListSuppliersResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListSuppliersResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return ListSuppliersResponse{}, errorModels.PermissionDeniedServerError()
	}
	suppliers, hasMore, err := h.Service.ListSuppliers(ctx, businessID, req.Page, req.PageSize)
	if err != nil {
		return ListSuppliersResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to list suppliers",
		}
	}

	payloads := make([]*erpsupplierv1.Supplier, 0, len(suppliers))
	for _, s := range suppliers {
		payloads = append(payloads, newSupplierPayload(s))
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 50
	}
	return ListSuppliersResponse{Suppliers: payloads, HasMore: hasMore, Page: page, PageSize: pageSize}, nil
}

// ─── Create Supplier ──────────────────────────────────────────────────────────

func CreateSupplierRoutes() []string {
	return []string{"/erp/suppliers"}
}

type CreateSupplierHandler struct {
	Service *application.Service
}

type CreateSupplierRequest struct {
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

type CreateSupplierResponse struct {
	SupplierID int64  `json:"supplier_id"`
	Message    string `json:"message"`
}

func (h *CreateSupplierHandler) Validate(_ context.Context, req *CreateSupplierRequest) *errorModels.ServerError {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Supplier name is required",
		}
	}
	return nil
}

func (h *CreateSupplierHandler) Serve(ctx context.Context, req *CreateSupplierRequest) (CreateSupplierResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateSupplierResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
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
		return CreateSupplierResponse{}, errorModels.PermissionDeniedServerError()
	}
	id, err := h.Service.CreateSupplier(ctx, businessID, input)
	if err != nil {
		return CreateSupplierResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to create supplier",
		}
	}

	return CreateSupplierResponse{SupplierID: id, Message: "Supplier created successfully"}, nil
}

// ─── Update Supplier ──────────────────────────────────────────────────────────

func UpdateSupplierRoutes() []string {
	return []string{"/erp/suppliers"}
}

type UpdateSupplierHandler struct {
	Service *application.Service
}

type UpdateSupplierRequest struct {
	ID           int64  `json:"id"`
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

type UpdateSupplierResponse struct {
	Message string `json:"message"`
}

func (h *UpdateSupplierHandler) Validate(_ context.Context, req *UpdateSupplierRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Supplier ID is required",
		}
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Supplier name is required",
		}
	}
	return nil
}

func (h *UpdateSupplierHandler) Serve(ctx context.Context, req *UpdateSupplierRequest) (UpdateSupplierResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateSupplierResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
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
		return UpdateSupplierResponse{}, errorModels.PermissionDeniedServerError()
	}
	if err := h.Service.UpdateSupplier(ctx, businessID, req.ID, input); err != nil {
		if errors.Is(err, domain.ErrSupplierNotFound) {
			return UpdateSupplierResponse{}, &errorModels.ServerError{
				Code:     http.StatusNotFound,
				HTTPCode: http.StatusNotFound,
				Message:  "Supplier not found",
			}
		}
		return UpdateSupplierResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to update supplier",
		}
	}

	return UpdateSupplierResponse{Message: "Supplier updated successfully"}, nil
}
