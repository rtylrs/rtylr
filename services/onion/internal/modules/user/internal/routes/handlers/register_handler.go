package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/netip"
	"strings"

	"github.com/voxire/onion/internal/modules/user/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	accountv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/shared/account/v1"
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
}

func (h *RegisterHandler) Validate(_ context.Context, req *accountv1.RegisterRequest) *errorModels.ServerError {
	if req.Email == "" || req.Password == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Email and password are required",
		}
	}

	if err := h.UserService.ValidateEmail(req.Email); err != nil {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid email format",
		}
	}

	if err := h.UserService.ValidatePassword(req.Password); err != nil {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Password must be at least 8 characters long and contain uppercase, lowercase, and numeric characters",
		}
	}

	return nil
}

func (h *RegisterHandler) Serve(ctx context.Context, req *accountv1.RegisterRequest) (accountv1.RegisterResponse, *errorModels.ServerError) {
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
		switch {
		case errors.Is(err, application.ErrUserAlreadyExists):
			return accountv1.RegisterResponse{}, &errorModels.ServerError{
				Code:    http.StatusConflict,
				Message: "User with this email already exists",
			}
		case errors.Is(err, application.ErrInvalidEmail):
			return accountv1.RegisterResponse{}, &errorModels.ServerError{
				Code:    http.StatusBadRequest,
				Message: "Invalid email format",
			}
		case errors.Is(err, application.ErrWeakPassword):
			return accountv1.RegisterResponse{}, &errorModels.ServerError{
				Code:    http.StatusBadRequest,
				Message: "Password does not meet security requirements",
			}
		default:
			return accountv1.RegisterResponse{}, &errorModels.ServerError{
				Code:    http.StatusTeapot,
				Message: "Registration failed",
			}
		}
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
		return accountv1.RegisterResponse{}, &errorModels.ServerError{
			Code:    http.StatusTeapot,
			Message: "Failed to generate authentication token",
		}
	}

	return accountv1.RegisterResponse{
		Token:   token,
		Message: "Registration successful",
		UserId:  userId,
		Email:   registerData.Email,
		Name:    registerData.FullName,
	}, nil
}
