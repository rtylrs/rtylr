package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/user/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	accountv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/shared/account/v1"
)

func LoginRoutes() []string {
	return []string{
		"/login",
		"/auth/login",
	}
}

type LoginHandler struct {
	UserService         *application.Service
	ProvisionerResolver authPkg.PlatformProvisionerResolver
}

func (h *LoginHandler) Validate(_ context.Context, req *accountv1.LoginRequest) *errorModels.ServerError {
	if req.Email == "" || req.Password == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Email and password are required",
		}
	}

	return nil
}

func (h *LoginHandler) Serve(ctx context.Context, req *accountv1.LoginRequest) (accountv1.LoginResponse, *errorModels.ServerError) {
	userId, err := h.UserService.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, application.ErrInvalidCredentials) {
			return accountv1.LoginResponse{}, &errorModels.ServerError{
				Code:    http.StatusUnauthorized,
				Message: "Invalid email or password",
			}
		}

		return accountv1.LoginResponse{}, &errorModels.ServerError{
			Code:    http.StatusTeapot,
			Message: "Authentication failed",
		}
	}

	var platformClaims *application.PlatformClaims

	productPlatform := authPkg.ProductPlatformCtx(ctx)
	if h.ProvisionerResolver != nil {
		provisioner := h.ProvisionerResolver.ResolveProvisioner(productPlatform)
		if provisioner != nil {
			pc, resolveErr := provisioner.ResolvePlatformClaims(ctx, userId)
			if resolveErr == nil && pc != nil {
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
		return accountv1.LoginResponse{}, &errorModels.ServerError{
			Code:    http.StatusTeapot,
			Message: "Failed to generate authentication token",
		}
	}

	user, err := h.UserService.GetUserByID(ctx, userId)
	if err != nil {
		return accountv1.LoginResponse{}, &errorModels.ServerError{
			Code:    http.StatusTeapot,
			Message: "Failed to retrieve user information",
		}
	}

	return accountv1.LoginResponse{
		Token:   token,
		Message: "Login successful",
		UserId:  userId,
		Email:   user.Email,
		Name:    user.FullName,
	}, nil
}
