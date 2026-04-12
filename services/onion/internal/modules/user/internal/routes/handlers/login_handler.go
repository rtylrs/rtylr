package handlers

import (
	"context"
	"errors"
	"net/http"

	authv1 "github.com/voxire/voxie/gen/go/auth/v1"

	"github.com/voxire/onion/internal/modules/user/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
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

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse = authv1.LoginResponse

func (h *LoginHandler) Validate(_ context.Context, req *LoginRequest) *errorModels.ServerError {
	if req.Email == "" || req.Password == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Email and password are required",
		}
	}

	return nil
}

func (h *LoginHandler) Serve(ctx context.Context, req *LoginRequest) (LoginResponse, *errorModels.ServerError) {
	userId, err := h.UserService.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, application.ErrInvalidCredentials) {
			return LoginResponse{}, &errorModels.ServerError{
				Code:    http.StatusUnauthorized,
				Message: "Invalid email or password",
			}
		}

		return LoginResponse{}, &errorModels.ServerError{
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
		return LoginResponse{}, &errorModels.ServerError{
			Code:    http.StatusTeapot,
			Message: "Failed to generate authentication token",
		}
	}

	user, err := h.UserService.GetUserByID(ctx, userId)
	if err != nil {
		return LoginResponse{}, &errorModels.ServerError{
			Code:    http.StatusTeapot,
			Message: "Failed to retrieve user information",
		}
	}

	// Record login in history (non-blocking, through service -> repo)
	go func() {
		ip := ""
		if device := authPkg.DeviceCtx(ctx); device.User != nil {
			ip = device.User.IP
		}
		_ = h.UserService.RecordLogin(context.Background(), userId, req.Email, "email", ip, true)
	}()

	return LoginResponse{
		Token:   token,
		Message: "Login successful",
		UserId:  userId,
		Email:   user.Email,
		Name:    user.FullName,
	}, nil
}
