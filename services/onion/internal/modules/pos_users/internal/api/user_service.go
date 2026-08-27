package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/pos_users/internal/application"
	"github.com/voxire/onion/internal/modules/pos_users/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	posuserv1 "github.com/voxire/voxie/gen/go/pos/user/v1"
)

type UserAPI struct {
	Service *application.Service
}

func NewUserAPI(service *application.Service) *UserAPI {
	return &UserAPI{Service: service}
}

func (a *UserAPI) ListPOSUsers(ctx context.Context, req *posuserv1.ListPOSUsersRequest) (*posuserv1.ListPOSUsersResponse, error) {
	if !hasPOSUserManagement(ctx) {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, err := authenticatedPOSUserBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	var users []domain.POSUser
	var listErr error
	if req.GetLocationId() > 0 {
		users, listErr = a.Service.ListByLocation(ctx, businessID, req.GetLocationId())
	} else {
		users, listErr = a.Service.ListByOwner(ctx, businessID)
	}
	if listErr != nil {
		logger.Ctx(ctx).Error("failed to list pos_users",
			logger.Any("business_id", businessID),
			logger.Error(listErr),
		)
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list POS users")
	}

	out := make([]*posuserv1.POSUser, 0, len(users))
	for _, u := range users {
		out = append(out, toProtoPOSUser(u))
	}
	return &posuserv1.ListPOSUsersResponse{Users: out}, nil
}

func (a *UserAPI) CreatePOSUser(ctx context.Context, req *posuserv1.CreatePOSUserRequest) (*posuserv1.CreatePOSUserResponse, error) {
	if !hasPOSUserManagement(ctx) {
		return nil, onekiterrors.PermissionDenied()
	}
	device, businessID, err := authenticatedPOSUserDevice(ctx)
	if err != nil {
		return nil, err
	}

	id, createErr := a.Service.Create(ctx, businessID, device.User.UserID, domain.CreatePOSUserInput{
		UserID:           req.GetUserId(),
		LocationID:       req.GetLocationId(),
		Name:             req.GetName(),
		PIN:              req.GetPin(),
		Role:             domain.RoleCashier,
		HourlyRate:       req.GetHourlyRate(),
		MaxDiscountPct:   domain.DefaultCapabilities(domain.RoleCashier).MaxDiscountPct,
		CanVoid:          domain.DefaultCapabilities(domain.RoleCashier).CanVoid,
		CanRefund:        domain.DefaultCapabilities(domain.RoleCashier).CanRefund,
		CanOpenDrawer:    domain.DefaultCapabilities(domain.RoleCashier).CanOpenDrawer,
		CanApplyDiscount: domain.DefaultCapabilities(domain.RoleCashier).CanApplyDiscount,
	})
	if createErr != nil {
		if errors.Is(createErr, application.ErrInvalidInput) {
			return nil, onekiterrors.New(http.StatusBadRequest, "Invalid input")
		}
		logger.Ctx(ctx).Error("failed to create pos_user",
			logger.Any("business_id", businessID),
			logger.Error(createErr),
		)
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create POS user")
	}
	return &posuserv1.CreatePOSUserResponse{Id: id, RoleManagement: roleManagementManagedInDash()}, nil
}

func (a *UserAPI) UpdatePOSUser(ctx context.Context, req *posuserv1.UpdatePOSUserRequest) (*posuserv1.UpdatePOSUserResponse, error) {
	if !hasPOSUserManagement(ctx) {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, err := authenticatedPOSUserBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "ID is required")
	}

	updateErr := a.Service.Update(ctx, businessID, req.GetId(), domain.UpdatePOSUserInput{
		Name:       req.GetName(),
		PIN:        req.GetPin(),
		IsActive:   req.GetIsActive(),
		HourlyRate: req.GetHourlyRate(),
	})
	if updateErr != nil {
		logger.Ctx(ctx).Error("failed to update pos_user",
			logger.Any("business_id", businessID),
			logger.Any("pos_user_id", req.GetId()),
			logger.Error(updateErr),
		)
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update POS user")
	}
	return &posuserv1.UpdatePOSUserResponse{Success: true, RoleManagement: roleManagementManagedInDash()}, nil
}

func (a *UserAPI) DeactivatePOSUser(ctx context.Context, req *posuserv1.DeactivatePOSUserRequest) (*posuserv1.DeactivatePOSUserResponse, error) {
	if !hasPOSUserManagement(ctx) {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, err := authenticatedPOSUserBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "ID is required")
	}

	if deactivateErr := a.Service.Deactivate(ctx, businessID, req.GetId()); deactivateErr != nil {
		logger.Ctx(ctx).Error("failed to deactivate pos_user",
			logger.Any("business_id", businessID),
			logger.Any("pos_user_id", req.GetId()),
			logger.Error(deactivateErr),
		)
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to deactivate POS user")
	}
	return &posuserv1.DeactivatePOSUserResponse{Success: true}, nil
}

func (a *UserAPI) SetPIN(ctx context.Context, req *posuserv1.SetPINRequest) (*posuserv1.SetPINResponse, error) {
	device, businessID, err := authenticatedPOSUserDevice(ctx)
	if err != nil {
		return nil, err
	}

	setErr := a.Service.SetPIN(ctx, businessID, application.SetPINActor{
		UserID:          device.User.UserID,
		IsOwner:         device.User.UserID == businessID,
		TeamRole:        device.User.Role,
		EffectiveAccess: device.User.EffectiveAccess,
	}, domain.SetPINInput{
		UserID:     req.GetUserId(),
		LocationID: req.GetLocationId(),
		Name:       req.GetName(),
		PIN:        req.GetPin(),
	})
	if setErr != nil {
		if errors.Is(setErr, application.ErrInvalidInput) {
			return nil, onekiterrors.New(http.StatusBadRequest, "Not assigned to this location")
		}
		if errors.Is(setErr, application.ErrForbidden) {
			return nil, onekiterrors.PermissionDenied()
		}
		if errors.Is(setErr, application.ErrPINAlreadySet) {
			return nil, onekiterrors.New(http.StatusConflict, "A PIN is already set. Ask a manager to reset it.")
		}
		logger.Ctx(ctx).Error("pos_users set-pin failed",
			logger.Any("business_id", businessID),
			logger.Any("target_user", req.GetUserId()),
			logger.Error(setErr),
		)
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to set PIN")
	}
	return &posuserv1.SetPINResponse{Success: true}, nil
}

func (a *UserAPI) VerifyPOSUserPIN(ctx context.Context, req *posuserv1.VerifyPOSUserPINRequest) (*posuserv1.VerifyPOSUserPINResponse, error) {
	if !hasPOSOrderAccess(ctx) {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, err := authenticatedPOSUserBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	result, verifyErr := a.Service.VerifyPIN(ctx, businessID, req.GetLocationId(), req.GetUserId(), req.GetPin())
	if verifyErr != nil {
		if errors.Is(verifyErr, domain.ErrInvalidPIN) {
			return nil, onekiterrors.New(http.StatusUnauthorized, "Invalid credentials")
		}
		logger.Ctx(ctx).Error("pos_users verify-pin failed",
			logger.Any("business_id", businessID),
			logger.Any("pos_user_target", req.GetUserId()),
			logger.Error(verifyErr),
		)
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to verify PIN")
	}

	return &posuserv1.VerifyPOSUserPINResponse{
		UserId:               result.UserID,
		PosUserId:            result.POSUserID,
		LocationId:           result.LocationID,
		Role:                 string(result.Role),
		DisplayName:          result.DisplayName,
		CanVoid:              result.CanVoid,
		CanRefund:            result.CanRefund,
		CanOpenDrawer:        result.CanOpenDrawer,
		CanApplyDiscount:     result.CanApplyDiscount,
		MaxDiscountPct:       result.MaxDiscountPct,
		AuthorizationVersion: result.AuthorizationVersion,
		RoleManagement:       roleManagementManagedInDash(),
	}, nil
}

func roleManagementManagedInDash() *posuserv1.RoleManagementManagedInDash {
	return &posuserv1.RoleManagementManagedInDash{
		Code:    "ROLE_MANAGEMENT_MANAGED_IN_DASH",
		DashUrl: strings.TrimRight(config.HubAppURL, "/") + "/team",
	}
}

func authenticatedPOSUserBusinessID(ctx context.Context) (int64, error) {
	_, businessID, err := authenticatedPOSUserDevice(ctx)
	return businessID, err
}

func hasPOSUserManagement(ctx context.Context) bool {
	device := authPkg.DeviceCtx(ctx)
	return device.User != nil && authPkg.HasAppPermission(
		device.User,
		authorization.AppPOS,
		"pos:employees:manage",
	) && authPkg.HasAnyAppScope(device.User, authorization.AppPOS)
}

func hasPOSOrderAccess(ctx context.Context) bool {
	device := authPkg.DeviceCtx(ctx)
	return device.User != nil && authPkg.HasAppPermission(
		device.User,
		authorization.AppPOS,
		"pos:orders:view",
	) && authPkg.HasAnyAppScope(device.User, authorization.AppPOS)
}

func authenticatedPOSUserDevice(ctx context.Context) (authPkg.Device, int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return device, 0, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return device, 0, onekiterrors.PermissionDenied()
	}
	return device, businessID, nil
}

func toProtoPOSUser(u domain.POSUser) *posuserv1.POSUser {
	return &posuserv1.POSUser{
		Id:               u.ID,
		BusinessId:       u.BusinessID,
		LocationId:       u.LocationID,
		Name:             u.Name,
		Role:             string(u.Role),
		IsActive:         u.IsActive,
		MaxDiscountPct:   u.MaxDiscountPct,
		CanVoid:          u.CanVoid,
		CanRefund:        u.CanRefund,
		CanOpenDrawer:    u.CanOpenDrawer,
		CanApplyDiscount: u.CanApplyDiscount,
		RoleManagement:   roleManagementManagedInDash(),
	}
}
