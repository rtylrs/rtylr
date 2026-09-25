package store_products

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/store_products/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/pagination"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	listStoreProductsToolName         = "store.list_products"
	getStoreProductToolName           = "store.get_product"
	defaultListStoreProductsToolLimit = int64(20)
	maxListStoreProductsToolLimit     = int64(50)
	storeProductToolTimestampDate     = "2006-01-02T15:04:05Z"
)

type getStoreProductToolInput struct {
	StoreProductID int64
	ProductID      int64
	Slug           string
	Title          string
}

type listStoreProductsToolInput struct {
	Query     string
	ProductID int64
	Published *bool
	Page      int64
	Limit     int64
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listStoreProductsToolName:
		output, err := m.executeListStoreProductsTool(ctx, input)
		return output, true, err
	case getStoreProductToolName:
		output, err := m.executeGetStoreProductTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeGetStoreProductTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	parsed, err := parseGetStoreProductToolInput(input)
	if err != nil {
		return nil, err
	}

	products, err := m.service.List(ctx, businessID)
	if err != nil {
		return nil, err
	}
	product, err := findStoreProductForTool(products, parsed)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"product": storeProductToolPayload(product),
	}, nil
}

func (m *Module) executeListStoreProductsTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	parsed, err := parseListStoreProductsToolInput(input)
	if err != nil {
		return nil, err
	}

	products, err := m.service.List(ctx, businessID)
	if err != nil {
		return nil, err
	}
	products = filterStoreProductsForTool(products, parsed)

	start := pagination.SliceStart(parsed.Page, parsed.Limit, len(products))
	if start > len(products) {
		start = len(products)
	}
	end := start + int(parsed.Limit)
	if end > len(products) {
		end = len(products)
	}

	payloads := make([]map[string]any, 0, end-start)
	for _, product := range products[start:end] {
		payloads = append(payloads, storeProductToolPayload(product))
	}

	return map[string]any{
		"products":  payloads,
		"total":     len(products),
		"returned":  len(payloads),
		"page":      parsed.Page,
		"page_size": parsed.Limit,
		"has_more":  end < len(products),
		"filters": map[string]any{
			"query":        parsed.Query,
			"product_id":   parsed.ProductID,
			"is_published": storeProductToolOptionalBool(parsed.Published),
		},
	}, nil
}

func parseGetStoreProductToolInput(input map[string]any) (getStoreProductToolInput, error) {
	storeProductID, err := storeProductToolIDValue(firstStoreProductToolValue(input, "store_product_id", "storefront_product_id", "store_product", "id"))
	if err != nil {
		return getStoreProductToolInput{}, fmt.Errorf("invalid store_product_id: %w", err)
	}
	productID, err := storeProductToolIDValue(firstStoreProductToolValue(input, "product_id", "inventory_product_id"))
	if err != nil {
		return getStoreProductToolInput{}, fmt.Errorf("invalid product_id: %w", err)
	}
	slug := strings.ToLower(strings.TrimSpace(storeProductToolString(firstStoreProductToolValue(input, "slug", "handle"))))
	title := strings.ToLower(strings.TrimSpace(storeProductToolString(firstStoreProductToolValue(input, "title", "name"))))

	if storeProductID <= 0 && productID <= 0 && slug == "" && title == "" {
		return getStoreProductToolInput{}, errors.New("store_product_id, product_id, slug, or title is required")
	}
	return getStoreProductToolInput{
		StoreProductID: storeProductID,
		ProductID:      productID,
		Slug:           slug,
		Title:          title,
	}, nil
}

func parseListStoreProductsToolInput(input map[string]any) (listStoreProductsToolInput, error) {
	query := strings.TrimSpace(storeProductToolString(firstStoreProductToolValue(input, "query", "q", "search", "title")))

	productID, err := storeProductToolNonNegativeInt64(firstStoreProductToolValue(input, "product_id", "inventory_product_id"))
	if err != nil {
		return listStoreProductsToolInput{}, fmt.Errorf("invalid product_id: %w", err)
	}

	published, err := storeProductToolOptionalBoolInput(firstStoreProductToolValue(input, "published", "is_published", "active", "status"))
	if err != nil {
		return listStoreProductsToolInput{}, fmt.Errorf("invalid published: %w", err)
	}

	page, err := storeProductToolNonNegativeInt64(firstStoreProductToolValue(input, "page"))
	if err != nil {
		return listStoreProductsToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page == 0 {
		page = 1
	}

	limit, err := storeProductToolNonNegativeInt64(firstStoreProductToolValue(input, "limit", "page_size"))
	if err != nil {
		return listStoreProductsToolInput{}, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultListStoreProductsToolLimit
	}
	if limit > maxListStoreProductsToolLimit {
		limit = maxListStoreProductsToolLimit
	}

	return listStoreProductsToolInput{
		Query:     query,
		ProductID: productID,
		Published: published,
		Page:      page,
		Limit:     limit,
	}, nil
}

func filterStoreProductsForTool(products []domain.StoreProduct, input listStoreProductsToolInput) []domain.StoreProduct {
	if input.Query == "" && input.ProductID == 0 && input.Published == nil {
		return products
	}

	query := strings.ToLower(input.Query)
	out := make([]domain.StoreProduct, 0, len(products))
	for _, product := range products {
		if input.ProductID > 0 && product.ProductID != input.ProductID {
			continue
		}
		if input.Published != nil && product.IsPublished != *input.Published {
			continue
		}
		if query != "" && !storeProductToolMatchesQuery(product, query) {
			continue
		}
		out = append(out, product)
	}
	return out
}

func findStoreProductForTool(products []domain.StoreProduct, input getStoreProductToolInput) (domain.StoreProduct, error) {
	for _, product := range products {
		if input.StoreProductID > 0 && product.ID == input.StoreProductID {
			return product, nil
		}
		if input.ProductID > 0 && product.ProductID == input.ProductID {
			return product, nil
		}
		if input.Slug != "" && strings.EqualFold(product.Slug, input.Slug) {
			return product, nil
		}
		if input.Title != "" && strings.EqualFold(product.Title, input.Title) {
			return product, nil
		}
	}
	return domain.StoreProduct{}, domain.ErrStoreProductNotFound
}

func storeProductToolMatchesQuery(product domain.StoreProduct, query string) bool {
	return strings.Contains(strings.ToLower(product.Title), query) ||
		strings.Contains(strings.ToLower(product.Slug), query) ||
		strings.Contains(strings.ToLower(product.Description), query)
}

func storeProductToolPayload(product domain.StoreProduct) map[string]any {
	return map[string]any{
		"id":               product.ID,
		"owner_id":         product.OwnerID,
		"product_id":       product.ProductID,
		"title":            product.Title,
		"slug":             product.Slug,
		"description":      product.Description,
		"images":           product.Images,
		"price":            product.Price,
		"compare_at_price": product.CompareAtPrice,
		"is_published":     product.IsPublished,
		"sort_order":       product.SortOrder,
		"created_at":       formatStoreProductToolTime(product.CreatedAt),
		"updated_at":       formatStoreProductToolTime(product.UpdatedAt),
	}
}

func firstStoreProductToolValue(input map[string]any, keys ...string) any {
	for _, key := range keys {
		value, ok := input[key]
		if !ok || value == nil {
			continue
		}
		if s, ok := value.(string); ok && strings.TrimSpace(s) == "" {
			continue
		}
		return value
	}
	return nil
}

func storeProductToolOptionalBoolInput(value any) (*bool, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case bool:
		return &v, nil
	case int:
		return storeProductToolOptionalBoolInt(int64(v))
	case int64:
		return storeProductToolOptionalBoolInt(v)
	case int32:
		return storeProductToolOptionalBoolInt(int64(v))
	case float64:
		if math.Trunc(v) != v {
			return nil, errors.New("must be 0 or 1")
		}
		return storeProductToolOptionalBoolInt(int64(v))
	case string:
		raw := strings.ToLower(strings.TrimSpace(v))
		switch raw {
		case "", "all", "any":
			return nil, nil
		case "true", "yes", "1", "published", "active", "enabled":
			out := true
			return &out, nil
		case "false", "no", "0", "unpublished", "inactive", "disabled":
			out := false
			return &out, nil
		default:
			return nil, errors.New("must be true, false, published, unpublished, or all")
		}
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			return nil, err
		}
		return storeProductToolOptionalBoolInt(i)
	default:
		return nil, fmt.Errorf("unsupported value %T", value)
	}
}

func storeProductToolOptionalBoolInt(value int64) (*bool, error) {
	if value == 0 {
		out := false
		return &out, nil
	}
	if value == 1 {
		out := true
		return &out, nil
	}
	return nil, errors.New("must be 0 or 1")
}

func storeProductToolOptionalBool(value *bool) any {
	if value == nil {
		return nil
	}
	return *value
}

func storeProductToolNonNegativeInt64(value any) (int64, error) {
	out, err := storeProductToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func storeProductToolIDValue(value any) (int64, error) {
	switch v := value.(type) {
	case string:
		raw := strings.TrimSpace(v)
		raw = strings.TrimPrefix(raw, "#")
		raw = strings.ToLower(raw)
		for _, prefix := range []string{"store-product-", "store_product_", "storefront-product-", "storefront_product_", "product-", "product_"} {
			raw = strings.TrimPrefix(raw, prefix)
		}
		return storeProductToolInt64(raw)
	default:
		return storeProductToolInt64(value)
	}
}

func storeProductToolInt64(value any) (int64, error) {
	switch v := value.(type) {
	case nil:
		return 0, nil
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case int32:
		return int64(v), nil
	case float64:
		if math.Trunc(v) != v {
			return 0, errors.New("must be an integer")
		}
		return int64(v), nil
	case json.Number:
		return v.Int64()
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return 0, nil
		}
		return strconv.ParseInt(raw, 10, 64)
	default:
		return 0, fmt.Errorf("unsupported value %T", value)
	}
}

func storeProductToolString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case json.Number:
		return v.String()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}

func formatStoreProductToolTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC().Format(storeProductToolTimestampDate)
}
