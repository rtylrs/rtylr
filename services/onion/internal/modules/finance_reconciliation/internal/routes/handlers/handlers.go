package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/finance_reconciliation/internal/application"
	"github.com/voxire/onion/internal/modules/finance_reconciliation/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// ---------------------------------------------------------------------------
// Overview
// ---------------------------------------------------------------------------

func OverviewRoutes() []string { return []string{"/finance/reconciliation"} }

type (
	OverviewHandler  struct{ Service *application.Service }
	OverviewRequest  struct{}
	OverviewResponse = domain.ReconciliationOverview
)

func (h *OverviewHandler) Serve(ctx context.Context, _ *OverviewRequest) (OverviewResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return OverviewResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return OverviewResponse{}, errorModels.PermissionDeniedServerError()
	}
	overview, err := h.Service.Overview(ctx, businessID)
	if err != nil {
		return OverviewResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to load reconciliation overview"}
	}
	return overview, nil
}

// ---------------------------------------------------------------------------
// Match
// ---------------------------------------------------------------------------

func MatchRoutes() []string { return []string{"/finance/reconciliation/match"} }

type MatchHandler struct{ Service *application.Service }

type MatchRequest struct {
	BankFeedID int64  `json:"bank_feed_id"`
	TargetType string `json:"target_type"`
	TargetID   int64  `json:"target_id"`
}

type MatchResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *MatchHandler) Validate(_ context.Context, req *MatchRequest) *errorModels.ServerError {
	if req.BankFeedID <= 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "bank_feed_id is required"}
	}
	if req.TargetType != domain.TargetTypeInvoice && req.TargetType != domain.TargetTypeExpense {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "target_type must be 'invoice' or 'expense'"}
	}
	if req.TargetID <= 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "target_id is required"}
	}
	return nil
}

func (h *MatchHandler) Serve(ctx context.Context, req *MatchRequest) (MatchResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return MatchResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return MatchResponse{}, errorModels.PermissionDeniedServerError()
	}
	id, err := h.Service.Match(ctx, businessID, domain.MatchInput{
		BankFeedID: req.BankFeedID,
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		MatchedBy:  device.User.UserID,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrBankFeedNotFound):
			return MatchResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Bank feed transaction not found"}
		case errors.Is(err, domain.ErrInvalidTarget):
			return MatchResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid match target"}
		default:
			return MatchResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create reconciliation match"}
		}
	}
	return MatchResponse{ID: id, Message: "Reconciliation match recorded"}, nil
}
