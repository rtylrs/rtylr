package handlers

import (
	"context"
	"net/http"
	"net/netip"
	"strings"

	"github.com/voxire/onion/internal/modules/user/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	commonauthv1 "github.com/voxire/voxie/gen/go/common/auth/v1"
)

func RegisterRoutes() []string {
	return []string{
		"/register",
		"/auth/register",
	}
}

type RegisterHandler struct {
	UserService         *application.Service
	ProvisionerResolver authPkg.PlatformProvisionerResolver
	// OnUserCreated is called non-fatally after a user is successfully registered.
	// Used to create a personal workspace, provision initial data, etc.
	OnUserCreated func(ctx context.Context, userID int64)
}

type RegisterRequest struct {
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=8"`
	Name         string `json:"name"`
	BusinessName string `json:"business_name"`
	PlanSlug     string `json:"plan_slug"`
}

type RegisterResponse = commonauthv1.RegisterResponse

func (h *RegisterHandler) Serve(ctx context.Context, req *RegisterRequest) (RegisterResponse, *errorModels.ServerError) {
	var clientIP netip.Addr

	var deviceInfo string

	registerData := application.RegisterUserData{
		Email:           strings.TrimSpace(req.Email),
		Password:        req.Password,
		FullName:        strings.TrimSpace(req.Name),
		SignupIPAddress: clientIP,
		DeviceInfo:      deviceInfo,
	}

	userId, err := h.UserService.Register(ctx, registerData)
	if err != nil {
		return RegisterResponse{}, errorModels.FromError(err, "Registration failed")
	}

	// Create business for this user (they are the owner)
	businessName := strings.TrimSpace(req.BusinessName)
	if businessName == "" {
		businessName = strings.TrimSpace(req.Name) + "'s Business"
	}

	if bizErr := h.UserService.CreateBusinessForUser(ctx, userId, businessName); bizErr != nil {
		// Non-fatal: the user is created even if business bootstrap fails, but
		// surface the error rather than silently discarding it.
		logger.Ctx(ctx).Warn("failed to create business for new user",
			logger.String("business_name", businessName),
			logger.Error(bizErr),
		)
	}

	// New owner accounts always start on the free plan. Paid upgrades happen
	// after signup through billing/checkout so registration is deterministic.
	planSlug := application.DefaultSignupPlanSlug
	if subErr := h.UserService.AssignSubscription(ctx, userId, planSlug); subErr != nil {
		logger.Ctx(ctx).Warn("failed to assign subscription on registration",
			logger.String("plan_slug", planSlug),
			logger.Error(subErr),
		)
	}

	// Create personal workspace for the new user (non-fatal)
	if h.OnUserCreated != nil {
		runAsyncUserHook(ctx, userId, h.OnUserCreated)
	}

	var platformClaims *application.PlatformClaims

	productPlatform := authPkg.ProductPlatformCtx(ctx)
	if h.ProvisionerResolver != nil {
		provisioner := h.ProvisionerResolver.ResolveProvisioner(productPlatform)
		if provisioner != nil {
			pc, provisionErr := provisioner.ProvisionUser(ctx, userId)
			if provisionErr == nil && pc != nil {
				platformClaims = &application.PlatformClaims{
					Platform:       pc.Platform,
					PlatformUserID: pc.PlatformUserID,
					PlanID:         pc.PlanID,
				}
			}
		}
	}

	token, err := h.UserService.GenerateTokenWithPlatform(ctx, userId, platformClaims)
	if err != nil {
		return RegisterResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to generate authentication token",
		}
	}

	return RegisterResponse{
		Token:   token,
		Message: "Registration successful",
		UserId:  userId,
		Email:   registerData.Email,
		Name:    registerData.FullName,
	}, nil
}
