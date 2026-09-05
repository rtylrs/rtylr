package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/console/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

const consoleJSONBodyLimit = 4 << 20

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	// Generic admin resources
	router.Get("/console/resources", rawListResources(ctx, service))
	router.Get("/console/resources/{table}", rawDescribeResource(ctx, service))
	router.Get("/console/resources/{table}/rows", rawListResourceRows(ctx, service))
	serverHttp.RegisterPOSTWithValidation[createResourceRowRequest, createResourceRowResponse](
		ctx, router, []string{"/console/resources/{table}/rows"},
		&createResourceRowHandler{Service: service},
	)
	router.Get("/console/resources/{table}/rows/{id}", rawGetResourceRow(ctx, service))
	serverHttp.RegisterPUTWithValidation[updateResourceRowRequest, updateResourceRowResponse](
		ctx, router, []string{"/console/resources/{table}/rows/{id}"},
		&updateResourceRowHandler{Service: service},
	)
	serverHttp.RegisterDELETEWithValidation[deleteResourceRowRequest, resourceMutationResponse](
		ctx, router, []string{"/console/resources/{table}/rows/{id}"},
		&deleteResourceRowHandler{Service: service},
	)

	// AI cache analytics has not moved to the generated console contract yet.
	serverHttp.RegisterGET[AICacheAnalyticsRequest, AICacheAnalyticsResponse](
		ctx, router, []string{"/console/ai-cache/analytics"},
		&AICacheAnalyticsHandler{},
	)

	// Grants
	serverHttp.RegisterGET[ListGrantsRequest, ListGrantsResponse](
		ctx, router, []string{"/console/grants"},
		&ListGrantsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[CreateGrantRequest, CreateGrantResponse](
		ctx, router, []string{"/console/grants"},
		&CreateGrantHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[grantIDRequest, resourceMutationResponse](
		ctx, router, []string{"/console/grants/{id}/revoke"},
		&revokeGrantHandler{Service: service},
	)
	serverHttp.RegisterDELETEWithValidation[grantIDRequest, resourceMutationResponse](
		ctx, router, []string{"/console/grants/{id}"},
		&deleteGrantHandler{Service: service},
	)

	// Addon Grants
	serverHttp.RegisterGET[ListAddonsRequest, ListAddonsResponse](
		ctx, router, []string{"/console/addons"},
		&ListAddonsHandler{Service: service},
	)
	serverHttp.RegisterGET[ListAddonGrantsRequest, ListAddonGrantsResponse](
		ctx, router, []string{"/console/addon-grants"},
		&ListAddonGrantsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[CreateAddonGrantRequest, CreateAddonGrantResponse](
		ctx, router, []string{"/console/addon-grants"},
		&CreateAddonGrantHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[grantIDRequest, resourceMutationResponse](
		ctx, router, []string{"/console/addon-grants/{id}/revoke"},
		&revokeAddonGrantHandler{Service: service},
	)
	serverHttp.RegisterDELETEWithValidation[grantIDRequest, resourceMutationResponse](
		ctx, router, []string{"/console/addon-grants/{id}"},
		&deleteAddonGrantHandler{Service: service},
	)

	// Self-host licenses
	router.Get("/console/self-host-licenses", rawListSelfHostLicenses(ctx, service))
	serverHttp.RegisterPOSTWithValidation[createSelfHostLicenseRequest, application.SelfHostLicenseKeyResult](
		ctx, router, []string{"/console/self-host-licenses"},
		&createSelfHostLicenseHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[licenseIDRequest, application.SelfHostLicenseKeyResult](
		ctx, router, []string{"/console/self-host-licenses/{id}/regenerate"},
		&regenerateSelfHostLicenseHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[licenseIDRequest, application.SelfHostLicenseDTO](
		ctx, router, []string{"/console/self-host-licenses/{id}/revoke"},
		&revokeSelfHostLicenseHandler{Service: service},
	)
}

// ── helper ────────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logger.Global().Error("Failed to encode JSON response", logger.Error(err))
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	serverHttp.WriteOneKitError(context.Background(), w, onekiterrors.New(status, msg))
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, consoleJSONBodyLimit)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid body")
		return false
	}

	var extra struct{}
	if err := decoder.Decode(&extra); err != io.EOF {
		writeError(w, http.StatusBadRequest, "Invalid body")
		return false
	}

	return true
}

func decodeOptionalJSONBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	if r.Body == nil || r.Body == http.NoBody || r.ContentLength == 0 {
		return true
	}
	return decodeJSONBody(w, r, dst)
}

type updateResourceRowRequest struct {
	Values map[string]any `json:"values" validate:"required,min=1"`
}

type updateResourceRowResponse struct {
	Message string `json:"message"`
}

type updateResourceRowHandler struct {
	Service *application.Service
}

func (h *updateResourceRowHandler) Serve(ctx context.Context, req *updateResourceRowRequest) (updateResourceRowResponse, error) {
	user, ok := authPkg.GetAuthUser(ctx)
	if !ok || user.Role != "console_admin" {
		return updateResourceRowResponse{}, onekiterrors.New(http.StatusForbidden, "Console admin access required")
	}
	if err := h.Service.UpdateResourceRow(ctx, chi.URLParamFromCtx(ctx, "table"), chi.URLParamFromCtx(ctx, "id"), req.Values); err != nil {
		return updateResourceRowResponse{}, onekiterrors.New(http.StatusBadRequest, err.Error())
	}
	return updateResourceRowResponse{Message: "Resource row updated"}, nil
}

func getPositiveQueryInt(r *http.Request, name string, fallback int) int {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		logger.Global().Warn("invalid console query parameter",
			logger.String("parameter", name), logger.String("value", raw))
		return fallback
	}
	return value
}

func checkAdmin(r *http.Request) (*authPkg.AuthUser, bool) {
	user, ok := authPkg.GetAuthUser(r.Context())
	if !ok || user.Role != "console_admin" {
		return nil, false
	}
	return user, true
}

// ── raw handlers ──────────────────────────────────────────────────────────────

func rawListResources(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		result, err := svc.ListResources(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to list resources")
			return
		}
		writeJSON(w, http.StatusOK, result)
	}
}

func rawDescribeResource(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		table := chi.URLParam(r, "table")
		result, err := svc.DescribeResource(r.Context(), table)
		if err != nil {
			writeError(w, http.StatusNotFound, "Resource not found")
			return
		}
		writeJSON(w, http.StatusOK, result)
	}
}

func rawListResourceRows(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		table := chi.URLParam(r, "table")
		limit := getPositiveQueryInt(r, "limit", 50)
		offset := getPositiveQueryInt(r, "offset", 0)
		result, err := svc.ListResourceRows(r.Context(), table, limit, offset)
		if err != nil {
			writeError(w, http.StatusNotFound, "Resource not found")
			return
		}
		writeJSON(w, http.StatusOK, result)
	}
}

func rawGetResourceRow(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		result, err := svc.GetResourceRow(r.Context(), chi.URLParam(r, "table"), chi.URLParam(r, "id"))
		if err != nil {
			writeError(w, http.StatusNotFound, "Resource row not found")
			return
		}
		writeJSON(w, http.StatusOK, result)
	}
}
