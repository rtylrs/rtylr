package routes

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/console/internal/application"
	"github.com/voxire/onion/internal/modules/console/internal/routes/handlers"
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
	router.Post("/console/resources/{table}/rows", rawCreateResourceRow(ctx, service))
	router.Get("/console/resources/{table}/rows/{id}", rawGetResourceRow(ctx, service))
	serverHttp.RegisterPUTWithValidation[updateResourceRowRequest, updateResourceRowResponse](
		ctx, router, []string{"/console/resources/{table}/rows/{id}"},
		&updateResourceRowHandler{Service: service},
	)
	router.Delete("/console/resources/{table}/rows/{id}", rawDeleteResourceRow(ctx, service))

	// AI cache analytics has not moved to the generated console contract yet.
	serverHttp.RegisterGET[handlers.AICacheAnalyticsRequest, handlers.AICacheAnalyticsResponse](
		ctx, router, []string{"/console/ai-cache/analytics"},
		&handlers.AICacheAnalyticsHandler{},
	)

	// Grants
	serverHttp.RegisterGET[handlers.ListGrantsRequest, handlers.ListGrantsResponse](
		ctx, router, []string{"/console/grants"},
		&handlers.ListGrantsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateGrantRequest, handlers.CreateGrantResponse](
		ctx, router, []string{"/console/grants"},
		&handlers.CreateGrantHandler{Service: service},
	)
	router.Post("/console/grants/{id}/revoke", rawRevokeGrant(ctx, service))
	router.Delete("/console/grants/{id}", rawDeleteGrant(ctx, service))

	// Addon Grants
	serverHttp.RegisterGET[handlers.ListAddonsRequest, handlers.ListAddonsResponse](
		ctx, router, []string{"/console/addons"},
		&handlers.ListAddonsHandler{Service: service},
	)
	serverHttp.RegisterGET[handlers.ListAddonGrantsRequest, handlers.ListAddonGrantsResponse](
		ctx, router, []string{"/console/addon-grants"},
		&handlers.ListAddonGrantsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateAddonGrantRequest, handlers.CreateAddonGrantResponse](
		ctx, router, []string{"/console/addon-grants"},
		&handlers.CreateAddonGrantHandler{Service: service},
	)
	router.Post("/console/addon-grants/{id}/revoke", rawRevokeAddonGrant(ctx, service))
	router.Delete("/console/addon-grants/{id}", rawDeleteAddonGrant(ctx, service))

	// Self-host licenses
	router.Get("/console/self-host-licenses", rawListSelfHostLicenses(ctx, service))
	router.Post("/console/self-host-licenses", rawCreateSelfHostLicense(ctx, service))
	router.Post("/console/self-host-licenses/{id}/regenerate", rawRegenerateSelfHostLicenseKey(ctx, service))
	router.Post("/console/self-host-licenses/{id}/revoke", rawRevokeSelfHostLicense(ctx, service))
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

func getIDParam(r *http.Request, param string) (int64, bool) {
	raw := chi.URLParam(r, param)
	id, err := strconv.ParseInt(raw, 10, 64)
	return id, err == nil && id > 0
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

func rawCreateResourceRow(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		var req struct {
			Values map[string]any `json:"values"`
		}
		if !decodeJSONBody(w, r, &req) {
			return
		}
		if req.Values == nil {
			writeError(w, http.StatusBadRequest, "values are required")
			return
		}
		id, err := svc.CreateResourceRow(r.Context(), chi.URLParam(r, "table"), req.Values)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"id": id, "message": "Resource row created"})
	}
}

func rawDeleteResourceRow(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		if err := svc.DeleteResourceRow(r.Context(), chi.URLParam(r, "table"), chi.URLParam(r, "id")); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "Resource row deleted"})
	}
}

func rawRevokeGrant(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		id, ok := getIDParam(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, "Invalid id")
			return
		}
		if err := svc.RevokeGrant(r.Context(), id); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to revoke grant")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "Grant revoked"})
	}
}

func rawDeleteGrant(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		id, ok := getIDParam(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, "Invalid id")
			return
		}
		if err := svc.DeleteGrant(r.Context(), id); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to delete grant")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "Grant deleted"})
	}
}

func rawRevokeAddonGrant(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		id, ok := getIDParam(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, "Invalid id")
			return
		}
		if err := svc.RevokeAddonGrant(r.Context(), id); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to revoke addon grant")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "Addon grant revoked"})
	}
}

func rawDeleteAddonGrant(_ context.Context, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := checkAdmin(r); !ok {
			writeError(w, http.StatusForbidden, "Console admin access required")
			return
		}
		id, ok := getIDParam(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, "Invalid id")
			return
		}
		if err := svc.DeleteAddonGrant(r.Context(), id); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to delete addon grant")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "Addon grant deleted"})
	}
}
