package customers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	listCustomersToolName         = "crm.list_customers"
	getCustomerToolName           = "crm.get_customer"
	createCustomerToolName        = "crm.create_customer"
	defaultListCustomersToolLimit = 20
	maxListCustomersToolPageSize  = 50
	customerToolTimestampDate     = "2006-01-02T15:04:05Z"
	customerToolCalendarDate      = "2006-01-02"
)

type listCustomersToolInput struct {
	Search   string
	Tier     string
	Tag      string
	Page     int
	PageSize int
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	required, dispatched := customerToolPermissions[name]
	if !dispatched {
		return nil, false, nil
	}
	if err := requireCustomerToolPermission(ctx, required); err != nil {
		return nil, true, err
	}
	switch name {
	case listCustomersToolName:
		output, err := m.executeListCustomersTool(ctx, input)
		return output, true, err
	case getCustomerToolName:
		output, err := m.executeGetCustomerTool(ctx, input)
		return output, true, err
	case createCustomerToolName:
		output, err := m.executeCreateCustomerTool(ctx, input)
		return output, true, err
	default:
		return nil, true, fmt.Errorf("customer tool %q has no executor", name)
	}
}

func (m *Module) executeListCustomersTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseListCustomersToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, err := customerToolBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if err != nil {
		return nil, err
	}

	result, err := m.service.ListCustomers(ctx, domain.CustomerListFilter{
		OwnerID:    authUser.UserID,
		BusinessID: businessID,
		Search:     parsed.Search,
		Tier:       parsed.Tier,
		Tag:        parsed.Tag,
		Page:       parsed.Page,
		PageSize:   parsed.PageSize,
	})
	if err != nil {
		return nil, err
	}

	customers := make([]map[string]any, 0, len(result.Customers))
	for _, customer := range result.Customers {
		customers = append(customers, customerToolPayload(customer))
	}

	return map[string]any{
		"customers": customers,
		"total":     result.Total,
		"page":      result.Page,
		"page_size": result.PageSize,
		"returned":  len(customers),
		"filters": map[string]any{
			"search": parsed.Search,
			"tier":   parsed.Tier,
			"tag":    parsed.Tag,
		},
	}, nil
}

func (m *Module) executeGetCustomerTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	customerID, err := parseGetCustomerToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, err := customerToolBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if err != nil {
		return nil, err
	}
	customer, err := m.service.GetCustomer(ctx, businessID, customerID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"customer": customerDetailToolPayload(customer),
	}, nil
}

func (m *Module) executeCreateCustomerTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	firstName := strings.TrimSpace(toolStringValue(input["first_name"]))
	lastName := strings.TrimSpace(toolStringValue(input["last_name"]))
	if firstName == "" && lastName == "" {
		firstName, lastName = splitCustomerName(toolStringValue(input["name"]))
	}

	customer := domain.CustomerInput{
		OwnerID:   authUser.UserID,
		FirstName: firstName,
		LastName:  lastName,
		Email:     strings.TrimSpace(toolStringValue(input["email"])),
		Phone:     strings.TrimSpace(toolStringValue(input["phone"])),
		Gender:    strings.TrimSpace(toolStringValue(input["gender"])),
		Address:   strings.TrimSpace(toolStringValue(input["address"])),
		City:      strings.TrimSpace(toolStringValue(input["city"])),
		Country:   strings.TrimSpace(toolStringValue(input["country"])),
		Notes:     strings.TrimSpace(toolStringValue(input["notes"])),
	}

	tags, err := toolStringSlice(input["tags"])
	if err != nil {
		return nil, fmt.Errorf("invalid tags: %w", err)
	}
	customer.Tags = tags

	dateOfBirth, err := optionalCustomerDate(input["date_of_birth"])
	if err != nil {
		return nil, fmt.Errorf("invalid date_of_birth: %w", err)
	}
	customer.DateOfBirth = dateOfBirth
	businessID, err := customerToolBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if err != nil {
		return nil, err
	}
	customer.BusinessID = businessID

	if message := domain.ValidateCustomerInput(customer); message != "" {
		return nil, errors.New(message)
	}

	id, err := m.service.CreateCustomer(ctx, customer)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"customer": map[string]any{
			"id":            id,
			"first_name":    customer.FirstName,
			"last_name":     customer.LastName,
			"email":         customer.Email,
			"phone":         customer.Phone,
			"date_of_birth": formatCustomerDate(customer.DateOfBirth),
			"gender":        customer.Gender,
			"address":       customer.Address,
			"city":          customer.City,
			"country":       customer.Country,
			"tags":          customer.Tags,
			"notes":         customer.Notes,
		},
	}, nil
}

func parseGetCustomerToolInput(input map[string]any) (int64, error) {
	customerID, err := customerToolNonNegativeInt64(firstCustomerToolValue(input, "customer_id", "id"))
	if err != nil {
		return 0, fmt.Errorf("invalid customer_id: %w", err)
	}
	if customerID == 0 {
		return 0, errors.New("customer_id is required")
	}
	return customerID, nil
}

func splitCustomerName(value string) (string, string) {
	name := strings.TrimSpace(value)
	if name == "" {
		return "", ""
	}
	parts := strings.Fields(name)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}

func parseListCustomersToolInput(input map[string]any) (listCustomersToolInput, error) {
	page, err := customerToolNonNegativeInt(firstCustomerToolValue(input, "page"))
	if err != nil {
		return listCustomersToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page == 0 {
		page = 1
	}

	pageSize, err := customerToolNonNegativeInt(firstCustomerToolValue(input, "page_size", "limit"))
	if err != nil {
		return listCustomersToolInput{}, fmt.Errorf("invalid page_size: %w", err)
	}
	if pageSize == 0 {
		pageSize = defaultListCustomersToolLimit
	}
	if pageSize > maxListCustomersToolPageSize {
		pageSize = maxListCustomersToolPageSize
	}

	return listCustomersToolInput{
		Search:   strings.TrimSpace(toolStringValue(firstCustomerToolValue(input, "search", "query", "q"))),
		Tier:     strings.TrimSpace(toolStringValue(firstCustomerToolValue(input, "tier", "loyalty_tier"))),
		Tag:      strings.TrimSpace(toolStringValue(firstCustomerToolValue(input, "tag"))),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func firstCustomerToolValue(input map[string]any, keys ...string) any {
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

func customerToolNonNegativeInt(value any) (int, error) {
	out, err := customerToolInt(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func customerToolNonNegativeInt64(value any) (int64, error) {
	out, err := customerToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func customerToolInt64(value any) (int64, error) {
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
		if v > float64(math.MaxInt) || v < float64(math.MinInt) {
			return 0, errors.New("out of range")
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

func customerToolInt(value any) (int, error) {
	switch v := value.(type) {
	case nil:
		return 0, nil
	case int:
		return v, nil
	case int64:
		if v > math.MaxInt || v < math.MinInt {
			return 0, errors.New("out of range")
		}
		return int(v), nil
	case int32:
		return int(v), nil
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return 0, errors.New("must be finite")
		}
		if math.Trunc(v) != v {
			return 0, errors.New("must be an integer")
		}
		if v > float64(math.MaxInt) || v < float64(math.MinInt) {
			return 0, errors.New("out of range")
		}
		return int(v), nil
	case json.Number:
		n, err := v.Int64()
		if err != nil {
			return 0, err
		}
		if n > math.MaxInt || n < math.MinInt {
			return 0, errors.New("out of range")
		}
		return int(n), nil
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return 0, nil
		}
		n, err := strconv.ParseInt(raw, 10, 0)
		if err != nil {
			return 0, err
		}
		return int(n), nil
	default:
		return 0, fmt.Errorf("unsupported value %T", value)
	}
}

func customerToolBusinessID(ctx context.Context, businessID, userID int64) (int64, error) {
	resolved, err := tenant.ResolveBusinessID(ctx, businessID, userID)
	if err != nil {
		return 0, errors.New("cannot determine business context")
	}
	return resolved, nil
}

func customerToolPayload(c domain.Customer) map[string]any {
	fullName := strings.TrimSpace(strings.TrimSpace(c.FirstName) + " " + strings.TrimSpace(c.LastName))
	return map[string]any{
		"id":             c.ID,
		"first_name":     c.FirstName,
		"last_name":      c.LastName,
		"full_name":      fullName,
		"email":          c.Email,
		"phone":          c.Phone,
		"loyalty_points": c.LoyaltyPoints,
		"loyalty_tier":   c.LoyaltyTier,
		"total_spent":    c.TotalSpent,
		"visit_count":    c.VisitCount,
		"last_visit_at":  formatCustomerTimestamp(c.LastVisitAt),
		"tags":           c.Tags,
		"is_active":      c.IsActive,
		"created_at":     c.CreatedAt.Format(customerToolTimestampDate),
	}
}

func customerDetailToolPayload(c domain.Customer) map[string]any {
	payload := customerToolPayload(c)
	payload["business_id"] = c.BusinessID
	payload["date_of_birth"] = formatCustomerDate(c.DateOfBirth)
	payload["gender"] = c.Gender
	payload["address"] = c.Address
	payload["city"] = c.City
	payload["country"] = c.Country
	payload["notes"] = c.Notes
	payload["updated_at"] = c.UpdatedAt.Format(customerToolTimestampDate)
	return payload
}

func toolStringValue(value any) string {
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

func toolStringSlice(value any) ([]string, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case []string:
		return cleanCustomerTags(v), nil
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			tag := strings.TrimSpace(toolStringValue(item))
			if tag != "" {
				out = append(out, tag)
			}
		}
		return out, nil
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return nil, nil
		}
		parts := strings.Split(raw, ",")
		return cleanCustomerTags(parts), nil
	default:
		return nil, fmt.Errorf("unsupported value %T", value)
	}
}

func cleanCustomerTags(tags []string) []string {
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		clean := strings.TrimSpace(tag)
		if clean != "" {
			out = append(out, clean)
		}
	}
	return out
}

func optionalCustomerDate(value any) (*time.Time, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case time.Time:
		return &v, nil
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return nil, nil
		}
		for _, layout := range []string{customerToolCalendarDate, time.RFC3339, time.RFC3339Nano} {
			parsed, err := time.Parse(layout, raw)
			if err == nil {
				return &parsed, nil
			}
		}
		return nil, errors.New("must be YYYY-MM-DD or RFC3339")
	default:
		return nil, fmt.Errorf("unsupported value %T", value)
	}
}

func formatCustomerDate(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.Format(customerToolCalendarDate)
}

func formatCustomerTimestamp(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.Format(customerToolTimestampDate)
}
