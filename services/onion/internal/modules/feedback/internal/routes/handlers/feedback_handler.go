package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/feedback/internal/application"
	"github.com/voxire/onion/internal/modules/feedback/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func FeedbackListRoutes() []string {
	return []string{"/crm/feedback"}
}

func FeedbackCreateRoutes() []string {
	return []string{"/crm/feedback"}
}

type FeedbackPayload struct {
	ID         int64  `json:"id"`
	OwnerID    int64  `json:"owner_id"`
	CustomerID int64  `json:"customer_id"`
	Type       string `json:"type"`
	Score      int    `json:"score"`
	Comment    string `json:"comment"`
	Source     string `json:"source"`
	CreatedAt  string `json:"created_at"`
}

// -- LIST --------------------------------------------------------------------

type ListFeedbackHandler struct {
	Service *application.Service
}

type ListFeedbackRequest struct {
	Type     string `json:"type"`
	Source   string `json:"source"`
	MinScore *int   `json:"min_score"`
	MaxScore *int   `json:"max_score"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type ListFeedbackResponse struct {
	Feedback []FeedbackPayload `json:"feedback"`
	Total    int               `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

func (h *ListFeedbackHandler) Serve(ctx context.Context, req *ListFeedbackRequest) (ListFeedbackResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListFeedbackResponse{}, errorModels.PermissionDeniedServerError()
	}

	if req.Page < 1 {
		req.Page = 1
	}

	if req.PageSize < 1 {
		req.PageSize = 20
	}

	filters := domain.ListFilters{
		Type:   req.Type,
		Source: req.Source,
	}

	if req.MinScore != nil {
		filters.MinScore = *req.MinScore
		filters.MinScoreSet = true
	}

	if req.MaxScore != nil {
		filters.MaxScore = *req.MaxScore
		filters.MaxScoreSet = true
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListFeedbackResponse{}, errorModels.PermissionDeniedServerError()
	}
	items, total, err := h.Service.List(ctx, businessID, filters, req.Page, req.PageSize)
	if err != nil {
		return ListFeedbackResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list feedback",
		}
	}

	payloads := make([]FeedbackPayload, len(items))
	for i, f := range items {
		payloads[i] = FeedbackPayload{
			ID:         f.ID,
			OwnerID:    f.OwnerID,
			CustomerID: f.CustomerID,
			Type:       f.Type,
			Score:      f.Score,
			Comment:    f.Comment,
			Source:     f.Source,
			CreatedAt:  f.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return ListFeedbackResponse{
		Feedback: payloads,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// -- CREATE ------------------------------------------------------------------

type CreateFeedbackHandler struct {
	Service *application.Service
}

type CreateFeedbackRequest struct {
	CustomerID int64  `json:"customer_id"`
	Type       string `json:"type"`
	Score      int    `json:"score"`
	Comment    string `json:"comment"`
	Source     string `json:"source"`
}

type CreateFeedbackResponse struct {
	FeedbackID int64  `json:"feedback_id"`
	Message    string `json:"message"`
}

func (h *CreateFeedbackHandler) Validate(_ context.Context, req *CreateFeedbackRequest) *errorModels.ServerError {
	if req.CustomerID == 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Customer ID is required",
		}
	}

	if req.Type == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Feedback type is required",
		}
	}

	validTypes := map[string]bool{"nps": true, "review": true}
	if !validTypes[req.Type] {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Feedback type must be 'nps' or 'review'",
		}
	}

	return nil
}

func (h *CreateFeedbackHandler) Serve(ctx context.Context, req *CreateFeedbackRequest) (CreateFeedbackResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateFeedbackResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateFeedbackResponse{}, errorModels.PermissionDeniedServerError()
	}
	id, err := h.Service.Create(ctx, domain.FeedbackInput{
		OwnerID:    device.User.UserID,
		BusinessID: businessID,
		CustomerID: req.CustomerID,
		Type:       req.Type,
		Score:      req.Score,
		Comment:    req.Comment,
		Source:     req.Source,
	})
	if err != nil {
		return CreateFeedbackResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create feedback",
		}
	}

	return CreateFeedbackResponse{FeedbackID: id, Message: "Feedback created successfully"}, nil
}
