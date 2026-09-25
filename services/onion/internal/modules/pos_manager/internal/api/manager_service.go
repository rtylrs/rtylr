package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/pos_manager/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	posmanagerv1 "github.com/voxire/voxie/gen/go/pos/manager/v1"
)

type ManagerAPI struct {
	Service *application.Service
}

func NewManagerAPI(service *application.Service) *ManagerAPI {
	return &ManagerAPI{Service: service}
}

func (a *ManagerAPI) VerifyPIN(ctx context.Context, req *posmanagerv1.VerifyPINRequest) (*posmanagerv1.VerifyPINResponse, error) {
	pin := strings.TrimSpace(req.GetPin())
	if pin == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "PIN is required")
	}

	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	result, err := a.Service.VerifyPIN(ctx, businessID, pin)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to verify PIN")
	}
	if !result.Valid {
		return &posmanagerv1.VerifyPINResponse{Valid: false, Message: "Invalid PIN"}, nil
	}

	return &posmanagerv1.VerifyPINResponse{Valid: true, StaffId: result.StaffID, Message: "PIN verified"}, nil
}
