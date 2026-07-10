package api

import (
	"context"
	"net/http"
	"strings"

	erpbrandv1 "github.com/voxire/voxie/gen/go/erp/brand/v1"

	"github.com/voxire/onion/internal/modules/brands/internal/application"
	"github.com/voxire/onion/internal/modules/brands/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// listBrandsPageSize bounds ListBrands server-side. The voxie
// erp.brand.v1.ListBrandsRequest contract has no pagination fields yet, so
// the endpoint returns at most one maximum-size page instead of every row.
const listBrandsPageSize int32 = 100

type BrandAPI struct {
	Service *application.Service
}

func NewBrandAPI(service *application.Service) *BrandAPI {
	return &BrandAPI{Service: service}
}

func (a *BrandAPI) ListBrands(ctx context.Context, _ *erpbrandv1.ListBrandsRequest) (*erpbrandv1.ListBrandsResponse, error) {
	ownerID, err := brandOwnerID(ctx)
	if err != nil {
		return nil, err
	}
	brands, listErr := a.Service.ListBrandsPaginated(ctx, ownerID, listBrandsPageSize, 0)
	if listErr != nil {
		return nil, brandStatusError(http.StatusInternalServerError, "Failed to list brands")
	}
	out := make([]*erpbrandv1.Brand, 0, len(brands))
	for _, b := range brands {
		out = append(out, newBrandPayload(b))
	}
	return &erpbrandv1.ListBrandsResponse{Brands: out}, nil
}

func (a *BrandAPI) CreateBrand(ctx context.Context, req *erpbrandv1.CreateBrandRequest) (*erpbrandv1.CreateBrandResponse, error) {
	ownerID, err := brandOwnerID(ctx)
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
	ownerID, err := brandOwnerID(ctx)
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
	ownerID, err := brandOwnerID(ctx)
	if err != nil {
		return nil, err
	}
	if deleteErr := a.Service.DeleteBrand(ctx, ownerID, req.GetId()); deleteErr != nil {
		return nil, brandStatusError(http.StatusInternalServerError, "Failed to delete brand")
	}
	return &erpbrandv1.DeleteBrandResponse{Message: "Brand deleted successfully"}, nil
}

func brandOwnerID(ctx context.Context) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return 0, brandStatusError(http.StatusUnauthorized, "unauthorized")
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
