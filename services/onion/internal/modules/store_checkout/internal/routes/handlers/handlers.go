package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/store_checkout/internal/application"
	"github.com/voxire/onion/internal/modules/store_checkout/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func CreateRoutes() []string   { return []string{"/store/checkout/create"} }
func CompleteRoutes() []string { return []string{"/store/checkout/complete"} }

type (
	CreateHandler struct{ Service *application.Service }
	CreateRequest struct {
		SessionID     string  `json:"session_id"`
		CustomerEmail string  `json:"customer_email"`
		Items         string  `json:"items"`
		Subtotal      float64 `json:"subtotal"`
		Tax           float64 `json:"tax"`
		Total         float64 `json:"total"`
	}
)

type CreateResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) error {
	if req.SessionID == "" {
		return onekiterrors.New(http.StatusBadRequest, "session_id is required")
	}
	if strings.TrimSpace(req.Items) == "" && (req.Subtotal != 0 || req.Tax != 0 || req.Total != 0) {
		return onekiterrors.New(http.StatusBadRequest, "catalog items are required for a non-zero checkout")
	}
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateResponse{}, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return CreateResponse{}, onekiterrors.PermissionDenied()
	}

	id, err := h.Service.Create(ctx, businessID, authUser.UserID, domain.CheckoutCreateInput{
		SessionID: req.SessionID, CustomerEmail: req.CustomerEmail, Items: req.Items,
		Subtotal: req.Subtotal, Tax: req.Tax, Total: req.Total,
	})
	if err != nil {
		return CreateResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create checkout")
	}
	return CreateResponse{ID: id, Message: "Checkout created"}, nil
}

type (
	CompleteHandler struct{ Service *application.Service }
	CompleteRequest struct {
		ID int64 `json:"id"`
	}
)

type CompleteResponse struct {
	Message string `json:"message"`
}

func (h *CompleteHandler) Validate(_ context.Context, req *CompleteRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "ID is required")
	}
	return nil
}

func (h *CompleteHandler) Serve(ctx context.Context, req *CompleteRequest) (CompleteResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CompleteResponse{}, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return CompleteResponse{}, onekiterrors.PermissionDenied()
	}

	if err := h.Service.Complete(ctx, businessID, req.ID, authUser.UserID); err != nil {
		return CompleteResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to complete checkout")
	}
	return CompleteResponse{Message: "Checkout completed"}, nil
}
