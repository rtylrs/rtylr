package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/reviews/internal/application"
	"github.com/voxire/onion/internal/modules/reviews/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

type ReviewPayload struct {
	ID           int64     `json:"id"`
	OwnerID      int64     `json:"owner_id"`
	StaffID      int64     `json:"staff_id"`
	ReviewerID   int64     `json:"reviewer_id"`
	Period       string    `json:"period"`
	Rating       float64   `json:"rating"`
	GoalsRating  float64   `json:"goals_rating"`
	ValuesRating float64   `json:"values_rating"`
	SkillsRating float64   `json:"skills_rating"`
	Notes        string    `json:"notes"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func newPayload(r domain.Review) ReviewPayload {
	return ReviewPayload{
		ID: r.ID, OwnerID: r.OwnerID, StaffID: r.StaffID, ReviewerID: r.ReviewerID,
		Period: r.Period, Rating: r.Rating, GoalsRating: r.GoalsRating,
		ValuesRating: r.ValuesRating, SkillsRating: r.SkillsRating,
		Notes: r.Notes, Status: r.Status, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func ListRoutes() []string { return []string{"/hr/reviews"} }

type (
	ListHandler  struct{ Service *application.Service }
	ListRequest  struct{}
	ListResponse struct {
		Reviews []ReviewPayload `json:"reviews"`
	}
)

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}
	items, err := h.Service.List(ctx, businessID)
	if err != nil {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list reviews"}
	}
	payloads := make([]ReviewPayload, 0, len(items))
	for _, item := range items {
		payloads = append(payloads, newPayload(item))
	}
	return ListResponse{Reviews: payloads}, nil
}

func CreateRoutes() []string { return []string{"/hr/reviews"} }

type (
	CreateHandler struct{ Service *application.Service }
	CreateRequest struct {
		StaffID      int64   `json:"staff_id"`
		ReviewerID   int64   `json:"reviewer_id"`
		Period       string  `json:"period"`
		Rating       float64 `json:"rating"`
		GoalsRating  float64 `json:"goals_rating"`
		ValuesRating float64 `json:"values_rating"`
		SkillsRating float64 `json:"skills_rating"`
		Notes        string  `json:"notes"`
		Status       string  `json:"status"`
	}
)

type CreateResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
	if req.StaffID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "staffId is required"}
	}
	if req.Period == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "period is required"}
	}
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateResponse{}, errorModels.PermissionDeniedServerError()
	}
	input := domain.ReviewInput{
		StaffID: req.StaffID, ReviewerID: req.ReviewerID, Period: req.Period,
		Rating: req.Rating, GoalsRating: req.GoalsRating, ValuesRating: req.ValuesRating,
		SkillsRating: req.SkillsRating, Notes: req.Notes, Status: req.Status,
	}
	id, err := h.Service.Create(ctx, businessID, input)
	if err != nil {
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create review"}
	}
	return CreateResponse{ID: id, Message: "Review created successfully"}, nil
}
