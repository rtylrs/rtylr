package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/application"
	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

type adminCtxKey string

const adminIDCtxKey adminCtxKey = "voxhire_admin_id"

// Route patterns — also returned to GetAnonymousRoutes so the global JWT
// middleware skips them (the admin routes self-guard via the admin JWT).
func AdminLoginRoute() string        { return "/voxhire/admin/login" }
func AdminAdminsRoute() string       { return "/voxhire/admin/admins" }
func AdminMeRoute() string           { return "/voxhire/admin/me" }
func AdminAnalyticsRoute() string    { return "/voxhire/admin/analytics" }
func AdminOpeningsRoute() string     { return "/voxhire/admin/openings" }
func AdminCareersRoute() string      { return "/voxhire/admin/careers" }
func AdminCareerRoute() string       { return "/voxhire/admin/careers/{id}" }
func AdminCareerActiveRoute() string { return "/voxhire/admin/careers/{id}/active" }
func AdminLeadsRoute() string        { return "/voxhire/admin/leads" }
func AdminApplicationsRoute() string { return "/voxhire/admin/openings/{id}/applications" }
func AdminAppStatusRoute() string    { return "/voxhire/admin/applications/{id}/status" }
func AdminAppReviewRoute() string    { return "/voxhire/admin/applications/{id}/review" }
func AdminAppCVRoute() string        { return "/voxhire/admin/applications/{id}/cv" }
func AdminAppTimelineRoute() string  { return "/voxhire/admin/applications/{id}/timeline" }
func AdminCandidatesRoute() string   { return "/voxhire/admin/candidates" }
func AdminCVJobsRoute() string       { return "/voxhire/admin/cv-jobs" }
func AdminCVJobRetryRoute() string   { return "/voxhire/admin/cv-jobs/{source}/{id}/retry" }
func AdminTasksRoute() string        { return "/voxhire/admin/tasks" }
func AdminTaskRoute() string         { return "/voxhire/admin/tasks/{id}" }
func AdminShortlistsRoute() string   { return "/voxhire/admin/shortlists" }
func AdminShortlistItemsRoute() string {
	return "/voxhire/admin/shortlists/{id}/items"
}
func CareerApplyRoute() string { return "/voxhire/careers/apply" }

// RecruitAnonymousRoutes lists every recruit/admin path that bypasses
// the platform JWT middleware.
func RecruitAnonymousRoutes() []string {
	return []string{
		AdminLoginRoute(),
		AdminAdminsRoute(),
		AdminMeRoute(),
		AdminAnalyticsRoute(),
		AdminOpeningsRoute(),
		AdminCareersRoute(),
		AdminCareerRoute(),
		AdminCareerActiveRoute(),
		AdminLeadsRoute(),
		AdminApplicationsRoute(),
		AdminAppStatusRoute(),
		AdminAppReviewRoute(),
		AdminAppCVRoute(),
		AdminAppTimelineRoute(),
		AdminCandidatesRoute(),
		AdminCVJobsRoute(),
		AdminCVJobRetryRoute(),
		AdminTasksRoute(),
		AdminTaskRoute(),
		AdminShortlistsRoute(),
		AdminShortlistItemsRoute(),
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
			ctx = domain.WithOwnerID(ctx, claims.AdminID)
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

// ── Admin accounts ───────────────────────────────────────────────────

func AdminMeHandler(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		admin, err := svc.GetAdmin(r.Context(), adminIDFromCtx(r.Context()))
		if err != nil {
			writeApplyError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"admin": adminViewPointer(admin)})
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
		writeJSON(w, http.StatusOK, map[string]any{"admins": mapViews(admins, adminViewOf)})
	}
}

// ── Openings & applications ──────────────────────────────────────────

func ListOpeningsHandler(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		page, err := svc.ListOpeningsPage(r.Context(), parseInt32(r.Context(), "limit", q.Get("limit")), parseInt32(r.Context(), "offset", q.Get("offset")))
		if err != nil {
			logger.Ctx(r.Context()).Error("voxhire list openings failed", logger.Error(err))
			writeApplyError(w, http.StatusInternalServerError, "Failed to list openings")
			return
		}
		writeJSON(w, http.StatusOK, openingPageViewOf(page))
	}
}

func AnalyticsHandler(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		analytics, err := svc.GetAnalytics(r.Context())
		if err != nil {
			logger.Ctx(r.Context()).Error("voxhire analytics failed", logger.Error(err))
			writeApplyError(w, http.StatusInternalServerError, "Failed to load analytics")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"analytics": analyticsViewPointer(analytics)})
	}
}

type adminCareerRequest struct {
	Title          string  `json:"title"`
	Department     string  `json:"department"`
	Location       string  `json:"location"`
	EmploymentType string  `json:"employment_type"`
	WorkMode       string  `json:"work_mode"`
	Seniority      string  `json:"seniority"`
	Details        string  `json:"details"`
	SalaryMin      float64 `json:"salary_min"`
	SalaryMax      float64 `json:"salary_max"`
	SalaryCurrency string  `json:"salary_currency"`
}

func (b adminCareerRequest) careerInput(id int64) domain.CareerInput {
	return domain.CareerInput{
		ID:             id,
		Title:          b.Title,
		Department:     b.Department,
		Location:       b.Location,
		EmploymentType: b.EmploymentType,
		WorkMode:       b.WorkMode,
		Seniority:      b.Seniority,
		Details:        b.Details,
		SalaryMin:      b.SalaryMin,
		SalaryMax:      b.SalaryMax,
		SalaryCurrency: b.SalaryCurrency,
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
		writeJSON(w, http.StatusOK, map[string]any{"leads": mapViews(leads, leadViewOf)})
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
			Limit:    parseInt32(r.Context(), "limit", q.Get("limit")),
			Offset:   parseInt32(r.Context(), "offset", q.Get("offset")),
		}

		apps, total, err := svc.ListApplications(r.Context(), filter)
		if err != nil {
			logger.Ctx(r.Context()).Error("voxhire list applications failed", logger.Error(err))
			writeApplyError(w, http.StatusInternalServerError, "Failed to list applications")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"applications": mapViews(apps, applicationViewOf),
			"total":        total,
			"limit":        filter.Limit,
			"offset":       filter.Offset,
		})
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

func ApplicationTimelineHandler(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil || id <= 0 {
			writeApplyError(w, http.StatusBadRequest, "Invalid application id")
			return
		}
		events, err := svc.ListApplicationTimeline(r.Context(), id)
		if err != nil {
			if errors.Is(err, domain.ErrApplicationNotFound) {
				writeApplyError(w, http.StatusNotFound, "Application not found")
				return
			}
			logger.Ctx(r.Context()).Error("voxhire timeline failed", logger.Error(err))
			writeApplyError(w, http.StatusInternalServerError, "Failed to load timeline")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"events": mapViews(events, timelineEventViewOf)})
	}
}

func ListCandidatesHandler(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		filter := domain.CandidateFilter{
			Query:    strings.TrimSpace(q.Get("q")),
			Source:   strings.TrimSpace(q.Get("source")),
			CVStatus: strings.TrimSpace(q.Get("cv_status")),
			Limit:    parseInt32(r.Context(), "limit", q.Get("limit")),
			Offset:   parseInt32(r.Context(), "offset", q.Get("offset")),
		}
		candidates, total, err := svc.ListCandidates(r.Context(), filter)
		if err != nil {
			logger.Ctx(r.Context()).Error("voxhire list candidates failed", logger.Error(err))
			writeApplyError(w, http.StatusInternalServerError, "Failed to list candidates")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"candidates": mapViews(candidates, candidateViewOf),
			"total":      total,
			"limit":      filter.Limit,
			"offset":     filter.Offset,
		})
	}
}

func ListCVJobsHandler(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		filter := domain.CandidateFilter{
			Query:    strings.TrimSpace(q.Get("q")),
			Source:   strings.TrimSpace(q.Get("source")),
			CVStatus: strings.TrimSpace(q.Get("cv_status")),
			Limit:    parseInt32(r.Context(), "limit", q.Get("limit")),
		}
		jobs, err := svc.ListCVJobs(r.Context(), filter)
		if err != nil {
			logger.Ctx(r.Context()).Error("voxhire list cv jobs failed", logger.Error(err))
			writeApplyError(w, http.StatusInternalServerError, "Failed to list CV jobs")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"jobs": mapViews(jobs, cVJobViewOf)})
	}
}

type adminTaskRequest struct {
	ApplicationID int64           `json:"application_id"`
	Title         string          `json:"title"`
	Status        string          `json:"status"`
	DueAt         json.RawMessage `json:"due_at"`
}

// dueAtInput reports the due date the request carries and whether it was
// supplied at all. The raw JSON is kept because a decoded string cannot tell
// "due_at was omitted" (keep the stored date) apart from "due_at was sent empty
// or null" (clear the stored date).
func (b adminTaskRequest) dueAtInput() (*time.Time, bool, error) {
	raw := strings.TrimSpace(string(b.DueAt))
	if raw == "" {
		return nil, false, nil
	}
	if raw == "null" {
		return nil, true, nil
	}
	var value string
	if err := json.Unmarshal(b.DueAt, &value); err != nil {
		return nil, true, err
	}
	dueAt, err := parseAdminTime(value)
	if err != nil {
		return nil, true, err
	}
	return dueAt, true, nil
}

func (b adminTaskRequest) taskInput(id int64) (domain.TaskInput, error) {
	dueAt, dueAtSet, err := b.dueAtInput()
	if err != nil {
		return domain.TaskInput{}, err
	}
	return domain.TaskInput{
		ID:            id,
		ApplicationID: b.ApplicationID,
		Title:         b.Title,
		Status:        b.Status,
		DueAt:         dueAt,
		DueAtSet:      dueAtSet,
	}, nil
}

func ListTasksHandler(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tasks, err := svc.ListTasks(r.Context(), strings.TrimSpace(r.URL.Query().Get("status")))
		if err != nil {
			logger.Ctx(r.Context()).Error("voxhire list tasks failed", logger.Error(err))
			writeApplyError(w, http.StatusInternalServerError, "Failed to list tasks")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"tasks": mapViews(tasks, taskViewOf)})
	}
}

type adminShortlistRequest struct {
	JobID  int64  `json:"job_id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Notes  string `json:"notes"`
}

func ListShortlistsHandler(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var jobID int64
		if raw := r.URL.Query().Get("job_id"); raw != "" {
			parsed, err := strconv.ParseInt(raw, 10, 64)
			if err != nil || parsed <= 0 {
				writeApplyError(w, http.StatusBadRequest, "Invalid job id")
				return
			}
			jobID = parsed
		}
		shortlists, err := svc.ListShortlists(r.Context(), jobID)
		if err != nil {
			logger.Ctx(r.Context()).Error("voxhire list shortlists failed", logger.Error(err))
			writeApplyError(w, http.StatusInternalServerError, "Failed to list shortlists")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"shortlists": mapViews(shortlists, shortlistViewOf)})
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

func parseInt32(ctx context.Context, name, s string) int32 {
	if s == "" {
		return 0
	}
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		logger.Ctx(ctx).Warn("invalid voxhire query parameter",
			logger.String("parameter", name), logger.String("value", s), logger.Error(err))
		return 0
	}
	return int32(n)
}

func parseAdminTime(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return &t, nil
	}
	if t, err := time.Parse("2006-01-02", raw); err == nil {
		return &t, nil
	}
	return nil, errors.New("invalid time")
}
