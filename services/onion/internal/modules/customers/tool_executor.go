package customers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
)

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	if name != "crm.create_customer" {
		return nil, false, nil
	}
	output, err := m.executeCreateCustomerTool(ctx, input)
	return output, true, err
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
		for _, layout := range []string{"2006-01-02", time.RFC3339, time.RFC3339Nano} {
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
	return value.Format("2006-01-02")
}
