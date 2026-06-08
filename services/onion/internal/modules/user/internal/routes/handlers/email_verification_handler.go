package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func VerifyEmailTokenRoutes() []string {
	return []string{"/auth/verify/{token}"}
}

func VerifyEmailBodyRoutes() []string {
	return []string{"/auth/verify-email"}
}

func ResendVerificationRoutes() []string {
	return []string{"/auth/resend-verification"}
}

type VerifyEmailHandler struct{}

type VerifyEmailRequest struct {
	Token string `json:"token"`
}

type VerifyEmailResponse struct {
	Verified bool   `json:"verified"`
	Message  string `json:"message"`
}

func (h *VerifyEmailHandler) Validate(_ context.Context, req *VerifyEmailRequest) *errorModels.ServerError {
	req.Token = strings.TrimSpace(req.Token)
	if req.Token == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "token is required"}
	}
	return nil
}

func (h *VerifyEmailHandler) Serve(ctx context.Context, req *VerifyEmailRequest) (VerifyEmailResponse, *errorModels.ServerError) {
	if req.Token == "" {
		req.Token = strings.TrimSpace(chi.URLParamFromCtx(ctx, "token"))
	}
	if err := h.Validate(ctx, req); err != nil {
		return VerifyEmailResponse{}, err
	}

	return VerifyEmailResponse{Verified: true, Message: "Email verified"}, nil
}

type ResendVerificationHandler struct{}

type ResendVerificationRequest struct {
	Email string `json:"email"`
}

type ResendVerificationResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

func (h *ResendVerificationHandler) Validate(_ context.Context, req *ResendVerificationRequest) *errorModels.ServerError {
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "email is required"}
	}
	return nil
}

func (h *ResendVerificationHandler) Serve(_ context.Context, _ *ResendVerificationRequest) (ResendVerificationResponse, *errorModels.ServerError) {
	return ResendVerificationResponse{OK: true, Message: "If that email is registered, a verification email has been sent."}, nil
}
