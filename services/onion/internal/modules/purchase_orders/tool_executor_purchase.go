package purchase_orders

import (
	"errors"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/purchase_orders/internal/domain"
)

func parseListPurchaseOrdersToolInput(input map[string]any) (listPurchaseOrdersToolInput, error) {
	rawStatus := firstPurchaseOrderToolValue(input, "status")
	status := normalizePurchaseOrderToolStatus(rawStatus)
	if strings.TrimSpace(purchaseOrderToolString(rawStatus)) != "" && status == "" {
		return listPurchaseOrdersToolInput{}, errors.New("status must be one of draft, sent, approved, partial, received, or cancelled")
	}

	page, err := purchaseOrderToolInt64(firstPurchaseOrderToolValue(input, "page"))
	if err != nil {
		return listPurchaseOrdersToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page < 0 {
		return listPurchaseOrdersToolInput{}, errors.New("page must be non-negative")
	}
	if page == 0 {
		page = 1
	}

	pageSize, err := purchaseOrderToolInt64(firstPurchaseOrderToolValue(input, "page_size", "limit"))
	if err != nil {
		return listPurchaseOrdersToolInput{}, fmt.Errorf("invalid page_size: %w", err)
	}
	if pageSize < 0 {
		return listPurchaseOrdersToolInput{}, errors.New("page_size must be non-negative")
	}
	if pageSize == 0 {
		pageSize = defaultListPurchaseOrdersToolPageSize
	}
	if pageSize > maxListPurchaseOrdersToolPageSize {
		pageSize = maxListPurchaseOrdersToolPageSize
	}

	return listPurchaseOrdersToolInput{
		Status:   status,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func parseCreatePurchaseOrderToolInput(input map[string]any) (domain.PurchaseOrderInput, error) {
	supplierID, err := purchaseOrderToolPositiveInt64(firstPurchaseOrderToolValue(input, "supplier_id"))
	if err != nil {
		return domain.PurchaseOrderInput{}, fmt.Errorf("invalid supplier_id: %w", err)
	}

	poNumber := strings.TrimSpace(purchaseOrderToolString(firstPurchaseOrderToolValue(input, "po_number", "po", "purchase_order_number", "number")))
	if poNumber == "" {
		return domain.PurchaseOrderInput{}, errors.New("po_number is required")
	}

	taxAmount, err := purchaseOrderToolNonNegativeFloat64(firstPurchaseOrderToolValue(input, "tax_amount", "tax"))
	if err != nil {
		return domain.PurchaseOrderInput{}, fmt.Errorf("invalid tax_amount: %w", err)
	}

	expectedDelivery, err := purchaseOrderToolDateString(firstPurchaseOrderToolValue(input, "expected_delivery", "expected_delivery_date", "delivery_date"))
	if err != nil {
		return domain.PurchaseOrderInput{}, fmt.Errorf("invalid expected_delivery: %w", err)
	}

	items, err := purchaseOrderToolItems(firstPurchaseOrderToolValue(input, "items", "line_items"))
	if err != nil {
		return domain.PurchaseOrderInput{}, fmt.Errorf("invalid items: %w", err)
	}

	parsed := domain.PurchaseOrderInput{
		SupplierID:       supplierID,
		PONumber:         poNumber,
		TaxAmount:        taxAmount,
		ExpectedDelivery: expectedDelivery,
		Notes:            strings.TrimSpace(purchaseOrderToolString(firstPurchaseOrderToolValue(input, "notes"))),
		Items:            items,
	}
	if err := parsed.Validate(); err != nil {
		return domain.PurchaseOrderInput{}, err
	}
	return parsed, nil
}

func parseApprovePurchaseOrderToolInput(input map[string]any) (int64, error) {
	return parsePurchaseOrderIDToolInput(input)
}

func parsePurchaseOrderIDToolInput(input map[string]any) (int64, error) {
	poID, err := purchaseOrderIDToolValue(firstPurchaseOrderToolValue(input, "purchase_order_id", "po_id", "purchase_order", "po", "id"))
	if err != nil {
		return 0, fmt.Errorf("invalid purchase_order_id: %w", err)
	}
	if poID <= 0 {
		return 0, errors.New("purchase_order_id is required")
	}
	return poID, nil
}

func firstPurchaseOrderToolValue(input map[string]any, keys ...string) any {
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

func purchaseOrderToolItems(value any) ([]domain.PurchaseOrderItemInput, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case []domain.PurchaseOrderItemInput:
		return v, nil
	case []map[string]any:
		items := make([]domain.PurchaseOrderItemInput, 0, len(v))
		for idx, item := range v {
			parsed, err := purchaseOrderToolItem(item)
			if err != nil {
				return nil, fmt.Errorf("item %d: %w", idx+1, err)
			}
			items = append(items, parsed)
		}
		return items, nil
	case []any:
		items := make([]domain.PurchaseOrderItemInput, 0, len(v))
		for idx, item := range v {
			itemMap, ok := item.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("item %d must be an object", idx+1)
			}
			parsed, err := purchaseOrderToolItem(itemMap)
			if err != nil {
				return nil, fmt.Errorf("item %d: %w", idx+1, err)
			}
			items = append(items, parsed)
		}
		return items, nil
	default:
		return nil, fmt.Errorf("must be an array, got %T", value)
	}
}

func purchaseOrderToolItem(input map[string]any) (domain.PurchaseOrderItemInput, error) {
	productID, err := purchaseOrderToolPositiveInt64(firstPurchaseOrderToolValue(input, "product_id"))
	if err != nil {
		return domain.PurchaseOrderItemInput{}, fmt.Errorf("invalid product_id: %w", err)
	}

	quantityOrdered, err := purchaseOrderToolPositiveFloat64(firstPurchaseOrderToolValue(input, "quantity_ordered", "quantity", "qty", "units"))
	if err != nil {
		return domain.PurchaseOrderItemInput{}, fmt.Errorf("invalid quantity_ordered: %w", err)
	}

	unitCost, err := purchaseOrderToolNonNegativeFloat64(firstPurchaseOrderToolValue(input, "unit_cost", "cost", "price"))
	if err != nil {
		return domain.PurchaseOrderItemInput{}, fmt.Errorf("invalid unit_cost: %w", err)
	}

	parsed := domain.PurchaseOrderItemInput{
		ProductID:       productID,
		QuantityOrdered: quantityOrdered,
		UnitCost:        unitCost,
	}
	if err := parsed.Validate(); err != nil {
		return domain.PurchaseOrderItemInput{}, err
	}
	return parsed, nil
}

func purchaseOrderIDToolValue(value any) (int64, error) {
	switch v := value.(type) {
	case string:
		raw := strings.TrimSpace(v)
		raw = strings.TrimPrefix(raw, "#")
		raw = strings.ToLower(raw)
		for _, prefix := range []string{"purchase-order-", "purchase_order_", "purchaseorder-", "purchaseorder_", "po-", "po_"} {
			raw = strings.TrimPrefix(raw, prefix)
		}
		return purchaseOrderToolInt64(raw)
	default:
		return purchaseOrderToolInt64(value)
	}
}

func purchaseOrderToolPositiveInt64(value any) (int64, error) {
	out, err := purchaseOrderToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out <= 0 {
		return 0, errors.New("must be positive")
	}
	return out, nil
}
