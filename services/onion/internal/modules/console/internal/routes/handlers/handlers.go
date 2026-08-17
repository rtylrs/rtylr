package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/console/internal/application"
	"github.com/voxire/onion/internal/platform/aicache"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func requireConsoleAdmin(ctx context.Context) error {
	user, ok := authPkg.GetAuthUser(ctx)
	if !ok || user == nil {
		return onekiterrors.New(http.StatusUnauthorized, "Unauthorized")
	}
	if user.Role != "console_admin" {
		return onekiterrors.New(http.StatusForbidden, "Console admin access required")
	}
	return nil
}

func parseOptionalConsoleTime(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04", "2006-01-02"} {
		if t, err := time.Parse(layout, value); err == nil {
			return &t, nil
		}
	}
	return nil, errors.New("invalid time")
}

// ── Stats ─────────────────────────────────────────────────────────────────────

type (
	StatsRequest  struct{}
	StatsResponse = application.StatsDTO
)

type StatsHandler struct{ Service *application.Service }

func (h *StatsHandler) Serve(ctx context.Context, _ *StatsRequest) (StatsResponse, error) {
	if sErr := requireConsoleAdmin(ctx); sErr != nil {
		return StatsResponse{}, sErr
	}
	stats, err := h.Service.GetStats(ctx)
	if err != nil {
		return StatsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to fetch stats")
	}
	return *stats, nil
}

// ── AI cache analytics ───────────────────────────────────────────────────────

type (
	AICacheAnalyticsRequest  struct{}
	AICacheAnalyticsResponse = aicache.AnalyticsSnapshot
)

type AICacheAnalyticsHandler struct{}

func (h *AICacheAnalyticsHandler) Serve(ctx context.Context, _ *AICacheAnalyticsRequest) (AICacheAnalyticsResponse, error) {
	if sErr := requireConsoleAdmin(ctx); sErr != nil {
		return AICacheAnalyticsResponse{}, sErr
	}
	return aicache.Snapshot(), nil
}

// ── Users ─────────────────────────────────────────────────────────────────────

type ListUsersRequest struct {
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
}

type ListUsersResponse = application.ListUsersResult

type ListUsersHandler struct{ Service *application.Service }

func (h *ListUsersHandler) Serve(ctx context.Context, req *ListUsersRequest) (ListUsersResponse, error) {
	if sErr := requireConsoleAdmin(ctx); sErr != nil {
		return ListUsersResponse{}, sErr
	}
	limit := req.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	name := req.Name
	if name == "" {
		name = req.FullName
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}
	result, err := h.Service.ListUsers(ctx, limit, offset, req.ID, req.Email, name)
	if err != nil {
		return ListUsersResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list users")
	}
	return *result, nil
}

// CreateUser

type CreateUserRequest struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	FullName       string `json:"full_name"`
	IsConsoleAdmin bool   `json:"is_console_admin"`
}

type CreateUserResponse struct {
	User    application.UserDTO `json:"user"`
	Message string              `json:"message"`
}

type CreateUserHandler struct{ Service *application.Service }

func (h *CreateUserHandler) Validate(_ context.Context, req *CreateUserRequest) error {
	if req.Email == "" || req.Password == "" {
		return onekiterrors.New(http.StatusBadRequest, "Email and password are required")
	}
	return nil
}

func (h *CreateUserHandler) Serve(ctx context.Context, req *CreateUserRequest) (CreateUserResponse, error) {
	if sErr := requireConsoleAdmin(ctx); sErr != nil {
		return CreateUserResponse{}, sErr
	}
	user, err := h.Service.CreateUser(ctx, req.Email, req.Password, req.FullName, req.IsConsoleAdmin)
	if err != nil {
		return CreateUserResponse{}, onekiterrors.New(http.StatusBadRequest, err.Error())
	}
	return CreateUserResponse{User: *user, Message: "User created"}, nil
}

// Internal owner + business onboarding.

type CreateBusinessOwnerRequest struct {
	Email              string `json:"email"`
	Password           string `json:"password"`
	FullName           string `json:"full_name"`
	BusinessName       string `json:"business_name"`
	Industry           string `json:"industry"`
	Country            string `json:"country"`
	Timezone           string `json:"timezone"`
	Currency           string `json:"currency"`
	GrantPlan          bool   `json:"grant_plan"`
	PlanID             int64  `json:"plan_id"`
	SubscriptionStatus string `json:"subscription_status"`
	GrantDays          int    `json:"grant_days"`
}

type CreateBusinessOwnerResponse = application.CreateBusinessOwnerResult

type CreateBusinessOwnerHandler struct{ Service *application.Service }

func (h *CreateBusinessOwnerHandler) Validate(_ context.Context, req *CreateBusinessOwnerRequest) error {
	if req.Email == "" || req.Password == "" || req.BusinessName == "" {
		return onekiterrors.New(http.StatusBadRequest, "Email, password, and business name are required")
	}
	return nil
}

func (h *CreateBusinessOwnerHandler) Serve(ctx context.Context, req *CreateBusinessOwnerRequest) (CreateBusinessOwnerResponse, error) {
	if sErr := requireConsoleAdmin(ctx); sErr != nil {
		return CreateBusinessOwnerResponse{}, sErr
	}
	result, err := h.Service.CreateBusinessOwner(ctx, application.CreateBusinessOwnerInput{
		Email:              req.Email,
		Password:           req.Password,
		FullName:           req.FullName,
		BusinessName:       req.BusinessName,
		Industry:           req.Industry,
		Country:            req.Country,
		Timezone:           req.Timezone,
		Currency:           req.Currency,
		GrantPlan:          req.GrantPlan,
		PlanID:             req.PlanID,
		SubscriptionStatus: req.SubscriptionStatus,
		GrantDays:          req.GrantDays,
	})
	if err != nil {
		return CreateBusinessOwnerResponse{}, onekiterrors.New(http.StatusBadRequest, err.Error())
	}
	return *result, nil
}

// UpdateUser

type UpdateUserRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

type UpdateUserResponse struct {
	Message string `json:"message"`
}

type UpdateUserHandler struct{ Service *application.Service }

func (h *UpdateUserHandler) Serve(_ context.Context, _ *UpdateUserRequest) (UpdateUserResponse, error) {
	return UpdateUserResponse{}, nil // raw HTTP via route
}

// DeleteUser

type (
	DeleteUserRequest  struct{}
	DeleteUserResponse struct {
		Message string `json:"message"`
	}
)

type DeleteUserHandler struct{ Service *application.Service }

func (h *DeleteUserHandler) Serve(_ context.Context, _ *DeleteUserRequest) (DeleteUserResponse, error) {
	return DeleteUserResponse{}, nil // raw HTTP via route
}

// BanUser

type BanUserRequest struct {
	Reason string `json:"reason"`
}

type BanUserResponse struct {
	Message string `json:"message"`
}

type BanUserHandler struct{ Service *application.Service }

func (h *BanUserHandler) Serve(_ context.Context, _ *BanUserRequest) (BanUserResponse, error) {
	return BanUserResponse{}, nil // raw HTTP via route
}

// UnbanUser

type (
	UnbanUserRequest  struct{}
	UnbanUserResponse struct {
		Message string `json:"message"`
	}
)

type UnbanUserHandler struct{ Service *application.Service }

func (h *UnbanUserHandler) Serve(_ context.Context, _ *UnbanUserRequest) (UnbanUserResponse, error) {
	return UnbanUserResponse{}, nil // raw HTTP via route
}

// ── Businesses ────────────────────────────────────────────────────────────────

type ListBusinessesRequest struct {
	Limit   int    `json:"limit"`
	Offset  int    `json:"offset"`
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	OwnerID int64  `json:"owner_id"`
}

type ListBusinessesResponse = application.ListBusinessesResult

type ListBusinessesHandler struct{ Service *application.Service }

func (h *ListBusinessesHandler) Serve(ctx context.Context, req *ListBusinessesRequest) (ListBusinessesResponse, error) {
	if sErr := requireConsoleAdmin(ctx); sErr != nil {
		return ListBusinessesResponse{}, sErr
	}
	limit := req.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	result, err := h.Service.ListBusinesses(ctx, limit, req.Offset, req.ID, req.Name, req.OwnerID)
	if err != nil {
		return ListBusinessesResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list businesses")
	}
	return *result, nil
}

type UpdateBusinessRequest struct {
	Name string `json:"name"`
}

type UpdateBusinessResponse struct {
	Message string `json:"message"`
}

type UpdateBusinessHandler struct{ Service *application.Service }

func (h *UpdateBusinessHandler) Serve(_ context.Context, _ *UpdateBusinessRequest) (UpdateBusinessResponse, error) {
	return UpdateBusinessResponse{}, nil // raw HTTP via route
}

// ── Subscriptions ─────────────────────────────────────────────────────────────

type ListSubscriptionsRequest struct {
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Status string `json:"status"`
}

type ListSubscriptionsResponse = application.ListSubscriptionsResult

type ListSubscriptionsHandler struct{ Service *application.Service }

func (h *ListSubscriptionsHandler) Serve(ctx context.Context, req *ListSubscriptionsRequest) (ListSubscriptionsResponse, error) {
	if sErr := requireConsoleAdmin(ctx); sErr != nil {
		return ListSubscriptionsResponse{}, sErr
	}
	limit := req.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	result, err := h.Service.ListSubscriptions(ctx, limit, req.Offset, req.Status)
	if err != nil {
		return ListSubscriptionsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list subscriptions")
	}
	return *result, nil
}

type UpdateSubscriptionRequest struct {
	PlanID               int64  `json:"plan_id"`
	Status               string `json:"status"`
	SeatCount            int    `json:"seat_count"`
	TrialEndsAt          string `json:"trial_ends_at,omitempty"`
	NextBillingDate      string `json:"next_billing_date,omitempty"`
	CurrentPeriodEndsAt  string `json:"current_period_ends_at,omitempty"`
	TotalDays            *int   `json:"total_days,omitempty"`
	PaddleSubscriptionID string `json:"paddle_subscription_id,omitempty"`
	PaddleCustomerID     string `json:"paddle_customer_id,omitempty"`
}

type UpdateSubscriptionResponse struct {
	Message string `json:"message"`
}

type UpdateSubscriptionHandler struct{ Service *application.Service }

func (h *UpdateSubscriptionHandler) Serve(_ context.Context, _ *UpdateSubscriptionRequest) (UpdateSubscriptionResponse, error) {
	return UpdateSubscriptionResponse{}, nil // raw HTTP via route
}

// UpsertSubscriptionForUser

type UpsertSubscriptionRequest struct {
	OwnerID              int64  `json:"owner_id"`
	PlanID               int64  `json:"plan_id"`
	Status               string `json:"status"`
	SeatCount            int    `json:"seat_count"`
	ExpiresAt            string `json:"expires_at,omitempty"`
	TrialEndsAt          string `json:"trial_ends_at,omitempty"`
	NextBillingDate      string `json:"next_billing_date,omitempty"`
	CurrentPeriodEndsAt  string `json:"current_period_ends_at,omitempty"`
	TotalDays            *int   `json:"total_days,omitempty"`
	PaddleSubscriptionID string `json:"paddle_subscription_id,omitempty"`
	PaddleCustomerID     string `json:"paddle_customer_id,omitempty"`
}

type UpsertSubscriptionResponse struct {
	Message string `json:"message"`
}

type UpsertSubscriptionHandler struct{ Service *application.Service }

func (h *UpsertSubscriptionHandler) Validate(_ context.Context, req *UpsertSubscriptionRequest) error {
	if req.OwnerID == 0 || req.PlanID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "owner_id and plan_id are required")
	}
	if req.Status == "" {
		return onekiterrors.New(http.StatusBadRequest, "status is required")
	}
	return nil
}

func (h *UpsertSubscriptionHandler) Serve(ctx context.Context, req *UpsertSubscriptionRequest) (UpsertSubscriptionResponse, error) {
	if sErr := requireConsoleAdmin(ctx); sErr != nil {
		return UpsertSubscriptionResponse{}, sErr
	}
	trialEndsAt, err := parseOptionalConsoleTime(req.TrialEndsAt)
	if err != nil {
		return UpsertSubscriptionResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid trial_ends_at")
	}
	nextBillingDate, err := parseOptionalConsoleTime(req.NextBillingDate)
	if err != nil {
		return UpsertSubscriptionResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid next_billing_date")
	}
	currentPeriodEndsAt, err := parseOptionalConsoleTime(req.CurrentPeriodEndsAt)
	if err != nil {
		return UpsertSubscriptionResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid current_period_ends_at")
	}
	if currentPeriodEndsAt == nil && req.ExpiresAt != "" {
		currentPeriodEndsAt, err = parseOptionalConsoleTime(req.ExpiresAt)
		if err != nil {
			return UpsertSubscriptionResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid expires_at")
		}
	}

	if err := h.Service.UpsertSubscription(ctx, application.SubscriptionWriteInput{
		OwnerID:              req.OwnerID,
		PlanID:               req.PlanID,
		Status:               req.Status,
		SeatCount:            req.SeatCount,
		TrialEndsAt:          trialEndsAt,
		NextBillingDate:      nextBillingDate,
		CurrentPeriodEndsAt:  currentPeriodEndsAt,
		TotalDays:            req.TotalDays,
		PaddleSubscriptionID: req.PaddleSubscriptionID,
		PaddleCustomerID:     req.PaddleCustomerID,
	}); err != nil {
		if errors.Is(err, application.ErrNotOrganizationOwner) {
			return UpsertSubscriptionResponse{}, onekiterrors.New(http.StatusBadRequest, "Subscriptions can only be set for organization owners")
		}
		if errors.Is(err, application.ErrInvalidSubscription) {
			return UpsertSubscriptionResponse{}, onekiterrors.New(http.StatusBadRequest, err.Error())
		}
		return UpsertSubscriptionResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to upsert subscription")
	}
	return UpsertSubscriptionResponse{Message: "Subscription updated"}, nil
}

// DeleteSubscription

type (
	DeleteSubscriptionRequest  struct{}
	DeleteSubscriptionResponse struct {
		Message string `json:"message"`
	}
)

// ── Plans ─────────────────────────────────────────────────────────────────────

type (
	ListPlansRequest  struct{}
	ListPlansResponse struct {
		Plans []application.PlanDTO `json:"plans"`
	}
)

type ListPlansHandler struct{ Service *application.Service }

func (h *ListPlansHandler) Serve(ctx context.Context, _ *ListPlansRequest) (ListPlansResponse, error) {
	if sErr := requireConsoleAdmin(ctx); sErr != nil {
		return ListPlansResponse{}, sErr
	}
	plans, err := h.Service.ListPlans(ctx)
	if err != nil {
		return ListPlansResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list plans")
	}
	return ListPlansResponse{Plans: plans}, nil
}

type UpdatePlanRequest struct {
	Name                    string `json:"name"`
	PricePerSeatCents       int    `json:"price_per_seat_cents"`
	AnnualPricePerSeatCents int    `json:"annual_price_per_seat_cents"`
	IsActive                bool   `json:"is_active"`
}

type UpdatePlanResponse struct {
	Message string `json:"message"`
}

type UpdatePlanHandler struct{ Service *application.Service }

func (h *UpdatePlanHandler) Serve(_ context.Context, _ *UpdatePlanRequest) (UpdatePlanResponse, error) {
	return UpdatePlanResponse{}, nil // raw HTTP via route
}

// ── Grants ────────────────────────────────────────────────────────────────────

type ListGrantsRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type ListGrantsResponse = application.ListGrantsResult

type ListGrantsHandler struct{ Service *application.Service }

func (h *ListGrantsHandler) Serve(ctx context.Context, req *ListGrantsRequest) (ListGrantsResponse, error) {
	if sErr := requireConsoleAdmin(ctx); sErr != nil {
		return ListGrantsResponse{}, sErr
	}
	limit := req.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	result, err := h.Service.ListGrants(ctx, limit, req.Offset)
	if err != nil {
		return ListGrantsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list grants")
	}
	return *result, nil
}

type CreateGrantRequest struct {
	UserID    int64  `json:"user_id"`
	PlanID    int64  `json:"plan_id"`
	Note      string `json:"note"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

type CreateGrantResponse struct {
	Grant   application.GrantDTO `json:"grant"`
	Message string               `json:"message"`
}

type CreateGrantHandler struct{ Service *application.Service }

func (h *CreateGrantHandler) Validate(_ context.Context, req *CreateGrantRequest) error {
	if req.UserID == 0 || req.PlanID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "user_id and plan_id are required")
	}
	return nil
}

func (h *CreateGrantHandler) Serve(ctx context.Context, req *CreateGrantRequest) (CreateGrantResponse, error) {
	if sErr := requireConsoleAdmin(ctx); sErr != nil {
		return CreateGrantResponse{}, sErr
	}
	admin, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateGrantResponse{}, onekiterrors.New(http.StatusUnauthorized, "Unauthorized")
	}
	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err == nil {
			expiresAt = &t
		}
	}
	grant, err := h.Service.CreateGrant(ctx, req.UserID, req.PlanID, admin.UserID, req.Note, expiresAt)
	if err != nil {
		if errors.Is(err, application.ErrNotOrganizationOwner) {
			return CreateGrantResponse{}, onekiterrors.New(http.StatusBadRequest, "Access grants can only be created for organization owners")
		}
		return CreateGrantResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create grant")
	}
	return CreateGrantResponse{Grant: *grant, Message: "Access grant created"}, nil
}

// ── Addon Grants ──────────────────────────────────────────────────────────────

type ListAddonsRequest struct{}

type ListAddonsResponse = application.ListAddonsResult

type ListAddonsHandler struct{ Service *application.Service }

func (h *ListAddonsHandler) Serve(ctx context.Context, _ *ListAddonsRequest) (ListAddonsResponse, error) {
	if sErr := requireConsoleAdmin(ctx); sErr != nil {
		return ListAddonsResponse{}, sErr
	}
	result, err := h.Service.ListAddons(ctx)
	if err != nil {
		return ListAddonsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list addons")
	}
	return *result, nil
}

type ListAddonGrantsRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type ListAddonGrantsResponse = application.ListAddonGrantsResult

type ListAddonGrantsHandler struct{ Service *application.Service }

func (h *ListAddonGrantsHandler) Serve(ctx context.Context, req *ListAddonGrantsRequest) (ListAddonGrantsResponse, error) {
	if sErr := requireConsoleAdmin(ctx); sErr != nil {
		return ListAddonGrantsResponse{}, sErr
	}
	limit := req.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	result, err := h.Service.ListAddonGrants(ctx, limit, req.Offset)
	if err != nil {
		return ListAddonGrantsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list addon grants")
	}
	return *result, nil
}

type CreateAddonGrantRequest struct {
	UserID     int64  `json:"user_id"`
	BusinessID int64  `json:"business_id,omitempty"`
	AddonID    string `json:"addon_id"`
	Note       string `json:"note"`
	ExpiresAt  string `json:"expires_at,omitempty"`
}

type CreateAddonGrantResponse struct {
	Grant   application.AddonGrantDTO `json:"grant"`
	Message string                    `json:"message"`
}

type CreateAddonGrantHandler struct{ Service *application.Service }

func (h *CreateAddonGrantHandler) Validate(_ context.Context, req *CreateAddonGrantRequest) error {
	if req.UserID == 0 || req.AddonID == "" {
		return onekiterrors.New(http.StatusBadRequest, "user_id and addon_id are required")
	}
	return nil
}

func (h *CreateAddonGrantHandler) Serve(ctx context.Context, req *CreateAddonGrantRequest) (CreateAddonGrantResponse, error) {
	if sErr := requireConsoleAdmin(ctx); sErr != nil {
		return CreateAddonGrantResponse{}, sErr
	}
	admin, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateAddonGrantResponse{}, onekiterrors.New(http.StatusUnauthorized, "Unauthorized")
	}
	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err == nil {
			expiresAt = &t
		}
	}
	grant, err := h.Service.CreateAddonGrant(ctx, req.UserID, req.BusinessID, req.AddonID, admin.UserID, req.Note, expiresAt)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrNotOrganizationOwner):
			return CreateAddonGrantResponse{}, onekiterrors.New(http.StatusBadRequest, "Addon grants can only be created for organization owners")
		case errors.Is(err, application.ErrBusinessNotOwnedByUser):
			return CreateAddonGrantResponse{}, onekiterrors.New(http.StatusBadRequest, "Business does not belong to this owner")
		case errors.Is(err, application.ErrAddonNotFound):
			return CreateAddonGrantResponse{}, onekiterrors.New(http.StatusBadRequest, "Addon not found")
		default:
			return CreateAddonGrantResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create addon grant")
		}
	}
	return CreateAddonGrantResponse{Grant: *grant, Message: "Addon grant created"}, nil
}
