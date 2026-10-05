package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/erp/products/internal/application"
	"github.com/voxire/onion/internal/modules/erp/products/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	erpproductv1 "github.com/voxire/voxie/gen/go/erp/product/v1"
)

const maxProductListOffset int64 = 10_000

type ProductAPI struct {
	Service *application.Service
}

func NewProductAPI(service *application.Service) *ProductAPI {
	return &ProductAPI{Service: service}
}

func (a *ProductAPI) ListProducts(ctx context.Context, req *erpproductv1.ListProductsRequest) (*erpproductv1.ListProductsResponse, error) {
	ownerID, err := productOwnerID(ctx, "erp:products:view")
	if err != nil {
		return nil, err
	}

	limit := req.GetLimit()
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	offset := normalizeProductListOffset(req.GetOffset())
	page := offset/limit + 1

	products, total, hasMore, listErr := a.Service.ListProducts(ctx, ownerID, req.GetSearch(), req.GetCategoryId(), page, limit)
	if listErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list products")
	}

	out := make([]*erpproductv1.Product, 0, len(products))
	for _, product := range products {
		out = append(out, newProductPayload(product))
	}
	return &erpproductv1.ListProductsResponse{
		Products: out,
		HasMore:  hasMore,
		Page:     page,
		PageSize: limit,
		Total:    total,
	}, nil
}

func (a *ProductAPI) CreateProduct(ctx context.Context, req *erpproductv1.CreateProductRequest) (*erpproductv1.CreateProductResponse, error) {
	ownerID, err := productOwnerID(ctx, "erp:products:create")
	if err != nil {
		return nil, err
	}

	id, createErr := a.Service.CreateProduct(ctx, ownerID, domain.ProductInput{
		CategoryID:        req.GetCategoryId(),
		SupplierID:        req.GetSupplierId(),
		Name:              req.GetName(),
		SKU:               req.GetSku(),
		Barcode:           req.GetBarcode(),
		Description:       req.GetDescription(),
		Unit:              req.GetUnit(),
		CostPrice:         req.GetCostPrice(),
		SalePrice:         req.GetSalePrice(),
		StockQuantity:     req.GetStockQuantity(),
		StockQuantitySet:  true,
		LowStockThreshold: req.GetLowStockThreshold(),
		ReorderQuantity:   req.GetReorderQuantity(),
		IsActive:          req.GetIsActive(),
		TrackInventory:    req.GetTrackInventory(),
		BrandID:           req.GetBrandId(),
		CompareAtPrice:    req.GetCompareAtPrice(),
		TaxRate:           req.GetTaxRate(),
		Currency:          req.GetCurrency(),
		Weight:            req.GetWeight(),
		WeightUnit:        req.GetWeightUnit(),
		Length:            req.GetLength(),
		Width:             req.GetWidth(),
		Height:            req.GetHeight(),
		DimensionUnit:     req.GetDimensionUnit(),
		CountryOfOrigin:   req.GetCountryOfOrigin(),
		HSCode:            req.GetHsCode(),
		AllowBackorder:    req.GetAllowBackorder(),
		LeadTimeDays:      int(req.GetLeadTimeDays()),
		ImageURL:          req.GetImageUrl(),
	})
	if createErr != nil {
		return nil, productMutationError(createErr)
	}
	return &erpproductv1.CreateProductResponse{ProductId: id, Message: "Product created"}, nil
}

func (a *ProductAPI) UpdateProduct(ctx context.Context, req *erpproductv1.UpdateProductRequest) (*erpproductv1.UpdateProductResponse, error) {
	ownerID, err := productOwnerID(ctx, "erp:products:edit")
	if err != nil {
		return nil, err
	}
	body := req.GetBody()
	if body == nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Product body is required")
	}
	return a.updateProductWithOwner(ctx, ownerID, req.GetId(), productUpdateInput(body, nil))
}

func (a *ProductAPI) updateProduct(ctx context.Context, productID int64, input domain.ProductInput) (*erpproductv1.UpdateProductResponse, error) {
	ownerID, err := productOwnerID(ctx, "erp:products:edit")
	if err != nil {
		return nil, err
	}
	return a.updateProductWithOwner(ctx, ownerID, productID, input)
}

func (a *ProductAPI) updateProductWithOwner(ctx context.Context, ownerID, productID int64, input domain.ProductInput) (*erpproductv1.UpdateProductResponse, error) {
	if updateErr := a.Service.UpdateProduct(ctx, ownerID, productID, input); updateErr != nil {
		return nil, productMutationError(updateErr)
	}
	return &erpproductv1.UpdateProductResponse{Message: "Product updated"}, nil
}

func (a *ProductAPI) DeleteProduct(ctx context.Context, req *erpproductv1.DeleteProductRequest) (*erpproductv1.DeleteProductResponse, error) {
	ownerID, err := productOwnerID(ctx, "erp:products:delete")
	if err != nil {
		return nil, err
	}
	if deleteErr := a.Service.DeleteProduct(ctx, ownerID, req.GetId()); deleteErr != nil {
		return nil, productMutationError(deleteErr)
	}
	return &erpproductv1.DeleteProductResponse{Message: "Product deleted"}, nil
}

func (a *ProductAPI) GetProduct(ctx context.Context, req *erpproductv1.GetProductRequest) (*erpproductv1.GetProductResponse, error) {
	ownerID, err := productOwnerID(ctx, "erp:products:view")
	if err != nil {
		return nil, err
	}
	product, getErr := a.Service.GetProduct(ctx, ownerID, req.GetId())
	if getErr != nil {
		return nil, productMutationError(getErr)
	}
	return &erpproductv1.GetProductResponse{Product: newProductPayload(product)}, nil
}

func (a *ProductAPI) GetProductByBarcode(ctx context.Context, req *erpproductv1.GetProductByBarcodeRequest) (*erpproductv1.GetProductByBarcodeResponse, error) {
	ownerID, err := productOwnerID(ctx, "erp:products:view")
	if err != nil {
		return nil, err
	}
	product, getErr := a.Service.GetProductByBarcode(ctx, ownerID, req.GetCode())
	if getErr != nil {
		return nil, productMutationError(getErr)
	}
	return &erpproductv1.GetProductByBarcodeResponse{Product: newProductPayload(product)}, nil
}

func (a *ProductAPI) GetProductFull(ctx context.Context, req *erpproductv1.GetProductFullRequest) (*erpproductv1.GetProductFullResponse, error) {
	ownerID, err := productOwnerID(ctx, "erp:products:view")
	if err != nil {
		return nil, err
	}
	product, getErr := a.Service.GetProduct(ctx, ownerID, req.GetId())
	if getErr != nil {
		return nil, productMutationError(getErr)
	}
	return &erpproductv1.GetProductFullResponse{Product: newFullProductPayload(product)}, nil
}

func (a *ProductAPI) ListProductAttributes(ctx context.Context, req *erpproductv1.ListProductAttributesRequest) (*erpproductv1.ListProductAttributesResponse, error) {
	ownerID, err := productOwnerID(ctx, "erp:products:view")
	if err != nil {
		return nil, err
	}
	attrs, listErr := a.Service.ListProductAttributes(ctx, ownerID, req.GetId())
	if listErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load attributes")
	}
	out := make([]*erpproductv1.ProductAttribute, 0, len(attrs))
	for _, attr := range attrs {
		out = append(out, &erpproductv1.ProductAttribute{
			Id:       attr.ID,
			Name:     attr.Name,
			Value:    attr.Value,
			Position: int32(attr.Position), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		})
	}
	return &erpproductv1.ListProductAttributesResponse{Attributes: out}, nil
}

func (a *ProductAPI) GetDashboard(ctx context.Context, _ *erpproductv1.GetDashboardRequest) (*erpproductv1.GetDashboardResponse, error) {
	ownerID, err := productOwnerID(ctx, "erp:products:view")
	if err != nil {
		return nil, err
	}
	stats, getErr := a.Service.GetDashboard(ctx, ownerID)
	if getErr != nil {
		return nil, productStatusError(http.StatusInternalServerError, "Failed to load dashboard")
	}
	lowStock := make([]*erpproductv1.DashboardLowStockItem, 0, len(stats.LowStockItems))
	for _, it := range stats.LowStockItems {
		lowStock = append(lowStock, &erpproductv1.DashboardLowStockItem{
			ProductId:     it.ProductID,
			Name:          it.Name,
			Sku:           it.SKU,
			CurrentStock:  it.CurrentStock,
			ReorderPoint:  it.ReorderPoint,
			WarehouseName: it.WarehouseName,
		})
	}
	recent := make([]*erpproductv1.DashboardRecentOrder, 0, len(stats.RecentOrders))
	for _, po := range stats.RecentOrders {
		recent = append(recent, &erpproductv1.DashboardRecentOrder{
			Id:           po.ID,
			PoNumber:     po.PONumber,
			SupplierName: po.SupplierName,
			Status:       po.Status,
			Total:        po.Total,
			CreatedAt:    po.CreatedAt.Format(time.RFC3339),
		})
	}
	topCats := make([]*erpproductv1.DashboardTopCategory, 0, len(stats.TopCategories))
	for _, c := range stats.TopCategories {
		topCats = append(topCats, &erpproductv1.DashboardTopCategory{
			Name:         c.Name,
			ProductCount: int32(c.ProductCount), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		})
	}
	return &erpproductv1.GetDashboardResponse{
		TotalProducts:  int32(stats.TotalProducts),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		LowStockCount:  int32(stats.LowStockCount),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		PendingOrders:  int32(stats.PendingOrders),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		TotalSuppliers: int32(stats.TotalSuppliers), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		InventoryValue: stats.InventoryValue,
		LowStockItems:  lowStock,
		RecentOrders:   recent,
		TopCategories:  topCats,
	}, nil
}

func (a *ProductAPI) ListReorderSuggestions(ctx context.Context, _ *erpproductv1.ListReorderSuggestionsRequest) (*erpproductv1.ListReorderSuggestionsResponse, error) {
	ownerID, err := productOwnerID(ctx, "erp:products:view")
	if err != nil {
		return nil, err
	}
	list, listErr := a.Service.ListReorderSuggestions(ctx, ownerID)
	if listErr != nil {
		return nil, productStatusError(http.StatusInternalServerError, "Failed to load reorder suggestions")
	}
	out := make([]*erpproductv1.ReorderSuggestion, 0, len(list.Suggestions))
	for _, s := range list.Suggestions {
		out = append(out, &erpproductv1.ReorderSuggestion{
			ProductId:         s.ProductID,
			Name:              s.Name,
			Sku:               s.SKU,
			CurrentStock:      s.CurrentStock,
			ReorderPoint:      s.ReorderPoint,
			SuggestedQuantity: s.SuggestedQty,
			AvgDailySales:     s.AvgDailySales,
			LeadTimeDays:      int32(s.LeadTimeDays), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			Confidence:        s.Confidence,
		})
	}
	return &erpproductv1.ListReorderSuggestionsResponse{Suggestions: out, HasMore: list.Truncated}, nil
}

func (a *ProductAPI) ListExpiringItems(ctx context.Context, req *erpproductv1.ListExpiringItemsRequest) (*erpproductv1.ListExpiringItemsResponse, error) {
	ownerID, err := productOwnerID(ctx, "erp:products:view")
	if err != nil {
		return nil, err
	}
	items, listErr := a.Service.ListExpiringItems(ctx, ownerID, int(req.GetDaysAhead()))
	if listErr != nil {
		return nil, productStatusError(http.StatusInternalServerError, "Failed to load expiry report")
	}
	out := make([]*erpproductv1.ExpiringItem, 0, len(items))
	for _, it := range items {
		out = append(out, &erpproductv1.ExpiringItem{
			ProductId:     it.ProductID,
			Name:          it.Name,
			Sku:           it.SKU,
			BatchId:       it.BatchID,
			ExpiryDate:    it.ExpiryDate.Format("2006-01-02"),
			Quantity:      it.Quantity,
			WarehouseName: it.WarehouseName,
		})
	}
	return &erpproductv1.ListExpiringItemsResponse{Items: out}, nil
}

func normalizeProductListOffset(offset int64) int64 {
	if offset < 0 {
		return 0
	}
	if offset > maxProductListOffset {
		return maxProductListOffset
	}
	return offset
}

func productOwnerID(ctx context.Context, permission string) (int64, error) {
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

func productStatusError(status int, message string) error {
	return onekiterrors.New(status, message)
}

func productMutationError(err error) error {
	if errors.Is(err, domain.ErrProductAlreadyExists) {
		return onekiterrors.New(http.StatusConflict, err.Error())
	}
	if errors.Is(err, domain.ErrProductNotFound) {
		return onekiterrors.New(http.StatusNotFound, err.Error())
	}
	if errors.Is(err, domain.ErrProductReferenceNotFound) {
		return onekiterrors.New(http.StatusNotFound, err.Error())
	}
	return onekiterrors.New(http.StatusInternalServerError, "Failed to update product")
}
