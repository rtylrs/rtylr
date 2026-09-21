package pos_sessions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/pos_sessions/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	listSessionsToolName         = "pos.list_sessions"
	defaultListSessionsToolLimit = int64(20)
	maxListSessionsToolLimit     = int64(50)
	// maxListSessionsToolPage bounds the page so (page-1)*limit cannot leave
	// the int64 range before it is used as a slice offset.
	maxListSessionsToolPage     = int64(1_000_000)
	posSessionToolTimestampDate = "2006-01-02T15:04:05Z"
)

type listSessionsToolInput struct {
	Status     string
	LocationID int64
	StaffID    int64
	Page       int64
	Limit      int64
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listSessionsToolName:
		output, err := m.executeListSessionsTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListSessionsTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseListSessionsToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, errors.New("authentication is required")
	}
	sessions, err := m.service.List(ctx, businessID)
	if err != nil {
		return nil, err
	}
	sessions = filterSessionsForTool(sessions, parsed)

	start, end := listSessionsToolBounds(len(sessions), parsed.Page, parsed.Limit)

	payloads := make([]map[string]any, 0, end-start)
	for _, session := range sessions[start:end] {
		payloads = append(payloads, sessionToolPayload(session))
	}

	return map[string]any{
		"sessions":  payloads,
		"total":     len(sessions),
		"returned":  len(payloads),
		"page":      parsed.Page,
		"page_size": parsed.Limit,
		"has_more":  end < len(sessions),
		"filters": map[string]any{
			"status":      parsed.Status,
			"location_id": parsed.LocationID,
			"staff_id":    parsed.StaffID,
		},
	}, nil
}

// listSessionsToolBounds turns a 1-based page into a slice range. The offset
// is computed in int64 and clamped before narrowing so an out-of-range page
// can never overflow into a negative slice index.
func listSessionsToolBounds(total int, page, limit int64) (int, int) {
	start64 := (page - 1) * limit
	if start64 < 0 || start64 > int64(total) {
		start64 = int64(total)
	}
	end64 := start64 + limit
	if end64 < start64 || end64 > int64(total) {
		end64 = int64(total)
	}
	return int(start64), int(end64)
}

func parseListSessionsToolInput(input map[string]any) (listSessionsToolInput, error) {
	status := normalizeSessionToolStatus(input["status"])
	if strings.TrimSpace(posSessionToolString(input["status"])) != "" && status == "" {
		return listSessionsToolInput{}, errors.New("status must be open or closed")
	}

	locationID, err := posSessionToolNonNegativeInt64(firstPosSessionToolValue(input, "location_id", "branch_id", "location"))
	if err != nil {
		return listSessionsToolInput{}, fmt.Errorf("invalid location_id: %w", err)
	}
	staffID, err := posSessionToolNonNegativeInt64(firstPosSessionToolValue(input, "staff_id", "employee_id", "staff"))
	if err != nil {
		return listSessionsToolInput{}, fmt.Errorf("invalid staff_id: %w", err)
	}

	page, err := posSessionToolNonNegativeInt64(firstPosSessionToolValue(input, "page"))
	if err != nil {
		return listSessionsToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page == 0 {
		page = 1
	}
	if page > maxListSessionsToolPage {
		return listSessionsToolInput{}, fmt.Errorf("page must be at most %d", maxListSessionsToolPage)
	}

	limit, err := posSessionToolNonNegativeInt64(firstPosSessionToolValue(input, "limit", "page_size"))
	if err != nil {
		return listSessionsToolInput{}, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultListSessionsToolLimit
	}
	if limit > maxListSessionsToolLimit {
		limit = maxListSessionsToolLimit
	}

	return listSessionsToolInput{
		Status:     status,
		LocationID: locationID,
		StaffID:    staffID,
		Page:       page,
		Limit:      limit,
	}, nil
}

func filterSessionsForTool(sessions []domain.Session, input listSessionsToolInput) []domain.Session {
	if input.Status == "" && input.LocationID == 0 && input.StaffID == 0 {
		return sessions
	}

	out := make([]domain.Session, 0, len(sessions))
	for _, session := range sessions {
		if input.Status != "" && !strings.EqualFold(session.Status, input.Status) {
			continue
		}
		if input.LocationID > 0 && (session.LocationID == nil || *session.LocationID != input.LocationID) {
			continue
		}
		if input.StaffID > 0 && (session.StaffID == nil || *session.StaffID != input.StaffID) {
			continue
		}
		out = append(out, session)
	}
	return out
}

func sessionToolPayload(session domain.Session) map[string]any {
	return map[string]any{
		"id":            session.ID,
		"owner_id":      session.OwnerID,
		"location_id":   posSessionToolOptionalID(session.LocationID),
		"staff_id":      posSessionToolOptionalID(session.StaffID),
		"opening_float": session.OpeningFloat,
		"closing_float": session.ClosingFloat,
		"total_sales":   session.TotalSales,
		"status":        session.Status,
		"opened_at":     formatPosSessionToolTime(session.OpenedAt),
		"closed_at":     formatPosSessionToolOptionalTime(session.ClosedAt),
		"created_at":    formatPosSessionToolTime(session.CreatedAt),
	}
}

func normalizeSessionToolStatus(value any) string {
	status := strings.ToLower(strings.TrimSpace(posSessionToolString(value)))
	status = strings.ReplaceAll(status, "-", "_")
	status = strings.ReplaceAll(status, " ", "_")
	switch status {
	case "", "all":
		return ""
	case "open", "closed":
		return status
	default:
		return ""
	}
}

func firstPosSessionToolValue(input map[string]any, keys ...string) any {
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

func posSessionToolNonNegativeInt64(value any) (int64, error) {
	out, err := posSessionToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func posSessionToolInt64(value any) (int64, error) {
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

func posSessionToolString(value any) string {
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

func posSessionToolOptionalID(value *int64) any {
	if value == nil || *value <= 0 {
		return nil
	}
	return *value
}

func formatPosSessionToolOptionalTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return formatPosSessionToolTime(*value)
}

func formatPosSessionToolTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC().Format(posSessionToolTimestampDate)
}
