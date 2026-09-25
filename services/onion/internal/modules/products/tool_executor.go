package products

import (
	"context"
	"errors"
	"fmt"

	"github.com/voxire/onion/internal/modules/products/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	getProductToolName             = "erp.get_product"
	listReorderSuggestionsToolName = "erp.list_reorder_suggestions"
	createProductToolName          = "erp.create_product"
	updateProductToolName          = "erp.update_product"
	deleteProductToolName          = "erp.delete_product"
)
const maxProductToolLeadTimeDays = int64(1<<31 - 1)

type getProductToolInput struct {
	ProductID int64
	Barcode   string
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case getProductToolName:
		output, err := m.executeGetProductTool(ctx, input)
		return output, true, err
	case listReorderSuggestionsToolName:
		output, err := m.executeListReorderSuggestionsTool(ctx)
		return output, true, err
	case createProductToolName:
		output, err := m.executeCreateProductTool(ctx, input)
		return output, true, err
	case updateProductToolName:
		output, err := m.executeUpdateProductTool(ctx, input)
		return output, true, err
	case deleteProductToolName:
		output, err := m.executeDeleteProductTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListReorderSuggestionsTool(ctx context.Context) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	ownerID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, bizErr
	}
	suggestions, err := m.service.ListReorderSuggestions(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(suggestions.Suggestions))
	totalSuggested := 0.0
	for _, suggestion := range suggestions.Suggestions {
		payloads = append(payloads, reorderSuggestionToolPayload(suggestion))
		totalSuggested += suggestion.SuggestedQty
	}

	return map[string]any{
		"suggestions":         payloads,
		"returned":            len(payloads),
		"total_suggested_qty": totalSuggested,
		"confidence":          "low",
	}, nil
}

func (m *Module) executeGetProductTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseGetProductToolInput(input)
	if err != nil {
		return nil, err
	}

	ownerID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, bizErr
	}

	var product domain.Product
	lookupBy := "id"
	if parsed.ProductID > 0 {
		product, err = m.service.GetProduct(ctx, ownerID, parsed.ProductID)
	} else {
		lookupBy = "barcode"
		product, err = m.service.GetProductByBarcode(ctx, ownerID, parsed.Barcode)
	}
	if err != nil {
		return nil, err
	}

	attrs, err := m.service.ListProductAttributes(ctx, ownerID, product.ID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"product":   productToolPayload(product, attrs),
		"lookup_by": lookupBy,
	}, nil
}

func (m *Module) executeCreateProductTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseCreateProductToolInput(input)
	if err != nil {
		return nil, err
	}

	ownerID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, bizErr
	}
	id, err := m.service.CreateProduct(ctx, ownerID, parsed)
	if err != nil {
		return nil, err
	}
	created, err := m.service.GetProduct(ctx, ownerID, id)
	if err != nil {
		return nil, err
	}
	attrs, err := m.service.ListProductAttributes(ctx, ownerID, id)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"message": fmt.Sprintf("Product %q was added to inventory.", created.Name),
		"product": productToolPayload(created, attrs),
	}, nil
}

func (m *Module) executeUpdateProductTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	productID, err := parseProductToolID(input)
	if err != nil {
		return nil, err
	}
	ownerID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, bizErr
	}
	existing, err := m.service.GetProduct(ctx, ownerID, productID)
	if err != nil {
		return nil, err
	}
	parsed, err := mergeProductUpdateToolInput(existing, input)
	if err != nil {
		return nil, err
	}
	if err := m.service.UpdateProduct(ctx, ownerID, productID, parsed); err != nil {
		return nil, err
	}

	updated, err := m.service.GetProduct(ctx, ownerID, productID)
	if err != nil {
		return nil, err
	}
	attrs, err := m.service.ListProductAttributes(ctx, ownerID, productID)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"product": productToolPayload(updated, attrs),
		"updated": true,
	}, nil
}

func (m *Module) executeDeleteProductTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	productID, err := parseProductToolID(input)
	if err != nil {
		return nil, err
	}
	ownerID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, bizErr
	}
	product, err := m.service.GetProduct(ctx, ownerID, productID)
	if err != nil {
		return nil, err
	}
	if err := m.service.DeleteProduct(ctx, ownerID, productID); err != nil {
		return nil, err
	}

	return map[string]any{
		"product": map[string]any{
			"id":      product.ID,
			"name":    product.Name,
			"deleted": true,
		},
	}, nil
}
