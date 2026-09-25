package reviews

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/reviews/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/pagination"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	listReviewsToolName         = "hr.list_reviews"
	defaultListReviewsToolLimit = int64(20)
	maxListReviewsToolLimit     = int64(50)
	reviewToolTimestampDate     = "2006-01-02T15:04:05Z"
)

type listReviewsToolInput struct {
	Status  string
	Period  string
	StaffID int64
	Page    int64
	Limit   int64
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listReviewsToolName:
		output, err := m.executeListReviewsTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListReviewsTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	businessID, err := reviewToolBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	parsed, err := parseListReviewsToolInput(input)
	if err != nil {
		return nil, err
	}

	reviews, err := m.service.List(ctx, businessID)
	if err != nil {
		return nil, err
	}
	reviews = filterReviewsForTool(reviews, parsed)

	start := pagination.SliceStart(parsed.Page, parsed.Limit, len(reviews))
	if start > len(reviews) {
		start = len(reviews)
	}
	end := start + int(parsed.Limit)
	if end > len(reviews) {
		end = len(reviews)
	}

	payloads := make([]map[string]any, 0, end-start)
	for _, review := range reviews[start:end] {
		payloads = append(payloads, reviewToolPayload(review))
	}

	return map[string]any{
		"reviews":   payloads,
		"total":     len(reviews),
		"returned":  len(payloads),
		"page":      parsed.Page,
		"page_size": parsed.Limit,
		"has_more":  end < len(reviews),
		"filters": map[string]any{
			"status":   parsed.Status,
			"period":   parsed.Period,
			"staff_id": parsed.StaffID,
		},
	}, nil
}

func parseListReviewsToolInput(input map[string]any) (listReviewsToolInput, error) {
	status := strings.TrimSpace(reviewToolString(firstReviewToolValue(input, "status")))
	period := strings.TrimSpace(reviewToolString(firstReviewToolValue(input, "period")))

	staffID, err := reviewToolNonNegativeInt64(firstReviewToolValue(input, "staff_id", "employee_id", "staff"))
	if err != nil {
		return listReviewsToolInput{}, fmt.Errorf("invalid staff_id: %w", err)
	}

	page, err := reviewToolNonNegativeInt64(firstReviewToolValue(input, "page"))
	if err != nil {
		return listReviewsToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page == 0 {
		page = 1
	}

	limit, err := reviewToolNonNegativeInt64(firstReviewToolValue(input, "limit", "page_size"))
	if err != nil {
		return listReviewsToolInput{}, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultListReviewsToolLimit
	}
	if limit > maxListReviewsToolLimit {
		limit = maxListReviewsToolLimit
	}

	return listReviewsToolInput{Status: status, Period: period, StaffID: staffID, Page: page, Limit: limit}, nil
}

func reviewToolBusinessID(ctx context.Context) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User != nil && device.User.UserID > 0 {
		businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
		if bizErr != nil {
			return 0, onekiterrors.PermissionDenied()
		}
		return businessID, nil
	}
	authUser, ok := authPkg.GetAuthUser(ctx)
	if ok && authUser != nil && authUser.UserID > 0 {
		businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
		if bizErr != nil {
			return 0, onekiterrors.PermissionDenied()
		}
		return businessID, nil
	}
	return 0, errors.New("authentication is required")
}

func filterReviewsForTool(reviews []domain.Review, input listReviewsToolInput) []domain.Review {
	if input.Status == "" && input.Period == "" && input.StaffID == 0 {
		return reviews
	}

	out := make([]domain.Review, 0, len(reviews))
	for _, review := range reviews {
		if input.Status != "" && !strings.EqualFold(review.Status, input.Status) {
			continue
		}
		if input.Period != "" && !strings.EqualFold(review.Period, input.Period) {
			continue
		}
		if input.StaffID > 0 && review.StaffID != input.StaffID {
			continue
		}
		out = append(out, review)
	}
	return out
}

func reviewToolPayload(review domain.Review) map[string]any {
	return map[string]any{
		"id":            review.ID,
		"owner_id":      review.OwnerID,
		"staff_id":      review.StaffID,
		"reviewer_id":   review.ReviewerID,
		"period":        review.Period,
		"rating":        review.Rating,
		"goals_rating":  review.GoalsRating,
		"values_rating": review.ValuesRating,
		"skills_rating": review.SkillsRating,
		"notes":         review.Notes,
		"status":        review.Status,
		"created_at":    formatReviewToolTime(review.CreatedAt),
		"updated_at":    formatReviewToolTime(review.UpdatedAt),
	}
}

func firstReviewToolValue(input map[string]any, keys ...string) any {
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

func reviewToolNonNegativeInt64(value any) (int64, error) {
	out, err := reviewToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func reviewToolInt64(value any) (int64, error) {
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

func reviewToolString(value any) string {
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

func formatReviewToolTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC().Format(reviewToolTimestampDate)
}
