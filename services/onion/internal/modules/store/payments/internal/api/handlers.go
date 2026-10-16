package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/store/payments/internal/application"
	"github.com/voxire/onion/internal/modules/store/payments/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func ListRoutes() []string   { return []string{"/store/payments"} }
func CreateRoutes() []string { return []string{"/store/payments"} }
func UpdateRoutes() []string { return []string{"/store/payments"} }

type (
	ListHandler  struct{ Service *application.Service }
	ListRequest  struct{}
	ListResponse struct {
		Methods []paymentMethodPayload `json:"methods"`
	}
)

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListResponse{}, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return ListResponse{}, onekiterrors.PermissionDenied()
	}
	items, err := h.Service.List(ctx, businessID)
	if err != nil {
		return ListResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list payment methods")
	}
	return ListResponse{Methods: payloadPaymentMethodList(items)}, nil
}

type (
	CreateHandler struct{ Service *application.Service }
	CreateRequest struct {
		Provider string `json:"provider"`
		Config   string `json:"config"`
		IsActive bool   `json:"is_active"`
	}
)

type CreateResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) error {
	if strings.TrimSpace(req.Provider) == "" {
		return onekiterrors.New(http.StatusBadRequest, "Provider is required")
	}
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateResponse{}, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	if !authPkg.IsBusinessAdmin(authUser) {
		return CreateResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return CreateResponse{}, onekiterrors.PermissionDenied()
	}
	id, err := h.Service.Create(ctx, authUser.UserID, businessID, domain.PaymentMethodInput{Provider: req.Provider, Config: req.Config, IsActive: req.IsActive})
	if err != nil {
		return CreateResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create payment method")
	}
	return CreateResponse{ID: id, Message: "Payment method created"}, nil
}

type (
	UpdateHandler struct{ Service *application.Service }
	UpdateRequest struct {
		ID       int64  `json:"id"`
		Provider string `json:"provider"`
		Config   string `json:"config"`
		IsActive bool   `json:"is_active"`
	}
)

type UpdateResponse struct {
	Message string `json:"message"`
}

func (h *UpdateHandler) Validate(_ context.Context, req *UpdateRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "ID is required")
	}
	return nil
}

func (h *UpdateHandler) Serve(ctx context.Context, req *UpdateRequest) (UpdateResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateResponse{}, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	if !authPkg.IsBusinessAdmin(authUser) {
		return UpdateResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return UpdateResponse{}, onekiterrors.PermissionDenied()
	}
	if err := h.Service.Update(ctx, businessID, req.ID, domain.PaymentMethodInput{Provider: req.Provider, Config: req.Config, IsActive: req.IsActive}); err != nil {
		return UpdateResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update payment method")
	}
	return UpdateResponse{Message: "Payment method updated"}, nil
}
