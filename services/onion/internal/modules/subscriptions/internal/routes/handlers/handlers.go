package handlers

import (
	"context"
	"errors"
	"net/http"

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
	ID       int64    `json:"id"`
	Name     string   `json:"name"`
	Slug     string   `json:"slug"`
	Apps     []string `json:"apps"`
	IsActive bool     `json:"is_active"`
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
			ID:       p.ID,
			Name:     p.Name,
			Slug:     p.Slug,
			Apps:     p.Apps,
			IsActive: p.IsActive,
		})
	}

	return ListPlansResponse{Plans: payloads}, nil
}

// ─── My Subscription ──────────────────────────────────────────────────────────

func MySubscriptionRoutes() []string { return []string{"/subscriptions/my"} }

type MySubscriptionHandler struct{ Service *application.Service }
type MySubscriptionRequest struct{}

type MySubscriptionResponse struct {
	Subscription *SubscriptionPayload `json:"subscription"`
}

type SubscriptionPayload struct {
	ID       int64    `json:"id"`
	PlanName string   `json:"plan_name"`
	PlanSlug string   `json:"plan_slug"`
	Apps     []string `json:"apps"`
	Status   string   `json:"status"`
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
			ID:       sub.ID,
			PlanName: sub.PlanName,
			PlanSlug: sub.PlanSlug,
			Apps:     sub.Apps,
			Status:   sub.Status,
		},
	}, nil
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
