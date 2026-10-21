package segments

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/crm/segments/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/pagination"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	listSegmentsToolName         = "crm.list_segments"
	defaultListSegmentsToolLimit = int64(20)
	maxListSegmentsToolLimit     = int64(50)
	segmentToolTimestampDate     = "2006-01-02T15:04:05Z"
)

type listSegmentsToolInput struct {
	Query string
	Page  int64
	Limit int64
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listSegmentsToolName:
		output, err := m.executeListSegmentsTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListSegmentsTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	businessID, err := segmentToolBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	parsed, err := parseListSegmentsToolInput(input)
	if err != nil {
		return nil, err
	}

	segments, err := m.service.List(ctx, businessID)
	if err != nil {
		return nil, err
	}
	segments = filterSegmentsForTool(segments, parsed)

	start := pagination.SliceStart(parsed.Page, parsed.Limit, len(segments))
	if start > len(segments) {
		start = len(segments)
	}
	end := start + int(parsed.Limit)
	if end > len(segments) {
		end = len(segments)
	}

	payloads := make([]map[string]any, 0, end-start)
	for _, segment := range segments[start:end] {
		payloads = append(payloads, segmentToolPayload(segment))
	}

	return map[string]any{
		"segments":  payloads,
		"total":     len(segments),
		"returned":  len(payloads),
		"page":      parsed.Page,
		"page_size": parsed.Limit,
		"has_more":  end < len(segments),
		"filters": map[string]any{
			"query": parsed.Query,
		},
	}, nil
}

func parseListSegmentsToolInput(input map[string]any) (listSegmentsToolInput, error) {
	query := strings.TrimSpace(segmentToolString(firstSegmentToolValue(input, "query", "q", "search", "name")))

	page, err := segmentToolNonNegativeInt64(firstSegmentToolValue(input, "page"))
	if err != nil {
		return listSegmentsToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page == 0 {
		page = 1
	}

	limit, err := segmentToolNonNegativeInt64(firstSegmentToolValue(input, "limit", "page_size"))
	if err != nil {
		return listSegmentsToolInput{}, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultListSegmentsToolLimit
	}
	if limit > maxListSegmentsToolLimit {
		limit = maxListSegmentsToolLimit
	}

	return listSegmentsToolInput{Query: query, Page: page, Limit: limit}, nil
}

// segmentToolBusinessID resolves the business id used to scope tenant data for
// the caller. Segments are keyed by business_id, never by user id — scoping by
// user id would both hide the caller's own segments and leak other tenants'
// segments whenever users.id collides with businesses.id.
func segmentToolBusinessID(ctx context.Context) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User != nil {
		resolved, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
		if err == nil && resolved > 0 {
			return resolved, nil
		}
	}
	if authUser, ok := authPkg.GetAuthUser(ctx); ok && authUser != nil {
		resolved, err := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
		if err == nil && resolved > 0 {
			return resolved, nil
		}
	}
	return 0, errors.New("cannot determine business context")
}

func filterSegmentsForTool(segments []domain.Segment, input listSegmentsToolInput) []domain.Segment {
	if input.Query == "" {
		return segments
	}

	query := strings.ToLower(input.Query)
	out := make([]domain.Segment, 0, len(segments))
	for _, segment := range segments {
		if strings.Contains(strings.ToLower(segment.Name), query) ||
			strings.Contains(strings.ToLower(segment.Description), query) {
			out = append(out, segment)
		}
	}
	return out
}

func segmentToolPayload(segment domain.Segment) map[string]any {
	return map[string]any{
		"id":             segment.ID,
		"owner_id":       segment.OwnerID,
		"name":           segment.Name,
		"description":    segment.Description,
		"filters":        segmentToolFilters(segment.Filters),
		"customer_count": segment.CustomerCount,
		"created_at":     formatSegmentToolTime(segment.CreatedAt),
		"updated_at":     formatSegmentToolTime(segment.UpdatedAt),
	}
}

func segmentToolFilters(filters []byte) any {
	if len(filters) == 0 {
		return map[string]any{}
	}
	var out any
	if err := json.Unmarshal(filters, &out); err != nil {
		return string(filters)
	}
	return out
}

func firstSegmentToolValue(input map[string]any, keys ...string) any {
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

func segmentToolNonNegativeInt64(value any) (int64, error) {
	out, err := segmentToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func segmentToolInt64(value any) (int64, error) {
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

func segmentToolString(value any) string {
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

func formatSegmentToolTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC().Format(segmentToolTimestampDate)
}
