package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/console/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

type createResourceRowRequest struct {
	Values map[string]any `json:"values" validate:"required,min=1"`
}

type createResourceRowResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

type createResourceRowHandler struct {
	Service *application.Service
}

func (h *createResourceRowHandler) Validate(_ context.Context, req *createResourceRowRequest) error {
	if req == nil || len(req.Values) == 0 {
		return onekiterrors.New(http.StatusBadRequest, "values are required")
	}
	return nil
}

func (h *createResourceRowHandler) Serve(ctx context.Context, req *createResourceRowRequest) (createResourceRowResponse, error) {
	if err := requireConsoleResourceAdmin(ctx); err != nil {
		return createResourceRowResponse{}, err
	}
	id, err := h.Service.CreateResourceRow(ctx, chi.URLParamFromCtx(ctx, "table"), req.Values)
	if err != nil {
		return createResourceRowResponse{}, onekiterrors.New(http.StatusBadRequest, err.Error())
	}
	return createResourceRowResponse{ID: id, Message: "Resource row created"}, nil
}

type deleteResourceRowRequest struct{}

type resourceMutationResponse struct {
	Message string `json:"message"`
}

type deleteResourceRowHandler struct {
	Service *application.Service
}

func (h *deleteResourceRowHandler) Validate(ctx context.Context, _ *deleteResourceRowRequest) error {
	return validateIDPath(ctx, "id", "Invalid id")
}

func (h *deleteResourceRowHandler) Serve(ctx context.Context, _ *deleteResourceRowRequest) (resourceMutationResponse, error) {
	if err := requireConsoleResourceAdmin(ctx); err != nil {
		return resourceMutationResponse{}, err
	}
	if err := h.Service.DeleteResourceRow(ctx, chi.URLParamFromCtx(ctx, "table"), chi.URLParamFromCtx(ctx, "id")); err != nil {
		return resourceMutationResponse{}, onekiterrors.New(http.StatusBadRequest, err.Error())
	}
	return resourceMutationResponse{Message: "Resource row deleted"}, nil
}

type grantIDRequest struct{}

func validateGrantID(ctx context.Context, _ *grantIDRequest) error {
	return validateIDPath(ctx, "id", "Invalid id")
}

type revokeGrantHandler struct {
	Service *application.Service
}

func (h *revokeGrantHandler) Validate(ctx context.Context, req *grantIDRequest) error {
	return validateGrantID(ctx, req)
}

func (h *revokeGrantHandler) Serve(ctx context.Context, _ *grantIDRequest) (resourceMutationResponse, error) {
	if err := requireConsoleResourceAdmin(ctx); err != nil {
		return resourceMutationResponse{}, err
	}
	id, err := parseIDPath(ctx, "id")
	if err != nil {
		return resourceMutationResponse{}, err
	}
	if err := h.Service.RevokeGrant(ctx, id); err != nil {
		return resourceMutationResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to revoke grant")
	}
	return resourceMutationResponse{Message: "Grant revoked"}, nil
}

type deleteGrantHandler struct {
	Service *application.Service
}

func (h *deleteGrantHandler) Validate(ctx context.Context, req *grantIDRequest) error {
	return validateGrantID(ctx, req)
}

func (h *deleteGrantHandler) Serve(ctx context.Context, _ *grantIDRequest) (resourceMutationResponse, error) {
	if err := requireConsoleResourceAdmin(ctx); err != nil {
		return resourceMutationResponse{}, err
	}
	id, err := parseIDPath(ctx, "id")
	if err != nil {
		return resourceMutationResponse{}, err
	}
	if err := h.Service.DeleteGrant(ctx, id); err != nil {
		return resourceMutationResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to delete grant")
	}
	return resourceMutationResponse{Message: "Grant deleted"}, nil
}

type revokeAddonGrantHandler struct {
	Service *application.Service
}

func (h *revokeAddonGrantHandler) Validate(ctx context.Context, req *grantIDRequest) error {
	return validateGrantID(ctx, req)
}

func (h *revokeAddonGrantHandler) Serve(ctx context.Context, _ *grantIDRequest) (resourceMutationResponse, error) {
	if err := requireConsoleResourceAdmin(ctx); err != nil {
		return resourceMutationResponse{}, err
	}
	id, err := parseIDPath(ctx, "id")
	if err != nil {
		return resourceMutationResponse{}, err
	}
	if err := h.Service.RevokeAddonGrant(ctx, id); err != nil {
		return resourceMutationResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to revoke addon grant")
	}
	return resourceMutationResponse{Message: "Addon grant revoked"}, nil
}

type deleteAddonGrantHandler struct {
	Service *application.Service
}

func (h *deleteAddonGrantHandler) Validate(ctx context.Context, req *grantIDRequest) error {
	return validateGrantID(ctx, req)
}

func (h *deleteAddonGrantHandler) Serve(ctx context.Context, _ *grantIDRequest) (resourceMutationResponse, error) {
	if err := requireConsoleResourceAdmin(ctx); err != nil {
		return resourceMutationResponse{}, err
	}
	id, err := parseIDPath(ctx, "id")
	if err != nil {
		return resourceMutationResponse{}, err
	}
	if err := h.Service.DeleteAddonGrant(ctx, id); err != nil {
		return resourceMutationResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to delete addon grant")
	}
	return resourceMutationResponse{Message: "Addon grant deleted"}, nil
}

type createSelfHostLicenseRequest struct {
	OwnerEmail   string   `json:"owner_email" validate:"required"`
	CustomerName string   `json:"customer_name"`
	Plan         string   `json:"plan"`
	Features     []string `json:"features"`
	SeatLimit    int      `json:"seat_limit"`
	ExpiresAt    string   `json:"expires_at"`
	GraceUntil   string   `json:"grace_until"`
}

func (r createSelfHostLicenseRequest) input() application.CreateSelfHostLicenseInput {
	return application.CreateSelfHostLicenseInput{
		OwnerEmail:   r.OwnerEmail,
		CustomerName: r.CustomerName,
		Plan:         r.Plan,
		Features:     r.Features,
		SeatLimit:    r.SeatLimit,
		ExpiresAt:    r.ExpiresAt,
		GraceUntil:   r.GraceUntil,
	}
}

type createSelfHostLicenseHandler struct {
	Service *application.Service
}

func (h *createSelfHostLicenseHandler) Serve(ctx context.Context, req *createSelfHostLicenseRequest) (application.SelfHostLicenseKeyResult, error) {
	admin, err := consoleAdmin(ctx)
	if err != nil {
		return application.SelfHostLicenseKeyResult{}, err
	}
	result, err := h.Service.CreateSelfHostLicense(ctx, req.input(), admin.UserID)
	if err != nil {
		return application.SelfHostLicenseKeyResult{}, selfHostLicenseError(err)
	}
	return *result, nil
}

type licenseIDRequest struct{}

func (h *createSelfHostLicenseHandler) Validate(_ context.Context, _ *createSelfHostLicenseRequest) error {
	return nil
}

type regenerateSelfHostLicenseHandler struct {
	Service *application.Service
}

func (h *regenerateSelfHostLicenseHandler) Validate(ctx context.Context, _ *licenseIDRequest) error {
	return validateIDPath(ctx, "id", "Invalid self-host license ID")
}

func (h *regenerateSelfHostLicenseHandler) Serve(ctx context.Context, _ *licenseIDRequest) (application.SelfHostLicenseKeyResult, error) {
	admin, err := consoleAdmin(ctx)
	if err != nil {
		return application.SelfHostLicenseKeyResult{}, err
	}
	id, err := parseIDPath(ctx, "id")
	if err != nil {
		return application.SelfHostLicenseKeyResult{}, err
	}
	result, err := h.Service.RegenerateSelfHostLicenseKey(ctx, id, admin.UserID)
	if err != nil {
		return application.SelfHostLicenseKeyResult{}, selfHostLicenseError(err)
	}
	return *result, nil
}

type revokeSelfHostLicenseHandler struct {
	Service *application.Service
}

func (h *revokeSelfHostLicenseHandler) Validate(ctx context.Context, _ *licenseIDRequest) error {
	return validateIDPath(ctx, "id", "Invalid self-host license ID")
}

func (h *revokeSelfHostLicenseHandler) Serve(ctx context.Context, _ *licenseIDRequest) (application.SelfHostLicenseDTO, error) {
	admin, err := consoleAdmin(ctx)
	if err != nil {
		return application.SelfHostLicenseDTO{}, err
	}
	id, err := parseIDPath(ctx, "id")
	if err != nil {
		return application.SelfHostLicenseDTO{}, err
	}
	result, err := h.Service.RevokeSelfHostLicense(ctx, id, admin.UserID)
	if err != nil {
		return application.SelfHostLicenseDTO{}, selfHostLicenseError(err)
	}
	return *result, nil
}

func consoleAdmin(ctx context.Context) (*authPkg.AuthUser, error) {
	admin, ok := authPkg.GetAuthUser(ctx)
	if !ok || admin == nil {
		return nil, onekiterrors.New(http.StatusUnauthorized, "Unauthorized")
	}
	if !authPkg.IsConsoleAdmin(admin) {
		return nil, onekiterrors.New(http.StatusForbidden, "Console admin access required")
	}
	return admin, nil
}

func requireConsoleResourceAdmin(ctx context.Context) error {
	_, err := consoleAdmin(ctx)
	return err
}

func validateIDPath(ctx context.Context, param, message string) error {
	_, err := parseIDPath(ctx, param)
	if err != nil {
		return onekiterrors.New(http.StatusBadRequest, message)
	}
	return nil
}

func parseIDPath(ctx context.Context, param string) (int64, error) {
	raw := chi.URLParamFromCtx(ctx, param)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid path id")
	}
	return id, nil
}

func selfHostLicenseError(err error) error {
	switch {
	case errors.Is(err, application.ErrSelfHostLicenseNotFound):
		return onekiterrors.New(http.StatusNotFound, "Self-host license not found")
	case errors.Is(err, application.ErrInvalidSelfHostLicenseInput):
		return onekiterrors.New(http.StatusBadRequest, strings.TrimSpace(err.Error()))
	case errors.Is(err, application.ErrSelfHostLicenseSigningKeyMissing):
		return onekiterrors.New(http.StatusInternalServerError, "License signing key is not configured")
	default:
		return onekiterrors.New(http.StatusInternalServerError, "Self-host license request failed")
	}
}
