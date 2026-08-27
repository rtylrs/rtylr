package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	insightsproductv1 "github.com/voxire/voxie/gen/go/insights/product/v1"
)

type ProductAPI struct {
	Service *application.Service
}

func NewProductAPI(service *application.Service) *ProductAPI {
	return &ProductAPI{Service: service}
}

func (a *ProductAPI) GetBestSellers(ctx context.Context, req *insightsproductv1.GetBestSellersRequest) (*insightsproductv1.GetBestSellersResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:dashboard:view")
	if err != nil {
		return nil, err
	}

	rows, err := a.Service.GetBestSellers(ctx, businessID, normalizeLimit(req.GetLimit()))
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load best sellers")
	}

	products := make([]*insightsproductv1.BestSeller, 0, len(rows))
	for _, row := range rows {
		products = append(products, &insightsproductv1.BestSeller{
			ProductId:         row.ProductID,
			Name:              row.Name,
			UnitsSold:         row.UnitsSold,
			RevenueMinorUnits: row.RevenueMinorUnits,
		})
	}

	return &insightsproductv1.GetBestSellersResponse{Products: products}, nil
}

func (a *ProductAPI) GetCategoryPerformance(ctx context.Context, _ *insightsproductv1.GetCategoryPerformanceRequest) (*insightsproductv1.GetCategoryPerformanceResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:dashboard:view")
	if err != nil {
		return nil, err
	}

	rows, err := a.Service.GetSalesByCategory(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load category performance")
	}

	categories := make([]*insightsproductv1.CategorySales, 0, len(rows))
	for _, row := range rows {
		categories = append(categories, &insightsproductv1.CategorySales{
			CategoryId:        row.CategoryID,
			CategoryName:      row.CategoryName,
			UnitsSold:         row.UnitsSold,
			RevenueMinorUnits: row.RevenueMinorUnits,
		})
	}

	return &insightsproductv1.GetCategoryPerformanceResponse{Categories: categories}, nil
}

func (a *ProductAPI) GetProductMargins(ctx context.Context, _ *insightsproductv1.GetProductMarginsRequest) (*insightsproductv1.GetProductMarginsResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:dashboard:view")
	if err != nil {
		return nil, err
	}

	buckets, err := a.Service.GetMarginDistribution(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load product margins")
	}

	out := make([]*insightsproductv1.MarginBucket, 0, len(buckets))
	for _, b := range buckets {
		out = append(out, &insightsproductv1.MarginBucket{
			Label:        b.Label,
			LowerPct:     b.LowerPct,
			UpperPct:     b.UpperPct,
			ProductCount: b.ProductCount,
		})
	}

	// Per-product margin rows (#853): join sales volume with cost price.
	rows, rowsErr := a.Service.GetProductMarginRows(ctx, businessID)
	if rowsErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load product margin rows")
	}
	products := make([]*insightsproductv1.ProductMargin, 0, len(rows))
	for _, r := range rows {
		pm := &insightsproductv1.ProductMargin{
			ProductId:         r.ID,
			Name:              r.Name,
			UnitsSold:         r.UnitsSold,
			RevenueMinorUnits: r.RevenueMinor,
		}
		if r.CostMinorUnits != nil {
			v := *r.CostMinorUnits
			pm.CostMinorUnits = &v
		}
		products = append(products, pm)
	}

	return &insightsproductv1.GetProductMarginsResponse{Buckets: out, Products: products}, nil
}

func (a *ProductAPI) GetSlowMovers(ctx context.Context, req *insightsproductv1.GetSlowMoversRequest) (*insightsproductv1.GetSlowMoversResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:dashboard:view")
	if err != nil {
		return nil, err
	}

	days := req.GetDays()
	if days <= 0 {
		days = 60
	}

	rows, err := a.Service.GetSlowMovers(ctx, businessID, days)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load slow movers")
	}

	products := make([]*insightsproductv1.SlowMover, 0, len(rows))
	for _, row := range rows {
		products = append(products, &insightsproductv1.SlowMover{
			ProductId:     row.ProductID,
			Name:          row.Name,
			StockQuantity: row.StockQuantity,
		})
	}

	return &insightsproductv1.GetSlowMoversResponse{Products: products}, nil
}

func (a *ProductAPI) GetProductPerformance(ctx context.Context, _ *insightsproductv1.GetProductPerformanceRequest) (*insightsproductv1.GetProductPerformanceResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:dashboard:view")
	if err != nil {
		return nil, err
	}

	kpis, err := a.Service.GetProductKPIs(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load product performance")
	}

	return &insightsproductv1.GetProductPerformanceResponse{
		TotalSkus:    kpis.TotalSKUs,
		ActiveSkus:   kpis.ActiveSKUs,
		OutOfStock:   kpis.OutOfStock,
		AvgMarginPct: kpis.AvgMarginPct,
	}, nil
}
