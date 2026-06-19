package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/store_checkout/internal/application"
	"github.com/voxire/onion/internal/modules/store_checkout/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
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

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
	if req.SessionID == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "session_id is required"}
	}
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return CreateResponse{}, errorModels.PermissionDeniedServerError()
	}

	id, err := h.Service.Create(ctx, businessID, authUser.UserID, domain.CheckoutCreateInput{
		SessionID: req.SessionID, CustomerEmail: req.CustomerEmail, Items: req.Items,
		Subtotal: req.Subtotal, Tax: req.Tax, Total: req.Total,
	})
	if err != nil {
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to create checkout"}
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

func (h *CompleteHandler) Validate(_ context.Context, req *CompleteRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "ID is required"}
	}
	return nil
}

func (h *CompleteHandler) Serve(ctx context.Context, req *CompleteRequest) (CompleteResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CompleteResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return CompleteResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.Complete(ctx, businessID, req.ID, authUser.UserID); err != nil {
		return CompleteResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to complete checkout"}
	}
	return CompleteResponse{Message: "Checkout completed"}, nil
}
