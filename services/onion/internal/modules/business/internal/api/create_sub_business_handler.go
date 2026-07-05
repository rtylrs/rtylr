package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/business/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func CreateSubBusinessRoutes() []string { return []string{"/businesses"} }

type (
	CreateSubBusinessHandler struct{ Service *application.Service }
	CreateSubBusinessRequest struct {
		Name             string `json:"name"`
		ParentBusinessID int64  `json:"parent_business_id"`
	}
	CreateSubBusinessResponse struct {
		ID int64 `json:"id"`
	}
)

func (h *CreateSubBusinessHandler) Validate(_ context.Context, req *CreateSubBusinessRequest) error {
	if req.Name == "" {
		return onekiterrors.New(http.StatusBadRequest, "name is required")
	}
	if req.ParentBusinessID <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "parent_business_id is required")
	}
	return nil
}

func (h *CreateSubBusinessHandler) Serve(ctx context.Context, req *CreateSubBusinessRequest) (CreateSubBusinessResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateSubBusinessResponse{}, onekiterrors.PermissionDenied()
	}

	id, err := h.Service.CreateSubBusiness(ctx, req.Name, device.User.UserID, req.ParentBusinessID)
	if err != nil {
		if errors.Is(err, application.ErrSubBusinessNotOwned) {
			return CreateSubBusinessResponse{}, onekiterrors.New(http.StatusForbidden, "Parent business not found or not owned by you")
		}
		if errors.Is(err, application.ErrSubBusinessNameEmpty) {
			return CreateSubBusinessResponse{}, onekiterrors.New(http.StatusBadRequest, "Business name is required")
		}
		return CreateSubBusinessResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create business")
	}
	return CreateSubBusinessResponse{ID: id}, nil
}
