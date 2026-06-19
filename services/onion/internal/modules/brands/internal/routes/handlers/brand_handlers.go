package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/brands/internal/application"
	"github.com/voxire/onion/internal/modules/brands/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// --- List Brands -------------------------------------------------------------

func ListBrandsRoutes() []string {
	return []string{"/erp/brands"}
}

type ListBrandsHandler struct {
	Service *application.Service
}

type ListBrandsRequest struct{}

type BrandPayload struct {
	ID           int64  `json:"id"`
	OwnerID      int64  `json:"owner_id"`
	Name         string `json:"name"`
	LogoURL      string `json:"logo_url"`
	Description  string `json:"description"`
	ProductCount int64  `json:"product_count"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type ListBrandsResponse struct {
	Brands []BrandPayload `json:"brands"`
}

func newBrandPayload(b domain.Brand) BrandPayload {
	return BrandPayload{
		ID:           b.ID,
		OwnerID:      b.OwnerID,
		Name:         b.Name,
		LogoURL:      b.LogoURL,
		Description:  b.Description,
		ProductCount: b.ProductCount,
		CreatedAt:    b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    b.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *ListBrandsHandler) Serve(ctx context.Context, _ *ListBrandsRequest) (ListBrandsResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListBrandsResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return ListBrandsResponse{}, errorModels.PermissionDeniedServerError()
	}

	brands, err := h.Service.ListBrands(ctx, businessID)
	if err != nil {
		return ListBrandsResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to list brands",
		}
	}

	payloads := make([]BrandPayload, 0, len(brands))
	for _, b := range brands {
		payloads = append(payloads, newBrandPayload(b))
	}

	return ListBrandsResponse{Brands: payloads}, nil
}

// --- Create Brand ------------------------------------------------------------

func CreateBrandRoutes() []string {
	return []string{"/erp/brands"}
}

type CreateBrandHandler struct {
	Service *application.Service
}

type CreateBrandRequest struct {
	Name        string `json:"name"`
	LogoURL     string `json:"logo_url"`
	Description string `json:"description"`
}

type CreateBrandResponse struct {
	BrandID int64  `json:"brand_id"`
	Message string `json:"message"`
}

func (h *CreateBrandHandler) Validate(_ context.Context, req *CreateBrandRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.Name) == "" {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Brand name is required",
		}
	}

	return nil
}

func (h *CreateBrandHandler) Serve(ctx context.Context, req *CreateBrandRequest) (CreateBrandResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateBrandResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	input := domain.BrandInput{
		Name:        req.Name,
		LogoURL:     req.LogoURL,
		Description: req.Description,
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return CreateBrandResponse{}, errorModels.PermissionDeniedServerError()
	}

	id, err := h.Service.CreateBrand(ctx, businessID, input)
	if err != nil {
		return CreateBrandResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to create brand",
		}
	}

	return CreateBrandResponse{BrandID: id, Message: "Brand created successfully"}, nil
}

// --- Update Brand ------------------------------------------------------------

func UpdateBrandRoutes() []string {
	return []string{"/erp/brands"}
}

type UpdateBrandHandler struct {
	Service *application.Service
}

type UpdateBrandRequest struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	LogoURL     string `json:"logo_url"`
	Description string `json:"description"`
}

type UpdateBrandResponse struct {
	Message string `json:"message"`
}

func (h *UpdateBrandHandler) Validate(_ context.Context, req *UpdateBrandRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Brand ID is required",
		}
	}

	if strings.TrimSpace(req.Name) == "" {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Brand name is required",
		}
	}

	return nil
}

func (h *UpdateBrandHandler) Serve(ctx context.Context, req *UpdateBrandRequest) (UpdateBrandResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateBrandResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	input := domain.BrandInput{
		Name:        req.Name,
		LogoURL:     req.LogoURL,
		Description: req.Description,
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return UpdateBrandResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.UpdateBrand(ctx, businessID, req.ID, input); err != nil {
		return UpdateBrandResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to update brand",
		}
	}

	return UpdateBrandResponse{Message: "Brand updated successfully"}, nil
}

// --- Delete Brand ------------------------------------------------------------

func DeleteBrandRoutes() []string {
	return []string{"/erp/brands"}
}

type DeleteBrandHandler struct {
	Service *application.Service
}

type DeleteBrandRequest struct {
	ID int64 `json:"id"`
}

type DeleteBrandResponse struct {
	Message string `json:"message"`
}

func (h *DeleteBrandHandler) Validate(_ context.Context, req *DeleteBrandRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Brand ID is required",
		}
	}

	return nil
}

func (h *DeleteBrandHandler) Serve(ctx context.Context, req *DeleteBrandRequest) (DeleteBrandResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return DeleteBrandResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return DeleteBrandResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.DeleteBrand(ctx, businessID, req.ID); err != nil {
		return DeleteBrandResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to delete brand",
		}
	}

	return DeleteBrandResponse{Message: "Brand deleted successfully"}, nil
}
