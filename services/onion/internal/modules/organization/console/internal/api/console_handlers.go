package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/organization/console/internal/application"
	"github.com/voxire/onion/internal/platform/aicache"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

// ── AI cache analytics ───────────────────────────────────────────────────────

type (
	AICacheAnalyticsRequest  struct{}
	AICacheAnalyticsResponse = aicache.AnalyticsSnapshot
)

type AICacheAnalyticsHandler struct{}

func (h *AICacheAnalyticsHandler) Serve(ctx context.Context, _ *AICacheAnalyticsRequest) (AICacheAnalyticsResponse, error) {
	if sErr := requireConsoleResourceAdmin(ctx); sErr != nil {
		return AICacheAnalyticsResponse{}, sErr
	}
	return aicache.Snapshot(), nil
}

// ── Grants ────────────────────────────────────────────────────────────────────

type ListGrantsRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type ListGrantsResponse = application.ListGrantsResult

type ListGrantsHandler struct{ Service *application.Service }

func (h *ListGrantsHandler) Serve(ctx context.Context, req *ListGrantsRequest) (ListGrantsResponse, error) {
	if sErr := requireConsoleResourceAdmin(ctx); sErr != nil {
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

// CreateGrantRequest names the business directly or through its owner; see
// application.Service.CreateGrant.
type CreateGrantRequest struct {
	BusinessID int64  `json:"business_id"`
	UserID     int64  `json:"user_id"`
	PlanID     int64  `json:"plan_id"`
	Note       string `json:"note"`
	ExpiresAt  string `json:"expires_at,omitempty"`
}

type CreateGrantResponse struct {
	Grant   application.GrantDTO `json:"grant"`
	Message string               `json:"message"`
}

type CreateGrantHandler struct{ Service *application.Service }

func parseGrantExpiresAt(value string) (*time.Time, error) {
	return parseOptionalTime(&value)
}

func (h *CreateGrantHandler) Validate(_ context.Context, req *CreateGrantRequest) error {
	if (req.BusinessID == 0 && req.UserID == 0) || req.PlanID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "business_id (or user_id) and plan_id are required")
	}
	if _, err := parseGrantExpiresAt(req.ExpiresAt); err != nil {
		return badRequest("Invalid expires_at")
	}
	return nil
}

func (h *CreateGrantHandler) Serve(ctx context.Context, req *CreateGrantRequest) (CreateGrantResponse, error) {
	if sErr := requireConsoleResourceAdmin(ctx); sErr != nil {
		return CreateGrantResponse{}, sErr
	}
	admin, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateGrantResponse{}, onekiterrors.New(http.StatusUnauthorized, "Unauthorized")
	}
	expiresAt, err := parseGrantExpiresAt(req.ExpiresAt)
	if err != nil {
		return CreateGrantResponse{}, badRequest("Invalid expires_at")
	}
	grant, err := h.Service.CreateGrant(ctx, req.UserID, req.BusinessID, req.PlanID, admin.UserID, req.Note, expiresAt)
	if err != nil {
		if errors.Is(err, application.ErrNotOrganizationOwner) {
			return CreateGrantResponse{}, onekiterrors.New(http.StatusBadRequest, "Access grants can only be created for organization owners")
		}
		if mapped := billingBusinessError(err); mapped != nil {
			return CreateGrantResponse{}, mapped
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
	if sErr := requireConsoleResourceAdmin(ctx); sErr != nil {
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
	if sErr := requireConsoleResourceAdmin(ctx); sErr != nil {
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
	if (req.BusinessID == 0 && req.UserID == 0) || req.AddonID == "" {
		return onekiterrors.New(http.StatusBadRequest, "business_id (or user_id) and addon_id are required")
	}
	if _, err := parseGrantExpiresAt(req.ExpiresAt); err != nil {
		return badRequest("Invalid expires_at")
	}
	return nil
}

func (h *CreateAddonGrantHandler) Serve(ctx context.Context, req *CreateAddonGrantRequest) (CreateAddonGrantResponse, error) {
	if sErr := requireConsoleResourceAdmin(ctx); sErr != nil {
		return CreateAddonGrantResponse{}, sErr
	}
	admin, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateAddonGrantResponse{}, onekiterrors.New(http.StatusUnauthorized, "Unauthorized")
	}
	expiresAt, err := parseGrantExpiresAt(req.ExpiresAt)
	if err != nil {
		return CreateAddonGrantResponse{}, badRequest("Invalid expires_at")
	}
	grant, err := h.Service.CreateAddonGrant(ctx, req.UserID, req.BusinessID, req.AddonID, admin.UserID, req.Note, expiresAt)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrNotOrganizationOwner):
			return CreateAddonGrantResponse{}, onekiterrors.New(http.StatusBadRequest, "Addon grants can only be created for organization owners")
		case errors.Is(err, application.ErrAddonNotFound):
			return CreateAddonGrantResponse{}, onekiterrors.New(http.StatusBadRequest, "Addon not found")
		}
		if mapped := billingBusinessError(err); mapped != nil {
			return CreateAddonGrantResponse{}, mapped
		}
		return CreateAddonGrantResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create addon grant")
	}
	return CreateAddonGrantResponse{Grant: *grant, Message: "Addon grant created"}, nil
}

// billingBusinessError maps the errors that pick the business of a console
// billing write, and returns nil for any other error.
func billingBusinessError(err error) error {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return onekiterrors.New(http.StatusNotFound, "Business not found")
	case errors.Is(err, application.ErrBusinessRequired):
		return badRequest("business_id is required")
	case errors.Is(err, application.ErrOwnerHasSeveralBusinesses):
		return badRequest("This owner has several businesses; pass business_id")
	case errors.Is(err, application.ErrNotOrganizationOwner):
		return badRequest("This user owns no active business")
	case errors.Is(err, application.ErrBusinessNotOwnedByUser):
		return badRequest("Business does not belong to this owner")
	default:
		return nil
	}
}
