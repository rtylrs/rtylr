package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/loyalty/internal/application"
	"github.com/voxire/onion/internal/modules/loyalty/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// ─── List Members ─────────────────────────────────────────────────────────────

func LoyaltyMembersRoutes() []string { return []string{"/crm/loyalty/members"} }

type ListMembersHandler struct{ Service *application.Service }

type ListMembersRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type MemberPayload struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Initials     string `json:"initials"`
	Points       int    `json:"points"`
	Tier         string `json:"tier"`
	Visits       int    `json:"visits"`
	JoinedAt     string `json:"joined_at"`
	LastActiveAt string `json:"last_active"`
}

type ListMembersResponse struct {
	Members []MemberPayload `json:"members"`
}

func (h *ListMembersHandler) Serve(ctx context.Context, req *ListMembersRequest) (ListMembersResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListMembersResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListMembersResponse{}, onekiterrors.PermissionDenied()
	}
	members, err := h.Service.ListMembers(ctx, businessID, req.Limit, req.Offset)
	if err != nil {
		return ListMembersResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list loyalty members")
	}
	payloads := make([]MemberPayload, 0, len(members))
	for _, m := range members {
		payloads = append(payloads, MemberPayload{
			ID:           m.ID,
			Name:         m.Name,
			Initials:     m.Initials(),
			Points:       m.Points,
			Tier:         m.Tier,
			Visits:       m.Visits,
			JoinedAt:     m.JoinedAt.Format(time.RFC3339),
			LastActiveAt: m.LastActiveAt.Format(time.RFC3339),
		})
	}
	return ListMembersResponse{Members: payloads}, nil
}

// ─── List Rewards ─────────────────────────────────────────────────────────────

func LoyaltyRewardsListRoutes() []string { return []string{"/crm/loyalty/rewards"} }

type ListRewardsHandler struct{ Service *application.Service }

type ListRewardsRequest struct {
	ActiveOnly bool `json:"active_only"`
}

type RewardPayload struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	PointsCost  int    `json:"points_cost"`
	Stock       *int   `json:"stock"`
	Icon        string `json:"icon"`
	Active      bool   `json:"active"`
}

type ListRewardsResponse struct {
	Rewards []RewardPayload `json:"rewards"`
}

func (h *ListRewardsHandler) Serve(ctx context.Context, req *ListRewardsRequest) (ListRewardsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListRewardsResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListRewardsResponse{}, onekiterrors.PermissionDenied()
	}
	rewards, err := h.Service.ListRewards(ctx, businessID, req.ActiveOnly)
	if err != nil {
		return ListRewardsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list loyalty rewards")
	}
	payloads := make([]RewardPayload, 0, len(rewards))
	for _, rw := range rewards {
		payloads = append(payloads, RewardPayload{
			ID:          rw.ID,
			Name:        rw.Name,
			Description: rw.Description,
			Category:    rw.Category,
			PointsCost:  rw.PointsCost,
			Stock:       rw.Stock,
			Icon:        rw.Icon,
			Active:      rw.Active,
		})
	}
	return ListRewardsResponse{Rewards: payloads}, nil
}

// ─── Upsert Reward ────────────────────────────────────────────────────────────

func LoyaltyRewardsUpsertRoutes() []string { return []string{"/crm/loyalty/rewards/upsert"} }

type UpsertRewardHandler struct{ Service *application.Service }

type UpsertRewardRequest struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	PointsCost  int    `json:"points_cost"`
	Stock       *int   `json:"stock"`
	Icon        string `json:"icon"`
	Active      *bool  `json:"active"`
}

type UpsertRewardResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *UpsertRewardHandler) Validate(_ context.Context, req *UpsertRewardRequest) error {
	if req.Name == "" {
		return onekiterrors.New(http.StatusBadRequest, "Reward name is required")
	}
	if req.PointsCost < 0 {
		return onekiterrors.New(http.StatusBadRequest, "Points cost cannot be negative")
	}
	if req.Stock != nil && *req.Stock < 0 {
		return onekiterrors.New(http.StatusBadRequest, "Stock cannot be negative")
	}
	return nil
}

func (h *UpsertRewardHandler) Serve(ctx context.Context, req *UpsertRewardRequest) (UpsertRewardResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpsertRewardResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireLoyaltyManager(ctx); err != nil {
		return UpsertRewardResponse{}, err
	}
	active := true
	if req.Active != nil {
		active = *req.Active
	}
	if req.Category == "" {
		req.Category = "general"
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpsertRewardResponse{}, onekiterrors.PermissionDenied()
	}
	id, err := h.Service.UpsertReward(ctx, domain.LoyaltyReward{
		ID:          req.ID,
		OwnerID:     device.User.UserID,
		BusinessID:  businessID,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		PointsCost:  req.PointsCost,
		Stock:       req.Stock,
		Icon:        req.Icon,
		Active:      active,
	})
	if err != nil {
		if errors.Is(err, domain.ErrRewardNotFound) {
			return UpsertRewardResponse{}, onekiterrors.New(http.StatusNotFound, "Reward not found")
		}
		return UpsertRewardResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to save reward")
	}
	msg := "Reward created"
	if req.ID != 0 {
		msg = "Reward updated"
	}
	return UpsertRewardResponse{ID: id, Message: msg}, nil
}
