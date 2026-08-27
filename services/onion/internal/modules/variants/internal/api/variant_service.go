package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/variants/internal/application"
	"github.com/voxire/onion/internal/modules/variants/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	erpvariantv1 "github.com/voxire/voxie/gen/go/erp/variant/v1"
)

type VariantAPI struct {
	Service *application.Service
}

func NewVariantAPI(service *application.Service) *VariantAPI {
	return &VariantAPI{Service: service}
}

func (a *VariantAPI) ListVariants(ctx context.Context, req *erpvariantv1.ListVariantsRequest) (*erpvariantv1.ListVariantsResponse, error) {
	ownerID, err := variantOwnerID(ctx)
	if err != nil {
		return nil, err
	}
	variants, listErr := a.Service.ListVariants(ctx, ownerID, req.GetProductId())
	if listErr != nil {
		return nil, variantStatusError(http.StatusInternalServerError, "Failed to list variants")
	}
	out := make([]*erpvariantv1.Variant, 0, len(variants))
	for _, v := range variants {
		out = append(out, newVariantProto(v))
	}
	return &erpvariantv1.ListVariantsResponse{Variants: out}, nil
}

func (a *VariantAPI) CreateVariant(ctx context.Context, req *erpvariantv1.CreateVariantRequest) (*erpvariantv1.CreateVariantResponse, error) {
	ownerID, err := variantOwnerID(ctx)
	if err != nil {
		return nil, err
	}
	id, createErr := a.Service.CreateVariant(ctx, ownerID, domain.VariantInput{
		ProductID:     req.GetProductId(),
		Name:          strings.TrimSpace(req.GetName()),
		SKU:           req.GetSku(),
		Barcode:       req.GetBarcode(),
		CostPrice:     req.GetCostPrice(),
		SalePrice:     req.GetSalePrice(),
		StockQuantity: req.GetStockQuantity(),
		Attributes:    req.GetAttributes(),
		IsActive:      req.GetIsActive(),
	})
	if createErr != nil {
		return nil, variantStatusError(http.StatusInternalServerError, "Failed to create variant")
	}
	return &erpvariantv1.CreateVariantResponse{VariantId: id, Message: "Variant created successfully"}, nil
}

func (a *VariantAPI) UpdateVariant(ctx context.Context, req *erpvariantv1.UpdateVariantRequest) (*erpvariantv1.UpdateVariantResponse, error) {
	ownerID, err := variantOwnerID(ctx)
	if err != nil {
		return nil, err
	}
	body := req.GetBody()
	if body == nil {
		return nil, variantStatusError(http.StatusBadRequest, "Variant body is required")
	}
	if updateErr := a.Service.UpdateVariant(ctx, ownerID, req.GetId(), domain.VariantInput{
		ProductID:     body.GetProductId(),
		Name:          strings.TrimSpace(body.GetName()),
		SKU:           body.GetSku(),
		Barcode:       body.GetBarcode(),
		CostPrice:     body.GetCostPrice(),
		SalePrice:     body.GetSalePrice(),
		StockQuantity: body.GetStockQuantity(),
		Attributes:    body.GetAttributes(),
		IsActive:      body.GetIsActive(),
	}); updateErr != nil {
		return nil, variantStatusError(http.StatusInternalServerError, "Failed to update variant")
	}
	return &erpvariantv1.UpdateVariantResponse{Message: "Variant updated successfully"}, nil
}

func (a *VariantAPI) DeleteVariant(ctx context.Context, req *erpvariantv1.DeleteVariantRequest) (*erpvariantv1.DeleteVariantResponse, error) {
	ownerID, err := variantOwnerID(ctx)
	if err != nil {
		return nil, err
	}
	if deleteErr := a.Service.DeleteVariant(ctx, ownerID, req.GetId()); deleteErr != nil {
		return nil, variantStatusError(http.StatusInternalServerError, "Failed to delete variant")
	}
	return &erpvariantv1.DeleteVariantResponse{Message: "Variant deleted successfully"}, nil
}

func newVariantProto(v domain.Variant) *erpvariantv1.Variant {
	return &erpvariantv1.Variant{
		Id:            v.ID,
		OwnerId:       v.OwnerID,
		BusinessId:    v.BusinessID,
		ProductId:     &v.ProductID,
		Name:          v.Name,
		Sku:           v.SKU,
		Barcode:       v.Barcode,
		CostPrice:     v.CostPrice,
		SalePrice:     v.SalePrice,
		StockQuantity: v.StockQuantity,
		Attributes:    v.Attributes,
		IsActive:      v.IsActive,
		CreatedAt:     v.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     v.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func variantOwnerID(ctx context.Context) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return 0, variantStatusError(http.StatusUnauthorized, "unauthorized")
	}
	ownerID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return ownerID, nil
}

func variantStatusError(status int, message string) error {
	return onekiterrors.New(status, message)
}
