package suppliers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/suppliers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	createSupplierToolName           = "erp.create_supplier"
	getSupplierToolName              = "erp.get_supplier"
	listSuppliersToolName            = "erp.list_suppliers"
	defaultListSuppliersToolPageSize = int64(20)
	maxListSuppliersToolPageSize     = int64(50)
)

type getSupplierToolInput struct {
	SupplierID int64
}

type listSuppliersToolInput struct {
	Page     int64
	PageSize int64
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case getSupplierToolName:
		output, err := m.executeGetSupplierTool(ctx, input)
		return output, true, err
	case listSuppliersToolName:
		output, err := m.executeListSuppliersTool(ctx, input)
		return output, true, err
	case createSupplierToolName:
		output, err := m.executeCreateSupplierTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeGetSupplierTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseGetSupplierToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	supplier, err := m.service.GetSupplierByID(ctx, businessID, parsed.SupplierID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"supplier": supplierToolPayload(supplier),
	}, nil
}

func (m *Module) executeListSuppliersTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseListSuppliersToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	suppliers, hasMore, err := m.service.ListSuppliers(ctx, businessID, parsed.Page, parsed.PageSize)
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(suppliers))
	for _, supplier := range suppliers {
		payloads = append(payloads, supplierToolPayload(supplier))
	}

	return map[string]any{
		"suppliers": payloads,
		"returned":  len(payloads),
		"page":      parsed.Page,
		"page_size": parsed.PageSize,
		"has_more":  hasMore,
	}, nil
}

func (m *Module) executeCreateSupplierTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseCreateSupplierToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	id, err := m.service.CreateSupplier(ctx, businessID, parsed)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"supplier": map[string]any{
			"id":            id,
			"name":          parsed.Name,
			"email":         parsed.Email,
			"phone":         parsed.Phone,
			"address":       parsed.Address,
			"city":          parsed.City,
			"country":       parsed.Country,
			"tax_id":        parsed.TaxID,
			"payment_terms": parsed.PaymentTerms,
			"notes":         parsed.Notes,
			"is_active":     parsed.IsActive,
		},
	}, nil
}

func parseGetSupplierToolInput(input map[string]any) (getSupplierToolInput, error) {
	supplierID, err := supplierIDToolValue(firstSupplierToolValue(input, "supplier_id", "supplier", "id"))
	if err != nil {
		return getSupplierToolInput{}, fmt.Errorf("invalid supplier_id: %w", err)
	}
	if supplierID <= 0 {
		return getSupplierToolInput{}, errors.New("supplier_id is required")
	}
	return getSupplierToolInput{SupplierID: supplierID}, nil
}

func parseCreateSupplierToolInput(input map[string]any) (domain.SupplierInput, error) {
	isActive, err := supplierToolBool(firstSupplierToolValue(input, "is_active", "active"), true)
	if err != nil {
		return domain.SupplierInput{}, fmt.Errorf("invalid is_active: %w", err)
	}

	parsed := domain.SupplierInput{
		Name:         strings.TrimSpace(supplierToolString(firstSupplierToolValue(input, "name", "supplier_name"))),
		Email:        strings.TrimSpace(supplierToolString(firstSupplierToolValue(input, "email"))),
		Phone:        strings.TrimSpace(supplierToolString(firstSupplierToolValue(input, "phone"))),
		Address:      strings.TrimSpace(supplierToolString(firstSupplierToolValue(input, "address"))),
		City:         strings.TrimSpace(supplierToolString(firstSupplierToolValue(input, "city"))),
		Country:      strings.TrimSpace(supplierToolString(firstSupplierToolValue(input, "country"))),
		TaxID:        strings.TrimSpace(supplierToolString(firstSupplierToolValue(input, "tax_id"))),
		PaymentTerms: strings.TrimSpace(supplierToolString(firstSupplierToolValue(input, "payment_terms"))),
		Notes:        strings.TrimSpace(supplierToolString(firstSupplierToolValue(input, "notes"))),
		IsActive:     isActive,
	}
	if err := parsed.Validate(); err != nil {
		return domain.SupplierInput{}, err
	}
	return parsed, nil
}

func parseListSuppliersToolInput(input map[string]any) (listSuppliersToolInput, error) {
	page, err := supplierToolInt64(firstSupplierToolValue(input, "page"))
	if err != nil {
		return listSuppliersToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page < 0 {
		return listSuppliersToolInput{}, errors.New("page must be non-negative")
	}
	if page == 0 {
		page = 1
	}

	pageSize, err := supplierToolInt64(firstSupplierToolValue(input, "page_size", "limit"))
	if err != nil {
		return listSuppliersToolInput{}, fmt.Errorf("invalid page_size: %w", err)
	}
	if pageSize < 0 {
		return listSuppliersToolInput{}, errors.New("page_size must be non-negative")
	}
	if pageSize == 0 {
		pageSize = defaultListSuppliersToolPageSize
	}
	if pageSize > maxListSuppliersToolPageSize {
		pageSize = maxListSuppliersToolPageSize
	}

	return listSuppliersToolInput{
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func firstSupplierToolValue(input map[string]any, keys ...string) any {
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

func supplierIDToolValue(value any) (int64, error) {
	switch v := value.(type) {
	case string:
		raw := strings.TrimSpace(v)
		raw = strings.TrimPrefix(raw, "#")
		raw = strings.ToLower(raw)
		for _, prefix := range []string{"supplier-", "supplier_", "sup-", "sup_"} {
			raw = strings.TrimPrefix(raw, prefix)
		}
		return supplierToolInt64(raw)
	default:
		return supplierToolInt64(value)
	}
}

func supplierToolPayload(supplier domain.Supplier) map[string]any {
	return map[string]any{
		"id":            supplier.ID,
		"name":          supplier.Name,
		"email":         supplier.Email,
		"phone":         supplier.Phone,
		"address":       supplier.Address,
		"city":          supplier.City,
		"country":       supplier.Country,
		"tax_id":        supplier.TaxID,
		"payment_terms": supplier.PaymentTerms,
		"notes":         supplier.Notes,
		"is_active":     supplier.IsActive,
		"created_at":    supplierToolTime(supplier.CreatedAt),
		"updated_at":    supplierToolTime(supplier.UpdatedAt),
	}
}

func supplierToolInt64(value any) (int64, error) {
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
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return 0, errors.New("must be finite")
		}
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

func supplierToolBool(value any, defaultValue bool) (bool, error) {
	switch v := value.(type) {
	case nil:
		return defaultValue, nil
	case bool:
		return v, nil
	case json.Number:
		n, err := strconv.ParseInt(v.String(), 10, 64)
		if err != nil {
			return false, err
		}
		switch n {
		case 0:
			return false, nil
		case 1:
			return true, nil
		default:
			return false, errors.New("must be true or false")
		}
	case string:
		raw := strings.ToLower(strings.TrimSpace(v))
		if raw == "" {
			return defaultValue, nil
		}
		switch raw {
		case "true", "t", "yes", "y", "1", "active":
			return true, nil
		case "false", "f", "no", "n", "0", "inactive":
			return false, nil
		default:
			return false, errors.New("must be true or false")
		}
	default:
		return false, fmt.Errorf("unsupported value %T", value)
	}
}

func supplierToolString(value any) string {
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

func supplierToolTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
