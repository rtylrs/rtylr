package insights

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	insightsapp "github.com/voxire/onion/internal/modules/insights/internal/application"
)

func scheduleReportSlug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.NewReplacer(
		"_", "-",
		" ", "-",
		".", "-",
		"/", "-",
		":", "-",
	).Replace(value)
	for strings.Contains(value, "--") {
		value = strings.ReplaceAll(value, "--", "-")
	}
	return strings.Trim(value, "-")
}

func insightsToolDateRange(input map[string]any) (time.Time, time.Time, error) {
	from := insightsToolFirstString(input, "from", "start", "start_date", "date_from")
	to := insightsToolFirstString(input, "to", "end", "end_date", "date_to")
	return insightsapp.ResolveDateRange(from, to)
}

func insightsToolRangePayload(from, to time.Time) map[string]string {
	return map[string]string{
		"from": from.Format("2006-01-02"),
		"to":   to.Format("2006-01-02"),
	}
}

func insightsToolFirstString(input map[string]any, keys ...string) string {
	for _, key := range keys {
		value := strings.TrimSpace(insightsToolString(input[key]))
		if value != "" {
			return value
		}
	}
	return ""
}

func insightsToolString(value any) string {
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

func insightsToolInt32(input map[string]any, fallback int32, keys ...string) (int32, error) {
	for _, key := range keys {
		value, ok := input[key]
		if !ok || value == nil {
			continue
		}
		n, err := insightsToolNumericInt32(value)
		if err != nil {
			return 0, fmt.Errorf("invalid %s: %w", key, err)
		}
		if n <= 0 {
			return fallback, nil
		}
		return n, nil
	}
	return fallback, nil
}

func insightsToolNumericInt32(value any) (int32, error) {
	switch v := value.(type) {
	case int:
		return int32(v), nil // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	case int8:
		return int32(v), nil
	case int16:
		return int32(v), nil
	case int32:
		return v, nil
	case int64:
		return int32(v), nil // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	case uint:
		return int32(v), nil // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	case uint8:
		return int32(v), nil
	case uint16:
		return int32(v), nil
	case uint32:
		return int32(v), nil // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	case uint64:
		return int32(v), nil // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	case float32:
		return int32(v), nil
	case float64:
		return int32(v), nil
	case string:
		n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 32)
		if err != nil {
			return 0, err
		}
		return int32(n), nil
	case json.Number:
		n, err := strconv.ParseInt(v.String(), 10, 32)
		if err != nil {
			return 0, err
		}
		return int32(n), nil
	default:
		return 0, fmt.Errorf("expected integer, got %T", value)
	}
}
