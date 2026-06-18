package suppliers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/suppliers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	if name != "erp.create_supplier" {
		return nil, false, nil
	}
	output, err := m.executeCreateSupplierTool(ctx, input)
	return output, true, err
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

	ownerID := tenant.ResolveOwnerID(ctx, authUser.BusinessID, authUser.UserID)
	id, err := m.service.CreateSupplier(ctx, ownerID, parsed)
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

func parseCreateSupplierToolInput(input map[string]any) (domain.SupplierInput, error) {
	isActive, err := supplierToolBool(input["is_active"], true)
	if err != nil {
		return domain.SupplierInput{}, fmt.Errorf("invalid is_active: %w", err)
	}

	parsed := domain.SupplierInput{
		Name:         strings.TrimSpace(supplierToolString(input["name"])),
		Email:        strings.TrimSpace(supplierToolString(input["email"])),
		Phone:        strings.TrimSpace(supplierToolString(input["phone"])),
		Address:      strings.TrimSpace(supplierToolString(input["address"])),
		City:         strings.TrimSpace(supplierToolString(input["city"])),
		Country:      strings.TrimSpace(supplierToolString(input["country"])),
		TaxID:        strings.TrimSpace(supplierToolString(input["tax_id"])),
		PaymentTerms: strings.TrimSpace(supplierToolString(input["payment_terms"])),
		Notes:        strings.TrimSpace(supplierToolString(input["notes"])),
		IsActive:     isActive,
	}
	if err := parsed.Validate(); err != nil {
		return domain.SupplierInput{}, err
	}
	return parsed, nil
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
