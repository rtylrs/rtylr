package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/categories/internal/application"
	"github.com/voxire/onion/internal/modules/categories/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// --- List Categories ---------------------------------------------------------

func ListCategoriesRoutes() []string {
	return []string{"/erp/categories"}
}

type ListCategoriesHandler struct {
	Service *application.Service
}

type ListCategoriesRequest struct{}

type CategoryPayload struct {
	ID          int64  `json:"id"`
	OwnerID     int64  `json:"owner_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type ListCategoriesResponse struct {
	Categories []CategoryPayload `json:"categories"`
}

func newCategoryPayload(c domain.Category) CategoryPayload {
	return CategoryPayload{
		ID:          c.ID,
		OwnerID:     c.OwnerID,
		Name:        c.Name,
		Description: c.Description,
		CreatedAt:   c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *ListCategoriesHandler) Serve(ctx context.Context, _ *ListCategoriesRequest) (ListCategoriesResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListCategoriesResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return ListCategoriesResponse{}, errorModels.PermissionDeniedServerError()
	}

	categories, err := h.Service.ListCategories(ctx, businessID)
	if err != nil {
		return ListCategoriesResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to list categories",
		}
	}

	payloads := make([]CategoryPayload, 0, len(categories))
	for _, c := range categories {
		payloads = append(payloads, newCategoryPayload(c))
	}

	return ListCategoriesResponse{Categories: payloads}, nil
}

// --- Create Category ---------------------------------------------------------

func CreateCategoryRoutes() []string {
	return []string{"/erp/categories"}
}

type CreateCategoryHandler struct {
	Service *application.Service
}

type CreateCategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateCategoryResponse struct {
	CategoryID int64  `json:"category_id"`
	Message    string `json:"message"`
}

func (h *CreateCategoryHandler) Validate(_ context.Context, req *CreateCategoryRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Category name is required",
		}
	}

	return nil
}

func (h *CreateCategoryHandler) Serve(ctx context.Context, req *CreateCategoryRequest) (CreateCategoryResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateCategoryResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	input := domain.CategoryInput{
		Name:        req.Name,
		Description: req.Description,
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return CreateCategoryResponse{}, errorModels.PermissionDeniedServerError()
	}

	id, err := h.Service.CreateCategory(ctx, businessID, input)
	if err != nil {
		return CreateCategoryResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to create category",
		}
	}

	return CreateCategoryResponse{CategoryID: id, Message: "Category created successfully"}, nil
}

// --- Update Category ---------------------------------------------------------

func UpdateCategoryRoutes() []string {
	return []string{"/erp/categories"}
}

type UpdateCategoryHandler struct {
	Service *application.Service
}

type UpdateCategoryRequest struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateCategoryResponse struct {
	Message string `json:"message"`
}

func (h *UpdateCategoryHandler) Validate(_ context.Context, req *UpdateCategoryRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Category ID is required",
		}
	}

	if req.Name == "" {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Category name is required",
		}
	}

	return nil
}

func (h *UpdateCategoryHandler) Serve(ctx context.Context, req *UpdateCategoryRequest) (UpdateCategoryResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateCategoryResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	input := domain.CategoryInput{
		Name:        req.Name,
		Description: req.Description,
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return UpdateCategoryResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.UpdateCategory(ctx, businessID, req.ID, input); err != nil {
		return UpdateCategoryResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to update category",
		}
	}

	return UpdateCategoryResponse{Message: "Category updated successfully"}, nil
}

// --- Delete Category ---------------------------------------------------------

func DeleteCategoryRoutes() []string {
	return []string{"/erp/categories"}
}

type DeleteCategoryHandler struct {
	Service *application.Service
}

type DeleteCategoryRequest struct {
	ID int64 `json:"id"`
}

type DeleteCategoryResponse struct {
	Message string `json:"message"`
}

func (h *DeleteCategoryHandler) Validate(_ context.Context, req *DeleteCategoryRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Category ID is required",
		}
	}

	return nil
}

func (h *DeleteCategoryHandler) Serve(ctx context.Context, req *DeleteCategoryRequest) (DeleteCategoryResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return DeleteCategoryResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return DeleteCategoryResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.DeleteCategory(ctx, businessID, req.ID); err != nil {
		return DeleteCategoryResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to delete category",
		}
	}

	return DeleteCategoryResponse{Message: "Category deleted successfully"}, nil
}
