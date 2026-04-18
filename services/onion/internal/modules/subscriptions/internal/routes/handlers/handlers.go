package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/subscriptions/internal/application"
	"github.com/voxire/onion/internal/modules/subscriptions/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// ─── List Plans ───────────────────────────────────────────────────────────────

func ListPlansRoutes() []string { return []string{"/subscriptions/plans"} }

type ListPlansHandler struct{ Service *application.Service }
type ListPlansRequest struct{}

type PlanPayload struct {
	ID                int64    `json:"id"`
	Name              string   `json:"name"`
	Slug              string   `json:"slug"`
	Apps              []string `json:"apps"`
	PricePerSeatCents int      `json:"price_per_seat_cents"`
	TrialDays         int      `json:"trial_days"`
	IsActive          bool     `json:"is_active"`
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
			ID:                p.ID,
			Name:              p.Name,
			Slug:              p.Slug,
			Apps:              p.Apps,
			PricePerSeatCents: p.PricePerSeatCents,
			TrialDays:         p.TrialDays,
			IsActive:          p.IsActive,
		})
	}

	return ListPlansResponse{Plans: payloads}, nil
}

// ─── My Subscription ──────────────────────────────────────────────────────────

func MySubscriptionRoutes() []string { return []string{"/subscriptions/my"} }

type MySubscriptionHandler struct{ Service *application.Service }
type MySubscriptionRequest struct{}

type SubscriptionPayload struct {
	ID                   int64      `json:"id"`
	PlanName             string     `json:"plan_name"`
	PlanSlug             string     `json:"plan_slug"`
	Apps                 []string   `json:"apps"`
	PricePerSeatCents    int        `json:"price_per_seat_cents"`
	Status               string     `json:"status"`
	SeatCount            int        `json:"seat_count"`
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
			TrialEndsAt:          sub.TrialEndsAt,
			NextBillingDate:      sub.NextBillingDate,
			PaddleSubscriptionID: sub.PaddleSubscriptionID,
		},
	}, nil
}

// ─── Billing Detail (current month) ──────────────────────────────────────────

func BillingDetailRoutes() []string { return []string{"/subscriptions/billing"} }

type BillingDetailHandler struct{ Service *application.Service }
type BillingDetailRequest struct{}

type BillingLineItemPayload struct {
	PeriodStart       time.Time  `json:"period_start"`
	PeriodEnd         *time.Time `json:"period_end"`
	SeatCount         int        `json:"seat_count"`
	PricePerSeatCents int        `json:"price_per_seat_cents"`
	AmountCents       int        `json:"amount_cents"`
}

type BillingDetailPayload struct {
	BillingMonth         time.Time                `json:"billing_month"`
	NextBillingDate      *time.Time               `json:"next_billing_date"`
	LineItems            []BillingLineItemPayload  `json:"line_items"`
	TotalCents           int                      `json:"total_cents"`
	PlanName             string                   `json:"plan_name"`
	PricePerSeatCents    int                      `json:"price_per_seat_cents"`
	CurrentSeatCount     int                      `json:"current_seat_count"`
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

	detail, err := h.Service.GetBillingDetail(ctx, device.User.UserID)
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
			PricePerSeatCents: li.PricePerSeatCents,
			AmountCents:       li.AmountCents,
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
			Status:               detail.Status,
			TrialEndsAt:          detail.TrialEndsAt,
			PaddleSubscriptionID: detail.PaddleSubscriptionID,
		},
	}, nil
}

// ─── Billing History ──────────────────────────────────────────────────────────

func BillingHistoryRoutes() []string { return []string{"/subscriptions/billing/history"} }

type BillingHistoryHandler struct{ Service *application.Service }
type BillingHistoryRequest struct{}

type BillingMonthPayload struct {
	Month      time.Time                `json:"month"`
	TotalCents int                      `json:"total_cents"`
	LineItems  []BillingLineItemPayload  `json:"line_items"`
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

	months, err := h.Service.GetBillingHistory(ctx, device.User.UserID, 12)
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
				PricePerSeatCents: li.PricePerSeatCents,
				AmountCents:       li.AmountCents,
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
