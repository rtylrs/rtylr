package hr_core

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
)

func roleToolPermissionKeys(raw string) []string {
	var permissions map[string]bool
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &permissions); err != nil {
		return nil
	}

	keys := make([]string, 0, len(permissions))
	for key, enabled := range permissions {
		if enabled {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

func splitEmployeeName(value string) (string, string) {
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

func normalizeEmploymentType(value any) string {
	employeeType := strings.ToLower(strings.TrimSpace(hrToolString(value)))
	if employeeType == "" {
		return domain.EmploymentTypeFullTime
	}
	employeeType = strings.ReplaceAll(employeeType, "-", "_")
	employeeType = strings.ReplaceAll(employeeType, " ", "_")
	switch employeeType {
	case domain.EmploymentTypeFullTime, domain.EmploymentTypePartTime, domain.EmploymentTypeContract, domain.EmploymentTypeFreelance:
		return employeeType
	default:
		return ""
	}
}

func hrToolNonNegativeInt64(value any) (int64, error) {
	out, err := hrToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func hrToolInt64(value any) (int64, error) {
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

func hrToolNonNegativeFloat64(value any) (float64, error) {
	out, err := hrToolFloat64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func hrToolFloat64(value any) (float64, error) {
	switch v := value.(type) {
	case nil:
		return 0, nil
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case json.Number:
		return v.Float64()
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return 0, nil
		}
		return strconv.ParseFloat(raw, 64)
	default:
		return 0, fmt.Errorf("unsupported value %T", value)
	}
}

func optionalEmployeeToolDate(value any) (time.Time, error) {
	switch v := value.(type) {
	case nil:
		return time.Time{}, nil
	case time.Time:
		return v, nil
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return time.Time{}, nil
		}
		for _, layout := range []string{"2006-01-02", time.RFC3339, time.RFC3339Nano} {
			parsed, err := time.Parse(layout, raw)
			if err == nil {
				return parsed, nil
			}
		}
		return time.Time{}, errors.New("must be YYYY-MM-DD or RFC3339")
	default:
		return time.Time{}, fmt.Errorf("unsupported value %T", value)
	}
}

func hrToolString(value any) string {
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

func formatEmployeeToolDate(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.Format("2006-01-02")
}

func formatHRToolTimestamp(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC().Format(time.RFC3339)
}
