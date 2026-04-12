package handlers

import (
	posv1 "github.com/voxire/voxie/gen/go/pos/v1"

	"context"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/pos_manager/internal/application"
)

func VerifyPINRoutes() []string { return []string{"/pos/manager/verify"} }

type VerifyPINHandler struct{ Service *application.Service }

type VerifyPINRequest struct {
	PIN string `json:"pin"`
}

type VerifyPINResponse struct {
	Valid   bool   `json:"valid"`
	StaffID int64  `json:"staff_id"`
	Message string `json:"message"`
}

// verifyPINToProto converts to the proto response type for reference.
func verifyPINToProto(valid bool, staffID int64) *posv1.VerifyManagerPinResponse {
	return &posv1.VerifyManagerPinResponse{
		Valid:     valid,
		ManagerId: staffID,
	}
}

func (h *VerifyPINHandler) Validate(_ context.Context, req *VerifyPINRequest) *errorModels.ServerError {
	if req.PIN == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "PIN is required"}
	}
	return nil
}

func (h *VerifyPINHandler) Serve(ctx context.Context, req *VerifyPINRequest) (VerifyPINResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return VerifyPINResponse{}, errorModels.PermissionDeniedServerError()
	}

	result, err := h.Service.VerifyPIN(ctx, device.User.UserID, req.PIN)
	if err != nil {
		return VerifyPINResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to verify PIN"}
	}

	if !result.Valid {
		return VerifyPINResponse{Valid: false, Message: "Invalid PIN"}, nil
	}

	return VerifyPINResponse{
		Valid:   true,
		StaffID: result.StaffID,
		Message: "PIN verified",
	}, nil
}
