package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/voxhire/internal/application"
	"github.com/voxire/onion/internal/modules/voxhire/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

type adminCtxKey string

const adminIDCtxKey adminCtxKey = "voxhire_admin_id"

// Route patterns — also returned to GetAnonymousRoutes so the global JWT
// middleware skips them (the admin routes self-guard via the admin JWT).
func AdminLoginRoute() string        { return "/voxhire/admin/login" }
func AdminAdminsRoute() string       { return "/voxhire/admin/admins" }
func AdminMeRoute() string           { return "/voxhire/admin/me" }
func AdminOpeningsRoute() string     { return "/voxhire/admin/openings" }
func AdminLeadsRoute() string        { return "/voxhire/admin/leads" }
func AdminApplicationsRoute() string { return "/voxhire/admin/openings/{id}/applications" }
func AdminAppStatusRoute() string    { return "/voxhire/admin/applications/{id}/status" }
func AdminAppCVRoute() string        { return "/voxhire/admin/applications/{id}/cv" }
func CareerApplyRoute() string       { return "/voxhire/careers/apply" }

// RecruitAnonymousRoutes lists every recruit/admin path that bypasses
// the platform JWT middleware.
func RecruitAnonymousRoutes() []string {
	return []string{
		AdminLoginRoute(),
		AdminAdminsRoute(),
		AdminMeRoute(),
		AdminOpeningsRoute(),
		AdminLeadsRoute(),
		AdminApplicationsRoute(),
		AdminAppStatusRoute(),
		AdminAppCVRoute(),
		CareerApplyRoute(),
	}
}

// AdminAuthMiddleware validates the admin bearer token and injects the
// admin id into the request context.
func AdminAuthMiddleware(svc *application.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r)
			if token == "" {
				writeApplyError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}
			claims, err := svc.ValidateAdminToken(r.Context(), token)
			if err != nil {
				writeApplyError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}
			ctx := context.WithValue(r.Context(), adminIDCtxKey, claims.AdminID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if rest, ok := strings.CutPrefix(h, "Bearer "); ok {
		return rest
	}
	return h
}

func adminIDFromCtx(ctx context.Context) int64 {
	if v, ok := ctx.Value(adminIDCtxKey).(int64); ok {
		return v
	}
	return 0
}

// ── Login ────────────────────────────────────────────────────────────

func AdminLoginHandler(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeApplyError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		token, admin, err := svc.AdminLogin(r.Context(), body.Email, body.Password)
		if err != nil {
			if errors.Is(err, domain.ErrInvalidCredentials) {
				writeApplyError(w, http.StatusUnauthorized, "Invalid email or password")
				return
			}
			logger.Ctx(r.Context()).Error("voxhire admin login failed", logger.Error(err))
			writeApplyError(w, http.StatusInternalServerError, "Login failed")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"token": token, "admin": admin})
	}
}

// ── Admin accounts ───────────────────────────────────────────────────

func AdminMeHandler(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		admin, err := svc.GetAdmin(r.Context(), adminIDFromCtx(r.Context()))
		if err != nil {
			writeApplyError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"admin": admin})
	}
}

func ListAdminsHandler(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		admins, err := svc.ListAdmins(r.Context())
		if err != nil {
			logger.Ctx(r.Context()).Error("voxhire list admins failed", logger.Error(err))
			writeApplyError(w, http.StatusInternalServerError, "Failed to list admins")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"admins": admins})
	}
}

func CreateAdminHandler(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
			Name     string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeApplyError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		admin, err := svc.CreateAdmin(r.Context(), body.Email, body.Password, body.Name)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrAdminExists):
				writeApplyError(w, http.StatusConflict, "An admin with that email already exists")
			case errors.Is(err, domain.ErrInvalidInput):
				writeApplyError(w, http.StatusBadRequest, "Email and a password of at least 6 characters are required")
			default:
				logger.Ctx(r.Context()).Error("voxhire create admin failed", logger.Error(err))
				writeApplyError(w, http.StatusInternalServerError, "Failed to create admin")
			}
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"admin": admin})
	}
}

// ── Openings & applications ──────────────────────────────────────────

func ListOpeningsHandler(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		openings, err := svc.ListOpenings(r.Context())
		if err != nil {
			logger.Ctx(r.Context()).Error("voxhire list openings failed", logger.Error(err))
			writeApplyError(w, http.StatusInternalServerError, "Failed to list openings")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"openings": openings})
	}
}

func ListLeadsHandler(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		leads, err := svc.ListCompanyLeads(r.Context())
		if err != nil {
			logger.Ctx(r.Context()).Error("voxhire list leads failed", logger.Error(err))
			writeApplyError(w, http.StatusInternalServerError, "Failed to list leads")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"leads": leads})
	}
}

func ListApplicationsHandler(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil || jobID <= 0 {
			writeApplyError(w, http.StatusBadRequest, "Invalid opening id")
			return
		}

		q := r.URL.Query()
		filter := domain.ApplicationFilter{
			JobID:    jobID,
			Sort:     strings.TrimSpace(q.Get("sort")),
			Location: strings.TrimSpace(q.Get("location")),
			Query:    strings.TrimSpace(q.Get("q")),
			Limit:    parseInt32(q.Get("limit")),
			Offset:   parseInt32(q.Get("offset")),
		}

		apps, total, err := svc.ListApplications(r.Context(), filter)
		if err != nil {
			logger.Ctx(r.Context()).Error("voxhire list applications failed", logger.Error(err))
			writeApplyError(w, http.StatusInternalServerError, "Failed to list applications")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"applications": apps,
			"total":        total,
			"limit":        filter.Limit,
			"offset":       filter.Offset,
		})
	}
}

func UpdateApplicationStatusHandler(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil || id <= 0 {
			writeApplyError(w, http.StatusBadRequest, "Invalid application id")
			return
		}

		var body struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeApplyError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		if err := svc.UpdateApplicationStatus(r.Context(), id, body.Status); err != nil {
			switch {
			case errors.Is(err, domain.ErrApplicationNotFound):
				writeApplyError(w, http.StatusNotFound, "Application not found")
			case errors.Is(err, domain.ErrInvalidInput):
				writeApplyError(w, http.StatusBadRequest, "Status is required")
			default:
				logger.Ctx(r.Context()).Error("voxhire update application status failed", logger.Error(err))
				writeApplyError(w, http.StatusInternalServerError, "Failed to update status")
			}
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"status": "success"})
	}
}

func ApplicationCVHandler(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil || id <= 0 {
			writeApplyError(w, http.StatusBadRequest, "Invalid application id")
			return
		}

		url, err := svc.GetApplicationCVURL(r.Context(), id)
		if err != nil {
			if errors.Is(err, domain.ErrApplicationNotFound) {
				writeApplyError(w, http.StatusNotFound, "No CV on file for this application")
				return
			}
			logger.Ctx(r.Context()).Error("voxhire get cv url failed", logger.Error(err))
			writeApplyError(w, http.StatusInternalServerError, "Failed to load CV")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"url": url})
	}
}

// ── helpers ──────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logger.Global().Warn("voxhire: failed to encode JSON response", logger.Error(err))
	}
}

func parseInt32(s string) int32 {
	if s == "" {
		return 0
	}
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0
	}
	return int32(n)
}
