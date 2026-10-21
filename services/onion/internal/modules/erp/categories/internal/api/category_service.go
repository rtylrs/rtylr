package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/erp/categories/internal/application"
	"github.com/voxire/onion/internal/modules/erp/categories/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	erpcategoryv1 "github.com/voxire/voxie/gen/go/erp/category/v1"
)

type CategoryAPI struct {
	Service *application.Service
}

func NewCategoryAPI(service *application.Service) *CategoryAPI {
	return &CategoryAPI{Service: service}
}

func (a *CategoryAPI) ListCategories(ctx context.Context, req *erpcategoryv1.ListCategoriesRequest) (*erpcategoryv1.ListCategoriesResponse, error) {
	ownerID, err := categoryOwnerID(ctx, "erp:products:view")
	if err != nil {
		return nil, err
	}
	page, listErr := a.Service.ListCategoriesPage(ctx, ownerID, req.GetLimit(), req.GetOffset())
	if listErr != nil {
		return nil, categoryStatusError(http.StatusInternalServerError, "Failed to list categories")
	}
	out := make([]*erpcategoryv1.Category, 0, len(page.Categories))
	for _, c := range page.Categories {
		out = append(out, newCategoryPayload(c))
	}
	return &erpcategoryv1.ListCategoriesResponse{
		Categories: out,
		HasMore:    page.Offset+int64(len(out)) < page.Total,
		Limit:      page.Limit,
		Offset:     page.Offset,
		Total:      page.Total,
	}, nil
}

func (a *CategoryAPI) CreateCategory(ctx context.Context, req *erpcategoryv1.CreateCategoryRequest) (*erpcategoryv1.CreateCategoryResponse, error) {
	ownerID, err := categoryOwnerID(ctx, "erp:products:create")
	if err != nil {
		return nil, err
	}
	id, createErr := a.Service.CreateCategory(ctx, ownerID, domain.CategoryInput{
		Name:        strings.TrimSpace(req.GetName()),
		Description: req.GetDescription(),
	})
	if createErr != nil {
		return nil, categoryStatusError(http.StatusInternalServerError, "Failed to create category")
	}
	return &erpcategoryv1.CreateCategoryResponse{CategoryId: id, Message: "Category created successfully"}, nil
}

func (a *CategoryAPI) UpdateCategory(ctx context.Context, req *erpcategoryv1.UpdateCategoryRequest) (*erpcategoryv1.UpdateCategoryResponse, error) {
	ownerID, err := categoryOwnerID(ctx, "erp:products:edit")
	if err != nil {
		return nil, err
	}
	body := req.GetBody()
	if body == nil {
		return nil, categoryStatusError(http.StatusBadRequest, "Category body is required")
	}
	if updateErr := a.Service.UpdateCategory(ctx, ownerID, req.GetId(), domain.CategoryInput{
		Name:        strings.TrimSpace(body.GetName()),
		Description: body.GetDescription(),
	}); updateErr != nil {
		return nil, categoryStatusError(http.StatusInternalServerError, "Failed to update category")
	}
	return &erpcategoryv1.UpdateCategoryResponse{Message: "Category updated successfully"}, nil
}

func (a *CategoryAPI) DeleteCategory(ctx context.Context, req *erpcategoryv1.DeleteCategoryRequest) (*erpcategoryv1.DeleteCategoryResponse, error) {
	ownerID, err := categoryOwnerID(ctx, "erp:products:delete")
	if err != nil {
		return nil, err
	}
	if deleteErr := a.Service.DeleteCategory(ctx, ownerID, req.GetId()); deleteErr != nil {
		return nil, categoryStatusError(http.StatusInternalServerError, "Failed to delete category")
	}
	return &erpcategoryv1.DeleteCategoryResponse{Message: "Category deleted successfully"}, nil
}

func categoryOwnerID(ctx context.Context, permission string) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || !authPkg.HasAppPermission(authUser, authorization.AppERP, permission) {
		return 0, onekiterrors.PermissionDenied()
	}
	ownerID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return ownerID, nil
}

func categoryStatusError(status int, message string) error {
	return onekiterrors.New(status, message)
}
