package cashmovements

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/pos/cashmovements/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/pagination"
)

const (
	listCashMovementsToolName         = "pos.list_cash_movements"
	defaultListCashMovementsToolLimit = int64(20)
	maxListCashMovementsToolLimit     = int64(50)
	posCashToolTimestampDate          = "2006-01-02T15:04:05Z"
)

type listCashMovementsToolInput struct {
	SessionID  int64
	Type       string
	BranchID   int64
	TerminalID int64
	EmployeeID int64
	Page       int64
	Limit      int64
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listCashMovementsToolName:
		output, err := m.executeListCashMovementsTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListCashMovementsTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}
	if authUser.BusinessID <= 0 {
		return nil, errors.New("business context is required")
	}

	parsed, err := parseListCashMovementsToolInput(input)
	if err != nil {
		return nil, err
	}

	movements, err := m.service.ListBySession(ctx, authUser.BusinessID, parsed.SessionID)
	if err != nil {
		return nil, err
	}
	movements = filterCashMovementsForTool(movements, parsed)

	start := pagination.SliceStart(parsed.Page, parsed.Limit, len(movements))
	if start > len(movements) {
		start = len(movements)
	}
	end := start + int(parsed.Limit)
	if end > len(movements) {
		end = len(movements)
	}

	payloads := make([]map[string]any, 0, end-start)
	for _, movement := range movements[start:end] {
		payloads = append(payloads, cashMovementToolPayload(movement))
	}

	return map[string]any{
		"cash_movements": payloads,
		"total":          len(movements),
		"returned":       len(payloads),
		"page":           parsed.Page,
		"page_size":      parsed.Limit,
		"has_more":       end < len(movements),
		"filters": map[string]any{
			"session_id":  parsed.SessionID,
			"type":        parsed.Type,
			"branch_id":   parsed.BranchID,
			"terminal_id": parsed.TerminalID,
			"employee_id": parsed.EmployeeID,
		},
	}, nil
}

func parseListCashMovementsToolInput(input map[string]any) (listCashMovementsToolInput, error) {
	sessionID, err := posCashToolInt64(firstPosCashToolValue(input, "session_id", "session"))
	if err != nil {
		return listCashMovementsToolInput{}, fmt.Errorf("invalid session_id: %w", err)
	}
	if sessionID <= 0 {
		return listCashMovementsToolInput{}, errors.New("session_id is required")
	}

	movementType := normalizeCashMovementToolType(input["type"])
	if strings.TrimSpace(posCashToolString(input["type"])) != "" && movementType == "" {
		return listCashMovementsToolInput{}, errors.New("type must be one of drop, pickup, adjustment, float, tip_out, cash_in, or cash_out")
	}

	branchID, err := posCashToolNonNegativeInt64(firstPosCashToolValue(input, "branch_id", "location_id", "branch"))
	if err != nil {
		return listCashMovementsToolInput{}, fmt.Errorf("invalid branch_id: %w", err)
	}
	terminalID, err := posCashToolNonNegativeInt64(firstPosCashToolValue(input, "terminal_id", "terminal"))
	if err != nil {
		return listCashMovementsToolInput{}, fmt.Errorf("invalid terminal_id: %w", err)
	}
	employeeID, err := posCashToolNonNegativeInt64(firstPosCashToolValue(input, "employee_id", "staff_id", "employee"))
	if err != nil {
		return listCashMovementsToolInput{}, fmt.Errorf("invalid employee_id: %w", err)
	}

	page, err := posCashToolNonNegativeInt64(firstPosCashToolValue(input, "page"))
	if err != nil {
		return listCashMovementsToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page == 0 {
		page = 1
	}

	limit, err := posCashToolNonNegativeInt64(firstPosCashToolValue(input, "limit", "page_size"))
	if err != nil {
		return listCashMovementsToolInput{}, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultListCashMovementsToolLimit
	}
	if limit > maxListCashMovementsToolLimit {
		limit = maxListCashMovementsToolLimit
	}

	return listCashMovementsToolInput{
		SessionID:  sessionID,
		Type:       movementType,
		BranchID:   branchID,
		TerminalID: terminalID,
		EmployeeID: employeeID,
		Page:       page,
		Limit:      limit,
	}, nil
}

func filterCashMovementsForTool(movements []domain.CashMovement, input listCashMovementsToolInput) []domain.CashMovement {
	if input.Type == "" && input.BranchID == 0 && input.TerminalID == 0 && input.EmployeeID == 0 {
		return movements
	}

	out := make([]domain.CashMovement, 0, len(movements))
	for _, movement := range movements {
		if input.Type != "" && !strings.EqualFold(movement.Type, input.Type) {
			continue
		}
		if input.BranchID > 0 && movement.BranchID != input.BranchID {
			continue
		}
		if input.TerminalID > 0 && (movement.TerminalID == nil || *movement.TerminalID != input.TerminalID) {
			continue
		}
		if input.EmployeeID > 0 && movement.EmployeeID != input.EmployeeID {
			continue
		}
		out = append(out, movement)
	}
	return out
}

func cashMovementToolPayload(movement domain.CashMovement) map[string]any {
	return map[string]any{
		"id":          movement.ID,
		"business_id": movement.BusinessID,
		"branch_id":   movement.BranchID,
		"terminal_id": posCashToolOptionalID(movement.TerminalID),
		"session_id":  movement.SessionID,
		"employee_id": movement.EmployeeID,
		"type":        movement.Type,
		"amount":      movement.Amount,
		"reason":      posCashToolOptionalString(movement.Reason),
		"approved_by": posCashToolOptionalID(movement.ApprovedBy),
		"created_at":  formatPosCashToolTime(movement.CreatedAt),
	}
}

func normalizeCashMovementToolType(value any) string {
	movementType := strings.ToLower(strings.TrimSpace(posCashToolString(value)))
	movementType = strings.ReplaceAll(movementType, "-", "_")
	movementType = strings.ReplaceAll(movementType, " ", "_")
	switch movementType {
	case "", "all":
		return ""
	case "drop", "pickup", "adjustment", "float", "tip_out", "cash_in", "cash_out":
		return movementType
	default:
		return ""
	}
}

func firstPosCashToolValue(input map[string]any, keys ...string) any {
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

func posCashToolNonNegativeInt64(value any) (int64, error) {
	out, err := posCashToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func posCashToolInt64(value any) (int64, error) {
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

func posCashToolString(value any) string {
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

func posCashToolOptionalID(value *int64) any {
	if value == nil || *value <= 0 {
		return nil
	}
	return *value
}

func posCashToolOptionalString(value *string) any {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	return *value
}

func formatPosCashToolTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC().Format(posCashToolTimestampDate)
}
