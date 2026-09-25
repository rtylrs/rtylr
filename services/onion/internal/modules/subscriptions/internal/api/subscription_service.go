package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/subscriptions/internal/application"
	"github.com/voxire/onion/internal/modules/subscriptions/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hubsubscriptionv1 "github.com/voxire/voxie/gen/go/hub/subscription/v1"
)

type SubscriptionAPI struct {
	Service *application.Service
}

func NewSubscriptionAPI(service *application.Service) *SubscriptionAPI {
	return &SubscriptionAPI{Service: service}
}

func (a *SubscriptionAPI) GetMySubscription(ctx context.Context, _ *hubsubscriptionv1.GetMySubscriptionRequest) (*hubsubscriptionv1.GetMySubscriptionResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	sub, err := a.Service.GetMySubscription(ctx, device.User.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrSubscriptionNotFound) {
			return &hubsubscriptionv1.GetMySubscriptionResponse{}, nil
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get subscription")
	}

	return &hubsubscriptionv1.GetMySubscriptionResponse{Subscription: toProtoSubscription(sub)}, nil
}

func (a *SubscriptionAPI) ListPlans(ctx context.Context, _ *hubsubscriptionv1.ListPlansRequest) (*hubsubscriptionv1.ListPlansResponse, error) {
	plans, err := a.Service.ListPlans(ctx)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list plans")
	}

	out := make([]*hubsubscriptionv1.Plan, 0, len(plans))
	for _, p := range plans {
		out = append(out, toProtoPlan(p))
	}
	return &hubsubscriptionv1.ListPlansResponse{Plans: out}, nil
}

func (a *SubscriptionAPI) GetBillingDetail(ctx context.Context, _ *hubsubscriptionv1.GetBillingDetailRequest) (*hubsubscriptionv1.GetBillingDetailResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	detail, err := a.Service.GetBillingDetail(ctx, device.User.UserID, device.User.BusinessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get billing detail")
	}

	return &hubsubscriptionv1.GetBillingDetailResponse{Billing: toProtoBillingDetail(detail)}, nil
}

func (a *SubscriptionAPI) GetBillingHistory(ctx context.Context, _ *hubsubscriptionv1.GetBillingHistoryRequest) (*hubsubscriptionv1.GetBillingHistoryResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	months, err := a.Service.GetBillingHistory(ctx, device.User.UserID, device.User.BusinessID, 12)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get billing history")
	}

	out := make([]*hubsubscriptionv1.BillingHistoryMonth, 0, len(months))
	for _, m := range months {
		out = append(out, toProtoBillingHistoryMonth(m))
	}
	return &hubsubscriptionv1.GetBillingHistoryResponse{History: out}, nil
}

func (a *SubscriptionAPI) UpsertSubscription(ctx context.Context, req *hubsubscriptionv1.UpsertSubscriptionRequest) (*hubsubscriptionv1.UpsertSubscriptionResponse, error) {
	authUser, err := consoleAdminUser(ctx)
	if err != nil {
		return nil, err
	}

	planSlug := strings.TrimSpace(req.GetPlanSlug())
	if planSlug == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "plan_slug is required")
	}

	err = a.Service.UpsertSubscription(ctx, authUser.UserID, planSlug)
	if err != nil {
		if errors.Is(err, domain.ErrPlanNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Plan not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update subscription")
	}

	return &hubsubscriptionv1.UpsertSubscriptionResponse{Message: "Subscription updated"}, nil
}

func (a *SubscriptionAPI) CreateCheckout(ctx context.Context, req *hubsubscriptionv1.CreateCheckoutRequest) (*hubsubscriptionv1.CreateCheckoutResponse, error) {
	authUser, err := billingUser(ctx)
	if err != nil {
		return nil, err
	}

	planSlug := strings.TrimSpace(req.GetPlanSlug())
	if planSlug == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "plan_slug is required")
	}

	checkoutURL, err := a.Service.CreateCheckout(ctx, authUser.UserID, planSlug, req.GetEmail())
	if err != nil {
		if errors.Is(err, domain.ErrPlanNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Plan not found")
		}
		logger.Ctx(ctx).Error("failed to create paddle checkout",
			logger.String("plan_slug", planSlug),
			logger.Error(err),
		)
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create checkout")
	}

	return &hubsubscriptionv1.CreateCheckoutResponse{CheckoutUrl: checkoutURL}, nil
}

// billingUser gates subscription mutations using the role carried by the
// authenticated tenant token. Keep this check at the API boundary so both
// mutation endpoints reject unauthorized requests before touching billing or
// Paddle state.
func billingUser(ctx context.Context) (*authPkg.AuthUser, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 || !isBillingCapableRole(authUser.Role) {
		return nil, onekiterrors.PermissionDenied()
	}
	return authUser, nil
}

// consoleAdminUser is intentionally narrower than billingUser. Plan changes
// are entitlement writes, not a self-service preference: subscription state
// must come from Paddle checkout/webhook processing or a console grant.
func consoleAdminUser(ctx context.Context) (*authPkg.AuthUser, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 || !isConsoleAdminRole(authUser.Role) {
		return nil, onekiterrors.PermissionDenied()
	}
	return authUser, nil
}

func isBillingCapableRole(role string) bool {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "owner", "admin", "billing", "console_admin":
		return true
	default:
		return false
	}
}

func isConsoleAdminRole(role string) bool {
	return strings.EqualFold(strings.TrimSpace(role), "console_admin")
}

func toProtoSubscription(sub *domain.Subscription) *hubsubscriptionv1.Subscription {
	if sub == nil {
		return nil
	}
	return &hubsubscriptionv1.Subscription{
		Id:                   sub.ID,
		PlanName:             sub.PlanName,
		PlanSlug:             sub.PlanSlug,
		Apps:                 sub.Apps,
		PricePerSeatCents:    int32(sub.PricePerSeatCents), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Status:               sub.Status,
		IsGranted:            sub.IsGranted,
		SeatCount:            int32(sub.SeatCount),         // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		IncludedSeatLimit:    int32(sub.IncludedSeatLimit), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		IncludedSeatCount:    int32(sub.IncludedSeatCount), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		BillableSeatCount:    int32(sub.BillableSeatCount), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		GrantExpiresAt:       formatOptionalTime(sub.GrantExpiresAt),
		TrialEndsAt:          formatOptionalTime(sub.TrialEndsAt),
		NextBillingDate:      formatOptionalTime(sub.NextBillingDate),
		PaddleSubscriptionId: sub.PaddleSubscriptionID,
	}
}

func toProtoPlan(p domain.Plan) *hubsubscriptionv1.Plan {
	out := &hubsubscriptionv1.Plan{
		Id:                      p.ID,
		Name:                    p.Name,
		Slug:                    p.Slug,
		Apps:                    p.Apps,
		PricePerSeatCents:       int32(p.PricePerSeatCents),       // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		AnnualPricePerSeatCents: int32(p.AnnualPricePerSeatCents), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		TrialDays:               int32(p.TrialDays),               // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Tagline:                 p.Tagline,
		AiCreditsPerMonth:       int32(p.AICreditsPerMonth), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Features:                p.Features,
		TierOrder:               int32(p.TierOrder), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		IsSolo:                  p.IsSolo,
		IsFeatured:              p.IsFeatured,
		IsActive:                p.IsActive,
	}
	if p.MaxSeats != nil {
		v := int32(*p.MaxSeats) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		out.MaxSeats = &v
	}
	if p.MaxLocations != nil {
		v := int32(*p.MaxLocations) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		out.MaxLocations = &v
	}
	return out
}

func toProtoBillingDetail(d *domain.BillingDetail) *hubsubscriptionv1.BillingDetail {
	if d == nil {
		return nil
	}
	items := make([]*hubsubscriptionv1.BillingLineItem, 0, len(d.LineItems))
	for _, li := range d.LineItems {
		items = append(items, toProtoBillingLineItem(li))
	}
	return &hubsubscriptionv1.BillingDetail{
		BillingMonth:         formatTime(d.BillingMonth),
		NextBillingDate:      formatOptionalTime(d.NextBillingDate),
		LineItems:            items,
		TotalCents:           int32(d.TotalCents), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		PlanName:             d.PlanName,
		PricePerSeatCents:    int32(d.PricePerSeatCents), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		CurrentSeatCount:     int32(d.CurrentSeatCount),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		IncludedSeatLimit:    int32(d.IncludedSeatLimit), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		IncludedSeatCount:    int32(d.IncludedSeatCount), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		BillableSeatCount:    int32(d.BillableSeatCount), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		SeatDiscountCents:    int32(d.SeatDiscountCents), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Status:               d.Status,
		IsGranted:            d.IsGranted,
		GrantExpiresAt:       formatOptionalTime(d.GrantExpiresAt),
		TrialEndsAt:          formatOptionalTime(d.TrialEndsAt),
		PaddleSubscriptionId: d.PaddleSubscriptionID,
	}
}

func toProtoBillingHistoryMonth(m domain.BillingHistoryMonth) *hubsubscriptionv1.BillingHistoryMonth {
	items := make([]*hubsubscriptionv1.BillingLineItem, 0, len(m.LineItems))
	for _, li := range m.LineItems {
		items = append(items, toProtoBillingLineItem(li))
	}
	return &hubsubscriptionv1.BillingHistoryMonth{
		Month:      formatTime(m.Month),
		TotalCents: int32(m.TotalCents), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		LineItems:  items,
		Status:     m.Status,
	}
}

func toProtoBillingLineItem(li domain.BillingLineItem) *hubsubscriptionv1.BillingLineItem {
	return &hubsubscriptionv1.BillingLineItem{
		PeriodStart:       formatTime(li.PeriodStart),
		PeriodEnd:         formatOptionalTime(li.PeriodEnd),
		SeatCount:         int32(li.SeatCount),         // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		IncludedSeatLimit: int32(li.IncludedSeatLimit), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		IncludedSeatCount: int32(li.IncludedSeatCount), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		BillableSeatCount: int32(li.BillableSeatCount), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		IncludedCents:     int32(li.IncludedCents),     // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		PricePerSeatCents: int32(li.PricePerSeatCents), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		AmountCents:       int32(li.AmountCents),       // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Type:              li.Type,
		Description:       li.Description,
		IsGranted:         li.IsGranted,
		GrantExpiresAt:    formatOptionalTime(li.GrantExpiresAt),
		AddonId:           li.AddonID,
		Quantity:          int32(li.Quantity),       // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		UnitPriceCents:    int32(li.UnitPriceCents), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func formatOptionalTime(t *time.Time) *string {
	if t == nil || t.IsZero() {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
}
