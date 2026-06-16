package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/business/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
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

func (h *CreateSubBusinessHandler) Validate(_ context.Context, req *CreateSubBusinessRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "name is required"}
	}
	if req.ParentBusinessID <= 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "parent_business_id is required"}
	}
	return nil
}

func (h *CreateSubBusinessHandler) Serve(ctx context.Context, req *CreateSubBusinessRequest) (CreateSubBusinessResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateSubBusinessResponse{}, errorModels.PermissionDeniedServerError()
	}

	id, err := h.Service.CreateSubBusiness(ctx, req.Name, device.User.UserID, req.ParentBusinessID)
	if err != nil {
		if errors.Is(err, application.ErrSubBusinessNotOwned) {
			return CreateSubBusinessResponse{}, &errorModels.ServerError{
				HTTPCode: http.StatusForbidden,
				Code:     http.StatusForbidden,
				Message:  "Parent business not found or not owned by you",
			}
		}
		if errors.Is(err, application.ErrSubBusinessNameEmpty) {
			return CreateSubBusinessResponse{}, &errorModels.ServerError{
				HTTPCode: http.StatusBadRequest,
				Code:     http.StatusBadRequest,
				Message:  "Business name is required",
			}
		}
		return CreateSubBusinessResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to create business",
		}
	}
	return CreateSubBusinessResponse{ID: id}, nil
}
