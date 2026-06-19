package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/subscriptions/internal/application"
	"github.com/voxire/onion/internal/modules/subscriptions/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	paddlePkg "github.com/voxire/onion/internal/platform/paddle"
)

const maxPaddleWebhookBodyBytes = 1 << 20 // 1 MiB

// ─── List Plans ───────────────────────────────────────────────────────────────

func ListPlansRoutes() []string { return []string{"/subscriptions/plans"} }

type (
	ListPlansHandler struct{ Service *application.Service }
	ListPlansRequest struct{}
)

type PlanPayload struct {
	ID                      int64    `json:"id"`
	Name                    string   `json:"name"`
	Slug                    string   `json:"slug"`
	Apps                    []string `json:"apps"`
	PricePerSeatCents       int      `json:"price_per_seat_cents"`
	AnnualPricePerSeatCents int      `json:"annual_price_per_seat_cents"`
	TrialDays               int      `json:"trial_days"`
	Tagline                 string   `json:"tagline"`
	MaxSeats                *int     `json:"max_seats"`
	MaxLocations            *int     `json:"max_locations"`
	AICreditsPerMonth       int      `json:"ai_credits_per_month"`
	Features                []string `json:"features"`
	TierOrder               int      `json:"tier_order"`
	IsSolo                  bool     `json:"is_solo"`
	IsFeatured              bool     `json:"is_featured"`
	IsActive                bool     `json:"is_active"`
}

type ListPlansResponse struct {
	Plans []PlanPayload `json:"plans"`
}

func (h *ListPlansHandler) Serve(ctx context.Context, _ *ListPlansRequest) (ListPlansResponse, *errorModels.ServerError) {
	plans, err := h.Service.ListPlans(ctx)
	if err != nil {
		return ListPlansResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to list plans",
		}
	}

	payloads := make([]PlanPayload, 0, len(plans))
	for _, p := range plans {
		payloads = append(payloads, PlanPayload{
			ID:                      p.ID,
			Name:                    p.Name,
			Slug:                    p.Slug,
			Apps:                    p.Apps,
			PricePerSeatCents:       p.PricePerSeatCents,
			AnnualPricePerSeatCents: p.AnnualPricePerSeatCents,
			TrialDays:               p.TrialDays,
			Tagline:                 p.Tagline,
			MaxSeats:                p.MaxSeats,
			MaxLocations:            p.MaxLocations,
			AICreditsPerMonth:       p.AICreditsPerMonth,
			Features:                p.Features,
			TierOrder:               p.TierOrder,
			IsSolo:                  p.IsSolo,
			IsFeatured:              p.IsFeatured,
			IsActive:                p.IsActive,
		})
	}

	return ListPlansResponse{Plans: payloads}, nil
}

// ─── My Subscription ──────────────────────────────────────────────────────────

func MySubscriptionRoutes() []string { return []string{"/subscriptions/my"} }

type (
	MySubscriptionHandler struct{ Service *application.Service }
	MySubscriptionRequest struct{}
)

type SubscriptionPayload struct {
	ID                   int64      `json:"id"`
	PlanName             string     `json:"plan_name"`
	PlanSlug             string     `json:"plan_slug"`
	Apps                 []string   `json:"apps"`
	PricePerSeatCents    int        `json:"price_per_seat_cents"`
	Status               string     `json:"status"`
	SeatCount            int        `json:"seat_count"`
	IncludedSeatLimit    int        `json:"included_seat_limit"`
	IncludedSeatCount    int        `json:"included_seat_count"`
	BillableSeatCount    int        `json:"billable_seat_count"`
	TrialEndsAt          *time.Time `json:"trial_ends_at,omitempty"`
	NextBillingDate      *time.Time `json:"next_billing_date,omitempty"`
	PaddleSubscriptionID string     `json:"paddle_subscription_id,omitempty"`
}

type MySubscriptionResponse struct {
	Subscription *SubscriptionPayload `json:"subscription"`
}

func (h *MySubscriptionHandler) Serve(ctx context.Context, _ *MySubscriptionRequest) (MySubscriptionResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return MySubscriptionResponse{}, errorModels.PermissionDeniedServerError()
	}

	sub, err := h.Service.GetMySubscription(ctx, device.User.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrSubscriptionNotFound) {
			return MySubscriptionResponse{Subscription: nil}, nil
		}
		return MySubscriptionResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to get subscription",
		}
	}

	return MySubscriptionResponse{
		Subscription: &SubscriptionPayload{
			ID:                   sub.ID,
			PlanName:             sub.PlanName,
			PlanSlug:             sub.PlanSlug,
			Apps:                 sub.Apps,
			PricePerSeatCents:    sub.PricePerSeatCents,
			Status:               sub.Status,
			SeatCount:            sub.SeatCount,
			IncludedSeatLimit:    sub.IncludedSeatLimit,
			IncludedSeatCount:    sub.IncludedSeatCount,
			BillableSeatCount:    sub.BillableSeatCount,
			TrialEndsAt:          sub.TrialEndsAt,
			NextBillingDate:      sub.NextBillingDate,
			PaddleSubscriptionID: sub.PaddleSubscriptionID,
		},
	}, nil
}

// ─── Billing Detail (current month) ──────────────────────────────────────────

func BillingDetailRoutes() []string { return []string{"/subscriptions/billing"} }

type (
	BillingDetailHandler struct{ Service *application.Service }
	BillingDetailRequest struct{}
)

type BillingLineItemPayload struct {
	PeriodStart       time.Time  `json:"period_start"`
	PeriodEnd         *time.Time `json:"period_end"`
	SeatCount         int        `json:"seat_count"`
	IncludedSeatLimit int        `json:"included_seat_limit"`
	IncludedSeatCount int        `json:"included_seat_count"`
	BillableSeatCount int        `json:"billable_seat_count"`
	IncludedCents     int        `json:"included_cents"`
	PricePerSeatCents int        `json:"price_per_seat_cents"`
	AmountCents       int        `json:"amount_cents"`
	Type              string     `json:"type"`
	Description       string     `json:"description"`
	AddonID           string     `json:"addon_id,omitempty"`
	Quantity          int        `json:"quantity"`
	UnitPriceCents    int        `json:"unit_price_cents"`
}

type BillingDetailPayload struct {
	BillingMonth         time.Time                `json:"billing_month"`
	NextBillingDate      *time.Time               `json:"next_billing_date"`
	LineItems            []BillingLineItemPayload `json:"line_items"`
	TotalCents           int                      `json:"total_cents"`
	PlanName             string                   `json:"plan_name"`
	PricePerSeatCents    int                      `json:"price_per_seat_cents"`
	CurrentSeatCount     int                      `json:"current_seat_count"`
	IncludedSeatLimit    int                      `json:"included_seat_limit"`
	IncludedSeatCount    int                      `json:"included_seat_count"`
	BillableSeatCount    int                      `json:"billable_seat_count"`
	SeatDiscountCents    int                      `json:"seat_discount_cents"`
	Status               string                   `json:"status"`
	TrialEndsAt          *time.Time               `json:"trial_ends_at"`
	PaddleSubscriptionID string                   `json:"paddle_subscription_id,omitempty"`
}

type BillingDetailResponse struct {
	Billing *BillingDetailPayload `json:"billing"`
}

func (h *BillingDetailHandler) Serve(ctx context.Context, _ *BillingDetailRequest) (BillingDetailResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return BillingDetailResponse{}, errorModels.PermissionDeniedServerError()
	}

	detail, err := h.Service.GetBillingDetail(ctx, device.User.UserID, device.User.BusinessID)
	if err != nil {
		if errors.Is(err, domain.ErrSubscriptionNotFound) {
			return BillingDetailResponse{Billing: nil}, nil
		}
		return BillingDetailResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to get billing detail",
		}
	}

	items := make([]BillingLineItemPayload, 0, len(detail.LineItems))
	for _, li := range detail.LineItems {
		items = append(items, BillingLineItemPayload{
			PeriodStart:       li.PeriodStart,
			PeriodEnd:         li.PeriodEnd,
			SeatCount:         li.SeatCount,
			IncludedSeatLimit: li.IncludedSeatLimit,
			IncludedSeatCount: li.IncludedSeatCount,
			BillableSeatCount: li.BillableSeatCount,
			IncludedCents:     li.IncludedCents,
			PricePerSeatCents: li.PricePerSeatCents,
			AmountCents:       li.AmountCents,
			Type:              li.Type,
			Description:       li.Description,
			AddonID:           li.AddonID,
			Quantity:          li.Quantity,
			UnitPriceCents:    li.UnitPriceCents,
		})
	}

	return BillingDetailResponse{
		Billing: &BillingDetailPayload{
			BillingMonth:         detail.BillingMonth,
			NextBillingDate:      detail.NextBillingDate,
			LineItems:            items,
			TotalCents:           detail.TotalCents,
			PlanName:             detail.PlanName,
			PricePerSeatCents:    detail.PricePerSeatCents,
			CurrentSeatCount:     detail.CurrentSeatCount,
			IncludedSeatLimit:    detail.IncludedSeatLimit,
			IncludedSeatCount:    detail.IncludedSeatCount,
			BillableSeatCount:    detail.BillableSeatCount,
			SeatDiscountCents:    detail.SeatDiscountCents,
			Status:               detail.Status,
			TrialEndsAt:          detail.TrialEndsAt,
			PaddleSubscriptionID: detail.PaddleSubscriptionID,
		},
	}, nil
}

// ─── Billing History ──────────────────────────────────────────────────────────

func BillingHistoryRoutes() []string { return []string{"/subscriptions/billing/history"} }

type (
	BillingHistoryHandler struct{ Service *application.Service }
	BillingHistoryRequest struct{}
)

type BillingMonthPayload struct {
	Month      time.Time                `json:"month"`
	TotalCents int                      `json:"total_cents"`
	LineItems  []BillingLineItemPayload `json:"line_items"`
	Status     string                   `json:"status"`
}

type BillingHistoryResponse struct {
	History []BillingMonthPayload `json:"history"`
}

func (h *BillingHistoryHandler) Serve(ctx context.Context, _ *BillingHistoryRequest) (BillingHistoryResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return BillingHistoryResponse{}, errorModels.PermissionDeniedServerError()
	}

	months, err := h.Service.GetBillingHistory(ctx, device.User.UserID, device.User.BusinessID, 12)
	if err != nil {
		return BillingHistoryResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to get billing history",
		}
	}

	payloads := make([]BillingMonthPayload, 0, len(months))
	for _, m := range months {
		items := make([]BillingLineItemPayload, 0, len(m.LineItems))
		for _, li := range m.LineItems {
			items = append(items, BillingLineItemPayload{
				PeriodStart:       li.PeriodStart,
				PeriodEnd:         li.PeriodEnd,
				SeatCount:         li.SeatCount,
				IncludedSeatLimit: li.IncludedSeatLimit,
				IncludedSeatCount: li.IncludedSeatCount,
				BillableSeatCount: li.BillableSeatCount,
				IncludedCents:     li.IncludedCents,
				PricePerSeatCents: li.PricePerSeatCents,
				AmountCents:       li.AmountCents,
				Type:              li.Type,
				Description:       li.Description,
				AddonID:           li.AddonID,
				Quantity:          li.Quantity,
				UnitPriceCents:    li.UnitPriceCents,
			})
		}
		payloads = append(payloads, BillingMonthPayload{
			Month:      m.Month,
			TotalCents: m.TotalCents,
			LineItems:  items,
			Status:     m.Status,
		})
	}

	return BillingHistoryResponse{History: payloads}, nil
}

// ─── Upsert Subscription ──────────────────────────────────────────────────────

func UpsertSubscriptionRoutes() []string { return []string{"/subscriptions/my"} }

type UpsertSubscriptionHandler struct{ Service *application.Service }

type UpsertSubscriptionRequest struct {
	PlanSlug string `json:"plan_slug"`
}

type UpsertSubscriptionResponse struct {
	Message string `json:"message"`
}

func (h *UpsertSubscriptionHandler) Validate(_ context.Context, req *UpsertSubscriptionRequest) *errorModels.ServerError {
	req.PlanSlug = strings.TrimSpace(req.PlanSlug)
	if req.PlanSlug == "" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "plan_slug is required",
		}
	}
	return nil
}

func (h *UpsertSubscriptionHandler) Serve(ctx context.Context, req *UpsertSubscriptionRequest) (UpsertSubscriptionResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpsertSubscriptionResponse{}, errorModels.PermissionDeniedServerError()
	}

	err := h.Service.UpsertSubscription(ctx, device.User.UserID, req.PlanSlug)
	if err != nil {
		if errors.Is(err, domain.ErrPlanNotFound) {
			return UpsertSubscriptionResponse{}, &errorModels.ServerError{
				HTTPCode: http.StatusNotFound,
				Code:     http.StatusNotFound,
				Message:  "Plan not found",
			}
		}
		return UpsertSubscriptionResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to update subscription",
		}
	}

	return UpsertSubscriptionResponse{Message: "Subscription updated"}, nil
}

// ─── Checkout ─────────────────────────────────────────────────────────────────

func CheckoutRoutes() []string { return []string{"/subscriptions/checkout"} }

type CheckoutHandler struct{ Service *application.Service }

type CheckoutRequest struct {
	PlanSlug string `json:"plan_slug"`
	Email    string `json:"email,omitempty"`
}

type CheckoutResponse struct {
	CheckoutURL string `json:"checkout_url"`
}

func (h *CheckoutHandler) Validate(_ context.Context, req *CheckoutRequest) *errorModels.ServerError {
	req.PlanSlug = strings.TrimSpace(req.PlanSlug)
	if req.PlanSlug == "" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "plan_slug is required",
		}
	}
	return nil
}

func (h *CheckoutHandler) Serve(ctx context.Context, req *CheckoutRequest) (CheckoutResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CheckoutResponse{}, errorModels.PermissionDeniedServerError()
	}

	checkoutURL, err := h.Service.CreateCheckout(ctx, device.User.UserID, req.PlanSlug, req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrPlanNotFound) {
			return CheckoutResponse{}, &errorModels.ServerError{
				HTTPCode: http.StatusNotFound,
				Code:     http.StatusNotFound,
				Message:  "Plan not found",
			}
		}
		logger.Ctx(ctx).Error("failed to create paddle checkout",
			logger.String("owner_id", strconv.FormatInt(device.User.UserID, 10)),
			logger.String("plan_slug", req.PlanSlug),
			logger.Error(err),
		)
		return CheckoutResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to create checkout session",
		}
	}

	return CheckoutResponse{CheckoutURL: checkoutURL}, nil
}

// ─── Paddle Webhook ───────────────────────────────────────────────────────────

// PaddleWebhookHTTPHandler returns a raw http.HandlerFunc for Paddle webhook events.
// It bypasses the typed handler infrastructure so that the raw request body can be read
// for HMAC-SHA256 signature verification before any JSON parsing occurs.
func PaddleWebhookHTTPHandler(service *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		r.Body = http.MaxBytesReader(w, r.Body, maxPaddleWebhookBodyBytes)
		rawBody, err := io.ReadAll(r.Body)
		if err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				http.Error(w, `{"error":"request body too large"}`, http.StatusRequestEntityTooLarge)
				return
			}
			http.Error(w, `{"error":"failed to read request body"}`, http.StatusBadRequest)
			return
		}

		// Verify webhook signature
		sig := r.Header.Get("Paddle-Signature")
		if verifyErr := paddlePkg.VerifyWebhookSignature(rawBody, sig); verifyErr != nil {
			http.Error(w, `{"error":"invalid webhook signature"}`, http.StatusUnauthorized)
			return
		}

		var event paddlePkg.WebhookEvent
		if parseErr := json.Unmarshal(rawBody, &event); parseErr != nil {
			http.Error(w, `{"error":"failed to parse webhook payload"}`, http.StatusBadRequest)
			return
		}

		// Only handle subscription events
		switch event.EventType {
		case "subscription.created", "subscription.updated", "subscription.activated",
			"subscription.cancelled", "subscription.paused":
		default:
			// Unhandled event type — acknowledge and ignore
			w.WriteHeader(http.StatusOK)
			return
		}

		var data paddlePkg.SubscriptionData
		if parseErr := json.Unmarshal(event.Data, &data); parseErr != nil {
			http.Error(w, `{"error":"failed to parse subscription data"}`, http.StatusBadRequest)
			return
		}

		if handleErr := service.HandlePaddleWebhook(ctx, event, data); handleErr != nil {
			logger.Ctx(ctx).Error("paddle webhook processing failed",
				logger.String("event_type", event.EventType),
				logger.String("subscription_id", data.ID),
				logger.Error(handleErr),
			)
			// Return 200 to prevent Paddle retries for business logic errors
		}

		w.WriteHeader(http.StatusOK)
	}
}
