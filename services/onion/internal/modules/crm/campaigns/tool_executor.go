package campaigns

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/crm/campaigns/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/pagination"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	listCampaignsToolName         = "crm.list_campaigns"
	getCampaignToolName           = "crm.get_campaign"
	defaultListCampaignsToolLimit = int64(20)
	maxListCampaignsToolLimit     = int64(50)
	campaignToolTimestampDate     = "2006-01-02T15:04:05Z"
)

type listCampaignsToolInput struct {
	Query     string
	Status    string
	Type      string
	SegmentID int64
	Page      int64
	Limit     int64
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listCampaignsToolName:
		output, err := m.executeListCampaignsTool(ctx, input)
		return output, true, err
	case getCampaignToolName:
		output, err := m.executeGetCampaignTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListCampaignsTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	parsed, err := parseListCampaignsToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, err := campaignToolBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	campaigns, err := m.service.List(ctx, businessID)
	if err != nil {
		return nil, err
	}
	campaigns = filterCampaignsForTool(campaigns, parsed)

	start := pagination.SliceStart(parsed.Page, parsed.Limit, len(campaigns))
	if start > len(campaigns) {
		start = len(campaigns)
	}
	end := start + int(parsed.Limit)
	if end > len(campaigns) {
		end = len(campaigns)
	}

	payloads := make([]map[string]any, 0, end-start)
	for _, campaign := range campaigns[start:end] {
		payloads = append(payloads, campaignToolPayload(campaign))
	}

	return map[string]any{
		"campaigns": payloads,
		"total":     len(campaigns),
		"returned":  len(payloads),
		"page":      parsed.Page,
		"page_size": parsed.Limit,
		"has_more":  end < len(campaigns),
		"filters": map[string]any{
			"query":      parsed.Query,
			"status":     parsed.Status,
			"type":       parsed.Type,
			"segment_id": parsed.SegmentID,
		},
	}, nil
}

func (m *Module) executeGetCampaignTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	campaignID, err := campaignToolInt64(firstCampaignToolValue(input, "campaign_id", "id"))
	if err != nil {
		return nil, fmt.Errorf("invalid campaign_id: %w", err)
	}
	if campaignID <= 0 {
		return nil, errors.New("campaign_id is required")
	}

	businessID, err := campaignToolBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	campaign, err := m.service.GetByID(ctx, businessID, campaignID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"campaign": campaignToolPayload(campaign),
	}, nil
}

// campaignToolBusinessID resolves the caller's business scope. Campaign data
// is keyed by business_id — scoping by user id leaks other tenants' campaigns
// whenever users.id collides with businesses.id.
func campaignToolBusinessID(ctx context.Context) (int64, error) {
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

func parseListCampaignsToolInput(input map[string]any) (listCampaignsToolInput, error) {
	query := strings.TrimSpace(campaignToolString(firstCampaignToolValue(input, "query", "q", "search", "name")))
	status := strings.TrimSpace(campaignToolString(firstCampaignToolValue(input, "status")))
	campaignType := strings.TrimSpace(campaignToolString(firstCampaignToolValue(input, "type", "channel")))

	segmentID, err := campaignToolNonNegativeInt64(firstCampaignToolValue(input, "segment_id", "audience_segment_id"))
	if err != nil {
		return listCampaignsToolInput{}, fmt.Errorf("invalid segment_id: %w", err)
	}

	page, err := campaignToolNonNegativeInt64(firstCampaignToolValue(input, "page"))
	if err != nil {
		return listCampaignsToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page == 0 {
		page = 1
	}

	limit, err := campaignToolNonNegativeInt64(firstCampaignToolValue(input, "limit", "page_size"))
	if err != nil {
		return listCampaignsToolInput{}, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultListCampaignsToolLimit
	}
	if limit > maxListCampaignsToolLimit {
		limit = maxListCampaignsToolLimit
	}

	return listCampaignsToolInput{
		Query:     query,
		Status:    status,
		Type:      campaignType,
		SegmentID: segmentID,
		Page:      page,
		Limit:     limit,
	}, nil
}

func filterCampaignsForTool(campaigns []domain.Campaign, input listCampaignsToolInput) []domain.Campaign {
	if input.Query == "" && input.Status == "" && input.Type == "" && input.SegmentID == 0 {
		return campaigns
	}

	query := strings.ToLower(input.Query)
	out := make([]domain.Campaign, 0, len(campaigns))
	for _, campaign := range campaigns {
		if input.Status != "" && !strings.EqualFold(campaign.Status, input.Status) {
			continue
		}
		if input.Type != "" && !strings.EqualFold(campaign.Type, input.Type) {
			continue
		}
		if input.SegmentID > 0 && (campaign.AudienceSegmentID == nil || *campaign.AudienceSegmentID != input.SegmentID) {
			continue
		}
		if query != "" && !campaignToolMatchesQuery(campaign, query) {
			continue
		}
		out = append(out, campaign)
	}
	return out
}

func campaignToolMatchesQuery(campaign domain.Campaign, query string) bool {
	return strings.Contains(strings.ToLower(campaign.Name), query) ||
		strings.Contains(strings.ToLower(campaign.Subject), query) ||
		strings.Contains(strings.ToLower(campaign.Body), query)
}

func campaignToolPayload(campaign domain.Campaign) map[string]any {
	return map[string]any{
		"id":                  campaign.ID,
		"owner_id":            campaign.OwnerID,
		"name":                campaign.Name,
		"type":                campaign.Type,
		"status":              campaign.Status,
		"audience_segment_id": campaignToolOptionalInt64(campaign.AudienceSegmentID),
		"subject":             campaign.Subject,
		"body":                campaign.Body,
		"sent_count":          campaign.SentCount,
		"open_count":          campaign.OpenCount,
		"open_rate":           campaignToolOpenRate(campaign),
		"scheduled_at":        formatCampaignToolOptionalTime(campaign.ScheduledAt),
		"sent_at":             formatCampaignToolOptionalTime(campaign.SentAt),
		"created_at":          formatCampaignToolTime(campaign.CreatedAt),
		"updated_at":          formatCampaignToolTime(campaign.UpdatedAt),
	}
}

func campaignToolOpenRate(campaign domain.Campaign) any {
	if campaign.SentCount <= 0 {
		return nil
	}
	return float64(campaign.OpenCount) / float64(campaign.SentCount)
}

func firstCampaignToolValue(input map[string]any, keys ...string) any {
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

func campaignToolNonNegativeInt64(value any) (int64, error) {
	out, err := campaignToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func campaignToolInt64(value any) (int64, error) {
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

func campaignToolString(value any) string {
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

func campaignToolOptionalInt64(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}

func formatCampaignToolOptionalTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return formatCampaignToolTime(*value)
}

func formatCampaignToolTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC().Format(campaignToolTimestampDate)
}
