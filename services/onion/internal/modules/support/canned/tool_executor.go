package supportcanned

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/support/canned/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/pagination"
)

const (
	listSupportCannedResponsesToolName         = "support.list_canned_responses"
	defaultListSupportCannedResponsesToolLimit = int64(20)
	maxListSupportCannedResponsesToolLimit     = int64(50)
	supportCannedToolTimestampDate             = "2006-01-02T15:04:05Z"
)

type listSupportCannedResponsesToolInput struct {
	Query    string
	Category string
	Page     int64
	Limit    int64
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listSupportCannedResponsesToolName:
		output, err := m.executeListSupportCannedResponsesTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListSupportCannedResponsesTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	businessID, err := supportCannedToolBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	parsed, err := parseListSupportCannedResponsesToolInput(input)
	if err != nil {
		return nil, err
	}

	responses, err := m.service.List(ctx, businessID)
	if err != nil {
		return nil, err
	}
	responses = filterSupportCannedResponsesForTool(responses, parsed)

	start := pagination.SliceStart(parsed.Page, parsed.Limit, len(responses))
	if start > len(responses) {
		start = len(responses)
	}
	end := start + int(parsed.Limit)
	if end > len(responses) {
		end = len(responses)
	}

	payloads := make([]map[string]any, 0, end-start)
	for _, response := range responses[start:end] {
		payloads = append(payloads, supportCannedResponseToolPayload(response))
	}

	return map[string]any{
		"canned_responses": payloads,
		"total":            len(responses),
		"returned":         len(payloads),
		"page":             parsed.Page,
		"page_size":        parsed.Limit,
		"has_more":         end < len(responses),
		"filters": map[string]any{
			"query":    parsed.Query,
			"category": parsed.Category,
		},
	}, nil
}

func parseListSupportCannedResponsesToolInput(input map[string]any) (listSupportCannedResponsesToolInput, error) {
	query := strings.TrimSpace(supportCannedToolString(firstSupportCannedToolValue(input, "query", "q", "search", "name", "shortcut")))
	category := strings.TrimSpace(supportCannedToolString(firstSupportCannedToolValue(input, "category")))

	page, err := supportCannedToolNonNegativeInt64(firstSupportCannedToolValue(input, "page"))
	if err != nil {
		return listSupportCannedResponsesToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page == 0 {
		page = 1
	}

	limit, err := supportCannedToolNonNegativeInt64(firstSupportCannedToolValue(input, "limit", "page_size"))
	if err != nil {
		return listSupportCannedResponsesToolInput{}, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultListSupportCannedResponsesToolLimit
	}
	if limit > maxListSupportCannedResponsesToolLimit {
		limit = maxListSupportCannedResponsesToolLimit
	}

	return listSupportCannedResponsesToolInput{
		Query:    query,
		Category: category,
		Page:     page,
		Limit:    limit,
	}, nil
}

func supportCannedToolBusinessID(ctx context.Context) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User != nil && device.User.BusinessID > 0 {
		return device.User.BusinessID, nil
	}
	authUser, ok := authPkg.GetAuthUser(ctx)
	if ok && authUser != nil && authUser.BusinessID > 0 {
		return authUser.BusinessID, nil
	}
	return 0, errors.New("authentication is required")
}

func filterSupportCannedResponsesForTool(responses []domain.CannedResponse, input listSupportCannedResponsesToolInput) []domain.CannedResponse {
	if input.Query == "" && input.Category == "" {
		return responses
	}

	query := strings.ToLower(input.Query)
	category := strings.ToLower(input.Category)
	out := make([]domain.CannedResponse, 0, len(responses))
	for _, response := range responses {
		if category != "" && strings.ToLower(response.Category) != category {
			continue
		}
		if query != "" && !supportCannedResponseToolMatchesQuery(response, query) {
			continue
		}
		out = append(out, response)
	}
	return out
}

func supportCannedResponseToolMatchesQuery(response domain.CannedResponse, query string) bool {
	return strings.Contains(strings.ToLower(response.Name), query) ||
		strings.Contains(strings.ToLower(response.Body), query) ||
		strings.Contains(strings.ToLower(response.Category), query) ||
		strings.Contains(strings.ToLower(response.Shortcut), query)
}

func supportCannedResponseToolPayload(response domain.CannedResponse) map[string]any {
	return map[string]any{
		"id":          response.ID,
		"business_id": response.BusinessID,
		"name":        response.Name,
		"body":        response.Body,
		"category":    response.Category,
		"shortcut":    response.Shortcut,
		"created_at":  formatSupportCannedToolTime(response.CreatedAt),
	}
}

func firstSupportCannedToolValue(input map[string]any, keys ...string) any {
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

func supportCannedToolNonNegativeInt64(value any) (int64, error) {
	out, err := supportCannedToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func supportCannedToolInt64(value any) (int64, error) {
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

func supportCannedToolString(value any) string {
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

func formatSupportCannedToolTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC().Format(supportCannedToolTimestampDate)
}
