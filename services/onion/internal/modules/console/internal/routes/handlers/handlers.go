package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/console/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func requireConsoleAdmin(ctx context.Context) *errorModels.ServerError {
	user, ok := authPkg.GetAuthUser(ctx)
	if !ok || user == nil {
		return &errorModels.ServerError{Code: http.StatusUnauthorized, Message: "Unauthorized"}
	}
	if user.Role != "console_admin" {
		return &errorModels.ServerError{Code: http.StatusForbidden, Message: "Console admin access required"}
	}
	return nil
}

// ── Stats ─────────────────────────────────────────────────────────────────────

type (
	StatsRequest  struct{}
	StatsResponse = application.StatsDTO
)

type StatsHandler struct{ Service *application.Service }

func (h *StatsHandler) Serve(ctx context.Context, _ *StatsRequest) (StatsResponse, *errorModels.ServerError) {
	if sErr := requireConsoleAdmin(ctx); sErr != nil {
		return StatsResponse{}, sErr
	}
	stats, err := h.Service.GetStats(ctx)
	if err != nil {
		return StatsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to fetch stats"}
	}
	return *stats, nil
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

func (h *ListUsersHandler) Serve(ctx context.Context, req *ListUsersRequest) (ListUsersResponse, *errorModels.ServerError) {
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
	result, err := h.Service.ListUsers(ctx, limit, req.Offset, req.ID, req.Email, name)
	if err != nil {
		return ListUsersResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list users"}
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

func (h *CreateUserHandler) Validate(_ context.Context, req *CreateUserRequest) *errorModels.ServerError {
	if req.Email == "" || req.Password == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Email and password are required"}
	}
	return nil
}

func (h *CreateUserHandler) Serve(ctx context.Context, req *CreateUserRequest) (CreateUserResponse, *errorModels.ServerError) {
	if sErr := requireConsoleAdmin(ctx); sErr != nil {
		return CreateUserResponse{}, sErr
	}
	user, err := h.Service.CreateUser(ctx, req.Email, req.Password, req.FullName, req.IsConsoleAdmin)
	if err != nil {
		return CreateUserResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: err.Error()}
	}
	return CreateUserResponse{User: *user, Message: "User created"}, nil
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

func (h *UpdateUserHandler) Serve(_ context.Context, _ *UpdateUserRequest) (UpdateUserResponse, *errorModels.ServerError) {
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

func (h *DeleteUserHandler) Serve(_ context.Context, _ *DeleteUserRequest) (DeleteUserResponse, *errorModels.ServerError) {
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

func (h *BanUserHandler) Serve(_ context.Context, _ *BanUserRequest) (BanUserResponse, *errorModels.ServerError) {
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

func (h *UnbanUserHandler) Serve(_ context.Context, _ *UnbanUserRequest) (UnbanUserResponse, *errorModels.ServerError) {
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

func (h *ListBusinessesHandler) Serve(ctx context.Context, req *ListBusinessesRequest) (ListBusinessesResponse, *errorModels.ServerError) {
	if sErr := requireConsoleAdmin(ctx); sErr != nil {
		return ListBusinessesResponse{}, sErr
	}
	limit := req.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	result, err := h.Service.ListBusinesses(ctx, limit, req.Offset, req.ID, req.Name, req.OwnerID)
	if err != nil {
		return ListBusinessesResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list businesses"}
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

func (h *UpdateBusinessHandler) Serve(_ context.Context, _ *UpdateBusinessRequest) (UpdateBusinessResponse, *errorModels.ServerError) {
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

func (h *ListSubscriptionsHandler) Serve(ctx context.Context, req *ListSubscriptionsRequest) (ListSubscriptionsResponse, *errorModels.ServerError) {
	if sErr := requireConsoleAdmin(ctx); sErr != nil {
		return ListSubscriptionsResponse{}, sErr
	}
	limit := req.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	result, err := h.Service.ListSubscriptions(ctx, limit, req.Offset, req.Status)
	if err != nil {
		return ListSubscriptionsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list subscriptions"}
	}
	return *result, nil
}

type UpdateSubscriptionRequest struct {
	PlanID    int64  `json:"plan_id"`
	Status    string `json:"status"`
	SeatCount int    `json:"seat_count"`
}

type UpdateSubscriptionResponse struct {
	Message string `json:"message"`
}

type UpdateSubscriptionHandler struct{ Service *application.Service }

func (h *UpdateSubscriptionHandler) Serve(_ context.Context, _ *UpdateSubscriptionRequest) (UpdateSubscriptionResponse, *errorModels.ServerError) {
	return UpdateSubscriptionResponse{}, nil // raw HTTP via route
}

// UpsertSubscriptionForUser

type UpsertSubscriptionRequest struct {
	OwnerID   int64  `json:"owner_id"`
	PlanID    int64  `json:"plan_id"`
	Status    string `json:"status"`
	SeatCount int    `json:"seat_count"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

type UpsertSubscriptionResponse struct {
	Message string `json:"message"`
}

type UpsertSubscriptionHandler struct{ Service *application.Service }

func (h *UpsertSubscriptionHandler) Validate(_ context.Context, req *UpsertSubscriptionRequest) *errorModels.ServerError {
	if req.OwnerID == 0 || req.PlanID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "owner_id and plan_id are required"}
	}
	if req.Status == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "status is required"}
	}
	return nil
}

func (h *UpsertSubscriptionHandler) Serve(ctx context.Context, req *UpsertSubscriptionRequest) (UpsertSubscriptionResponse, *errorModels.ServerError) {
	if sErr := requireConsoleAdmin(ctx); sErr != nil {
		return UpsertSubscriptionResponse{}, sErr
	}
	seatCount := req.SeatCount
	if seatCount <= 0 {
		seatCount = 1
	}
	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err == nil {
			expiresAt = &t
		}
	}
	if err := h.Service.UpsertSubscription(ctx, req.OwnerID, req.PlanID, req.Status, seatCount, expiresAt); err != nil {
		return UpsertSubscriptionResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to upsert subscription"}
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

func (h *ListPlansHandler) Serve(ctx context.Context, _ *ListPlansRequest) (ListPlansResponse, *errorModels.ServerError) {
	if sErr := requireConsoleAdmin(ctx); sErr != nil {
		return ListPlansResponse{}, sErr
	}
	plans, err := h.Service.ListPlans(ctx)
	if err != nil {
		return ListPlansResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list plans"}
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

func (h *UpdatePlanHandler) Serve(_ context.Context, _ *UpdatePlanRequest) (UpdatePlanResponse, *errorModels.ServerError) {
	return UpdatePlanResponse{}, nil // raw HTTP via route
}

// ── Grants ────────────────────────────────────────────────────────────────────

type ListGrantsRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type ListGrantsResponse = application.ListGrantsResult

type ListGrantsHandler struct{ Service *application.Service }

func (h *ListGrantsHandler) Serve(ctx context.Context, req *ListGrantsRequest) (ListGrantsResponse, *errorModels.ServerError) {
	if sErr := requireConsoleAdmin(ctx); sErr != nil {
		return ListGrantsResponse{}, sErr
	}
	limit := req.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	result, err := h.Service.ListGrants(ctx, limit, req.Offset)
	if err != nil {
		return ListGrantsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list grants"}
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

func (h *CreateGrantHandler) Validate(_ context.Context, req *CreateGrantRequest) *errorModels.ServerError {
	if req.UserID == 0 || req.PlanID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "user_id and plan_id are required"}
	}
	return nil
}

func (h *CreateGrantHandler) Serve(ctx context.Context, req *CreateGrantRequest) (CreateGrantResponse, *errorModels.ServerError) {
	if sErr := requireConsoleAdmin(ctx); sErr != nil {
		return CreateGrantResponse{}, sErr
	}
	admin, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateGrantResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, Message: "Unauthorized"}
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
		return CreateGrantResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create grant"}
	}
	return CreateGrantResponse{Grant: *grant, Message: "Access grant created"}, nil
}
