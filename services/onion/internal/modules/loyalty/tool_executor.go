package loyalty

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/loyalty/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	getLoyaltyConfigToolName          = "loyalty.get_config"
	listLoyaltyMembersToolName        = "loyalty.list_members"
	listLoyaltyRewardsToolName        = "loyalty.list_rewards"
	listLoyaltyTransactionsToolName   = "loyalty.list_transactions"
	defaultListLoyaltyToolLimit       = int64(20)
	maxListLoyaltyToolLimit           = int64(50)
	maxListLoyaltyMembersServiceLimit = int64(200)
	loyaltyToolTimestampDate          = "2006-01-02T15:04:05Z"
)

type listLoyaltyTransactionsToolInput struct {
	CustomerID int64
	Page       int64
	Limit      int64
}

type listLoyaltyMembersToolInput struct {
	Page  int64
	Limit int64
}

type listLoyaltyRewardsToolInput struct {
	ActiveOnly bool
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case getLoyaltyConfigToolName:
		output, err := m.executeGetLoyaltyConfigTool(ctx)
		return output, true, err
	case listLoyaltyMembersToolName:
		output, err := m.executeListLoyaltyMembersTool(ctx, input)
		return output, true, err
	case listLoyaltyRewardsToolName:
		output, err := m.executeListLoyaltyRewardsTool(ctx, input)
		return output, true, err
	case listLoyaltyTransactionsToolName:
		output, err := m.executeListLoyaltyTransactionsTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeGetLoyaltyConfigTool(ctx context.Context) (map[string]any, error) {
	businessID, err := loyaltyToolBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	config, err := m.service.GetConfig(ctx, businessID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"config": loyaltyConfigToolPayload(config),
	}, nil
}

func (m *Module) executeListLoyaltyMembersTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	businessID, err := loyaltyToolBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	parsed, err := parseListLoyaltyMembersToolInput(input)
	if err != nil {
		return nil, err
	}
	offset := int((parsed.Page - 1) * parsed.Limit)

	members, err := m.service.ListMembers(ctx, businessID, int(parsed.Limit), offset)
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(members))
	for _, member := range members {
		payloads = append(payloads, loyaltyMemberToolPayload(member))
	}

	return map[string]any{
		"members":   payloads,
		"returned":  len(payloads),
		"page":      parsed.Page,
		"page_size": parsed.Limit,
		"has_more":  len(payloads) == int(parsed.Limit),
	}, nil
}

func (m *Module) executeListLoyaltyRewardsTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	businessID, err := loyaltyToolBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	parsed, err := parseListLoyaltyRewardsToolInput(input)
	if err != nil {
		return nil, err
	}

	rewards, err := m.service.ListRewards(ctx, businessID, parsed.ActiveOnly)
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(rewards))
	for _, reward := range rewards {
		payloads = append(payloads, loyaltyRewardToolPayload(reward))
	}

	return map[string]any{
		"rewards":     payloads,
		"total":       len(payloads),
		"returned":    len(payloads),
		"active_only": parsed.ActiveOnly,
	}, nil
}

func (m *Module) executeListLoyaltyTransactionsTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	businessID, err := loyaltyToolBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	parsed, err := parseListLoyaltyTransactionsToolInput(input)
	if err != nil {
		return nil, err
	}

	transactions, total, err := m.service.ListTransactions(ctx, businessID, parsed.CustomerID, int(parsed.Page), int(parsed.Limit))
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(transactions))
	for _, transaction := range transactions {
		payloads = append(payloads, loyaltyTransactionToolPayload(transaction))
	}

	return map[string]any{
		"transactions": payloads,
		"total":        total,
		"returned":     len(payloads),
		"page":         parsed.Page,
		"page_size":    parsed.Limit,
		"has_more":     parsed.Page*parsed.Limit < int64(total),
		"filters": map[string]any{
			"customer_id": parsed.CustomerID,
		},
	}, nil
}

func parseListLoyaltyTransactionsToolInput(input map[string]any) (listLoyaltyTransactionsToolInput, error) {
	customerID, err := loyaltyToolNonNegativeInt64(firstLoyaltyToolValue(input, "customer_id", "customer"))
	if err != nil {
		return listLoyaltyTransactionsToolInput{}, fmt.Errorf("invalid customer_id: %w", err)
	}

	page, limit, err := parseLoyaltyToolPageLimit(input, maxListLoyaltyToolLimit)
	if err != nil {
		return listLoyaltyTransactionsToolInput{}, err
	}

	return listLoyaltyTransactionsToolInput{CustomerID: customerID, Page: page, Limit: limit}, nil
}

func parseListLoyaltyMembersToolInput(input map[string]any) (listLoyaltyMembersToolInput, error) {
	page, limit, err := parseLoyaltyToolPageLimit(input, maxListLoyaltyMembersServiceLimit)
	if err != nil {
		return listLoyaltyMembersToolInput{}, err
	}
	return listLoyaltyMembersToolInput{Page: page, Limit: limit}, nil
}

func parseListLoyaltyRewardsToolInput(input map[string]any) (listLoyaltyRewardsToolInput, error) {
	activeOnly, err := loyaltyToolBool(firstLoyaltyToolValue(input, "active_only", "active"))
	if err != nil {
		return listLoyaltyRewardsToolInput{}, fmt.Errorf("invalid active_only: %w", err)
	}
	return listLoyaltyRewardsToolInput{ActiveOnly: activeOnly}, nil
}

func parseLoyaltyToolPageLimit(input map[string]any, maxLimit int64) (int64, int64, error) {
	page, err := loyaltyToolNonNegativeInt64(firstLoyaltyToolValue(input, "page"))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid page: %w", err)
	}
	if page == 0 {
		page = 1
	}

	limit, err := loyaltyToolNonNegativeInt64(firstLoyaltyToolValue(input, "limit", "page_size"))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultListLoyaltyToolLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	return page, limit, nil
}

// loyaltyToolBusinessID resolves the caller's business scope; loyalty reads
// keyed by user id would leak cross-tenant balances on id collisions.
func loyaltyToolBusinessID(ctx context.Context) (int64, error) {
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

func loyaltyConfigToolPayload(config domain.LoyaltyConfig) map[string]any {
	return map[string]any{
		"id":                    config.ID,
		"owner_id":              config.OwnerID,
		"points_per_dollar":     config.PointsPerDollar,
		"min_redemption_points": config.MinRedemptionPoints,
		"points_value":          config.PointsValue,
		"is_active":             config.IsActive,
		"created_at":            formatLoyaltyToolTime(config.CreatedAt),
		"updated_at":            formatLoyaltyToolTime(config.UpdatedAt),
	}
}

func loyaltyMemberToolPayload(member domain.LoyaltyMember) map[string]any {
	return map[string]any{
		"id":             member.ID,
		"owner_id":       member.OwnerID,
		"name":           member.Name,
		"points":         member.Points,
		"tier":           member.Tier,
		"visits":         member.Visits,
		"joined_at":      formatLoyaltyToolTime(member.JoinedAt),
		"last_active_at": formatLoyaltyToolTime(member.LastActiveAt),
	}
}

func loyaltyRewardToolPayload(reward domain.LoyaltyReward) map[string]any {
	return map[string]any{
		"id":          reward.ID,
		"owner_id":    reward.OwnerID,
		"name":        reward.Name,
		"description": reward.Description,
		"category":    reward.Category,
		"points_cost": reward.PointsCost,
		"stock":       loyaltyToolOptionalInt(reward.Stock),
		"icon":        reward.Icon,
		"active":      reward.Active,
		"created_at":  formatLoyaltyToolTime(reward.CreatedAt),
		"updated_at":  formatLoyaltyToolTime(reward.UpdatedAt),
	}
}

func loyaltyTransactionToolPayload(transaction domain.LoyaltyTransaction) map[string]any {
	return map[string]any{
		"id":             transaction.ID,
		"owner_id":       transaction.OwnerID,
		"customer_id":    transaction.CustomerID,
		"points":         transaction.Points,
		"type":           transaction.Type,
		"description":    transaction.Description,
		"reference_id":   loyaltyToolOptionalInt64(transaction.ReferenceID),
		"reference_type": transaction.ReferenceType,
		"created_at":     formatLoyaltyToolTime(transaction.CreatedAt),
	}
}

func firstLoyaltyToolValue(input map[string]any, keys ...string) any {
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

func loyaltyToolNonNegativeInt64(value any) (int64, error) {
	out, err := loyaltyToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func loyaltyToolInt64(value any) (int64, error) {
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

func loyaltyToolBool(value any) (bool, error) {
	switch v := value.(type) {
	case nil:
		return false, nil
	case bool:
		return v, nil
	case int:
		return loyaltyToolBoolInt(int64(v))
	case int64:
		return loyaltyToolBoolInt(v)
	case int32:
		return loyaltyToolBoolInt(int64(v))
	case float64:
		if math.Trunc(v) != v {
			return false, errors.New("must be 0 or 1")
		}
		return loyaltyToolBoolInt(int64(v))
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			return false, err
		}
		return loyaltyToolBoolInt(i)
	case string:
		raw := strings.ToLower(strings.TrimSpace(v))
		switch raw {
		case "", "false", "no", "0", "all", "any":
			return false, nil
		case "true", "yes", "1", "active", "enabled":
			return true, nil
		default:
			return false, errors.New("must be true or false")
		}
	default:
		return false, fmt.Errorf("unsupported value %T", value)
	}
}

func loyaltyToolBoolInt(value int64) (bool, error) {
	if value == 0 {
		return false, nil
	}
	if value == 1 {
		return true, nil
	}
	return false, errors.New("must be 0 or 1")
}

func loyaltyToolOptionalInt(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}

func loyaltyToolOptionalInt64(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}

func formatLoyaltyToolTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC().Format(loyaltyToolTimestampDate)
}
