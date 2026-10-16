package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/erp/brands/internal/application"
	"github.com/voxire/onion/internal/modules/erp/brands/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	erpbrandv1 "github.com/voxire/voxie/gen/go/erp/brand/v1"
)

type BrandAPI struct {
	Service *application.Service
}

func NewBrandAPI(service *application.Service) *BrandAPI {
	return &BrandAPI{Service: service}
}

func (a *BrandAPI) ListBrands(ctx context.Context, req *erpbrandv1.ListBrandsRequest) (*erpbrandv1.ListBrandsResponse, error) {
	ownerID, err := brandOwnerID(ctx, "erp:products:view")
	if err != nil {
		return nil, err
	}
	page, listErr := a.Service.ListBrandsPage(ctx, ownerID, req.GetLimit(), req.GetOffset())
	if listErr != nil {
		return nil, brandStatusError(http.StatusInternalServerError, "Failed to list brands")
	}
	out := make([]*erpbrandv1.Brand, 0, len(page.Brands))
	for _, b := range page.Brands {
		out = append(out, newBrandPayload(b))
	}
	return &erpbrandv1.ListBrandsResponse{
		Brands:  out,
		HasMore: page.Offset+int64(len(out)) < page.Total,
		Limit:   page.Limit,
		Offset:  page.Offset,
		Total:   page.Total,
	}, nil
}

func (a *BrandAPI) CreateBrand(ctx context.Context, req *erpbrandv1.CreateBrandRequest) (*erpbrandv1.CreateBrandResponse, error) {
	ownerID, err := brandOwnerID(ctx, "erp:products:create")
	if err != nil {
		return nil, err
	}
	id, createErr := a.Service.CreateBrand(ctx, ownerID, domain.BrandInput{
		Name:        strings.TrimSpace(req.GetName()),
		LogoURL:     req.GetLogoUrl(),
		Description: req.GetDescription(),
	})
	if createErr != nil {
		return nil, brandStatusError(http.StatusInternalServerError, "Failed to create brand")
	}
	return &erpbrandv1.CreateBrandResponse{BrandId: id, Message: "Brand created successfully"}, nil
}

func (a *BrandAPI) UpdateBrand(ctx context.Context, req *erpbrandv1.UpdateBrandRequest) (*erpbrandv1.UpdateBrandResponse, error) {
	ownerID, err := brandOwnerID(ctx, "erp:products:edit")
	if err != nil {
		return nil, err
	}
	body := req.GetBody()
	if body == nil {
		return nil, brandStatusError(http.StatusBadRequest, "Brand body is required")
	}
	if updateErr := a.Service.UpdateBrand(ctx, ownerID, req.GetId(), domain.BrandInput{
		Name:        strings.TrimSpace(body.GetName()),
		LogoURL:     body.GetLogoUrl(),
		Description: body.GetDescription(),
	}); updateErr != nil {
		return nil, brandStatusError(http.StatusInternalServerError, "Failed to update brand")
	}
	return &erpbrandv1.UpdateBrandResponse{Message: "Brand updated successfully"}, nil
}

func (a *BrandAPI) DeleteBrand(ctx context.Context, req *erpbrandv1.DeleteBrandRequest) (*erpbrandv1.DeleteBrandResponse, error) {
	ownerID, err := brandOwnerID(ctx, "erp:products:delete")
	if err != nil {
		return nil, err
	}
	if deleteErr := a.Service.DeleteBrand(ctx, ownerID, req.GetId()); deleteErr != nil {
		return nil, brandStatusError(http.StatusInternalServerError, "Failed to delete brand")
	}
	return &erpbrandv1.DeleteBrandResponse{Message: "Brand deleted successfully"}, nil
}

func brandOwnerID(ctx context.Context, permission string) (int64, error) {
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

func brandStatusError(status int, message string) error {
	return onekiterrors.New(status, message)
}
