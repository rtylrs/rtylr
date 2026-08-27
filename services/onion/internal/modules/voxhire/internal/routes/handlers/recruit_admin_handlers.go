package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

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
		q := r.URL.Query()
		page, err := svc.ListOpeningsPage(r.Context(), parseInt32(r.Context(), "limit", q.Get("limit")), parseInt32(r.Context(), "offset", q.Get("offset")))
		if err != nil {
			logger.Ctx(r.Context()).Error("voxhire list openings failed", logger.Error(err))
			writeApplyError(w, http.StatusInternalServerError, "Failed to list openings")
			return
		}
		writeJSON(w, http.StatusOK, page)
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
		writeJSON(w, http.StatusOK, map[string]any{"analytics": analytics})
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

func CreateCareerHandler(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body adminCareerRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeApplyError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		career, err := svc.CreateCareer(r.Context(), body.careerInput(0))
		if err != nil {
			writeCareerMutationError(w, r, err, "create")
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"career": career})
	}
}

func UpdateCareerHandler(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil || id <= 0 {
			writeApplyError(w, http.StatusBadRequest, "Invalid career id")
			return
		}

		var body adminCareerRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeApplyError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		career, err := svc.UpdateCareer(r.Context(), body.careerInput(id))
		if err != nil {
			writeCareerMutationError(w, r, err, "update")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"career": career})
	}
}

func SetCareerActiveHandler(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil || id <= 0 {
			writeApplyError(w, http.StatusBadRequest, "Invalid career id")
			return
		}

		var body struct {
			Active bool `json:"active"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeApplyError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		career, err := svc.SetCareerActive(r.Context(), id, body.Active)
		if err != nil {
			writeCareerMutationError(w, r, err, "update active state for")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"career": career})
	}
}

func writeCareerMutationError(w http.ResponseWriter, r *http.Request, err error, action string) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		writeApplyError(w, http.StatusBadRequest, "Title is required")
	case errors.Is(err, domain.ErrCareerNotFound):
		writeApplyError(w, http.StatusNotFound, "Career not found")
	default:
		logger.Ctx(r.Context()).Error("voxhire career mutation failed",
			logger.String("action", action),
			logger.Error(err),
		)
		writeApplyError(w, http.StatusInternalServerError, "Failed to "+action+" career")
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

func UpdateApplicationReviewHandler(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil || id <= 0 {
			writeApplyError(w, http.StatusBadRequest, "Invalid application id")
			return
		}

		var body struct {
			Status       string `json:"status"`
			ReviewRating int    `json:"review_rating"`
			ReviewNotes  string `json:"review_notes"`
			NextStep     string `json:"next_step"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeApplyError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		err = svc.UpdateApplicationReview(r.Context(), id, domain.ApplicationReviewInput{
			Status:       body.Status,
			ReviewRating: body.ReviewRating,
			ReviewNotes:  body.ReviewNotes,
			NextStep:     body.NextStep,
		})
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrApplicationNotFound):
				writeApplyError(w, http.StatusNotFound, "Application not found")
			case errors.Is(err, domain.ErrInvalidInput):
				writeApplyError(w, http.StatusBadRequest, "Invalid review")
			default:
				logger.Ctx(r.Context()).Error("voxhire update application review failed", logger.Error(err))
				writeApplyError(w, http.StatusInternalServerError, "Failed to update review")
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
		writeJSON(w, http.StatusOK, map[string]any{"events": events})
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
			"candidates": candidates,
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
		writeJSON(w, http.StatusOK, map[string]any{"jobs": jobs})
	}
}

func RetryCVJobHandler(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil || id <= 0 {
			writeApplyError(w, http.StatusBadRequest, "Invalid CV job id")
			return
		}
		source := strings.TrimSpace(chi.URLParam(r, "source"))
		if err := svc.RetryCVProcessing(r.Context(), source, id); err != nil {
			switch {
			case errors.Is(err, domain.ErrInvalidInput):
				writeApplyError(w, http.StatusBadRequest, "Invalid CV job")
			case errors.Is(err, domain.ErrApplicationNotFound):
				writeApplyError(w, http.StatusNotFound, "CV job not found")
			default:
				logger.Ctx(r.Context()).Error("voxhire retry cv job failed", logger.Error(err))
				writeApplyError(w, http.StatusInternalServerError, "Failed to retry CV job")
			}
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"status": "success"})
	}
}

type adminTaskRequest struct {
	ApplicationID int64  `json:"application_id"`
	Title         string `json:"title"`
	Status        string `json:"status"`
	DueAt         string `json:"due_at"`
}

func (b adminTaskRequest) taskInput(id int64) (domain.TaskInput, error) {
	dueAt, err := parseAdminTime(b.DueAt)
	if err != nil {
		return domain.TaskInput{}, err
	}
	return domain.TaskInput{
		ID:            id,
		ApplicationID: b.ApplicationID,
		Title:         b.Title,
		Status:        b.Status,
		DueAt:         dueAt,
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
		writeJSON(w, http.StatusOK, map[string]any{"tasks": tasks})
	}
}

func CreateTaskHandler(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body adminTaskRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeApplyError(w, http.StatusBadRequest, "Invalid request")
			return
		}
		in, err := body.taskInput(0)
		if err != nil {
			writeApplyError(w, http.StatusBadRequest, "Invalid due date")
			return
		}
		task, err := svc.UpsertTask(r.Context(), in)
		if err != nil {
			writeTaskMutationError(w, r, err, "create")
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"task": task})
	}
}

func UpdateTaskHandler(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil || id <= 0 {
			writeApplyError(w, http.StatusBadRequest, "Invalid task id")
			return
		}
		var body adminTaskRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeApplyError(w, http.StatusBadRequest, "Invalid request")
			return
		}
		in, err := body.taskInput(id)
		if err != nil {
			writeApplyError(w, http.StatusBadRequest, "Invalid due date")
			return
		}
		task, err := svc.UpdateTask(r.Context(), in)
		if err != nil {
			writeTaskMutationError(w, r, err, "update")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"task": task})
	}
}

func writeTaskMutationError(w http.ResponseWriter, r *http.Request, err error, action string) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		writeApplyError(w, http.StatusBadRequest, "Application and task title are required")
	case errors.Is(err, domain.ErrApplicationNotFound):
		writeApplyError(w, http.StatusNotFound, "Task or application not found")
	default:
		logger.Ctx(r.Context()).Error("voxhire task mutation failed",
			logger.String("action", action),
			logger.Error(err),
		)
		writeApplyError(w, http.StatusInternalServerError, "Failed to "+action+" task")
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
		writeJSON(w, http.StatusOK, map[string]any{"shortlists": shortlists})
	}
}

func CreateShortlistHandler(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body adminShortlistRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeApplyError(w, http.StatusBadRequest, "Invalid request")
			return
		}
		shortlist, err := svc.CreateShortlist(r.Context(), domain.ShortlistInput{
			JobID:  body.JobID,
			Name:   body.Name,
			Status: body.Status,
			Notes:  body.Notes,
		})
		if err != nil {
			if errors.Is(err, domain.ErrInvalidInput) {
				writeApplyError(w, http.StatusBadRequest, "Job and shortlist name are required")
				return
			}
			logger.Ctx(r.Context()).Error("voxhire create shortlist failed", logger.Error(err))
			writeApplyError(w, http.StatusInternalServerError, "Failed to create shortlist")
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"shortlist": shortlist})
	}
}

func AddShortlistItemHandler(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortlistID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil || shortlistID <= 0 {
			writeApplyError(w, http.StatusBadRequest, "Invalid shortlist id")
			return
		}
		var body struct {
			ApplicationID int64  `json:"application_id"`
			Rank          int    `json:"rank"`
			Notes         string `json:"notes"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeApplyError(w, http.StatusBadRequest, "Invalid request")
			return
		}
		item, err := svc.AddShortlistItem(r.Context(), domain.ShortlistItemInput{
			ShortlistID:   shortlistID,
			ApplicationID: body.ApplicationID,
			Rank:          body.Rank,
			Notes:         body.Notes,
		})
		if err != nil {
			if errors.Is(err, domain.ErrInvalidInput) {
				writeApplyError(w, http.StatusBadRequest, "Shortlist and application are required")
				return
			}
			if errors.Is(err, domain.ErrApplicationNotFound) {
				writeApplyError(w, http.StatusNotFound, "Shortlist item not found")
				return
			}
			logger.Ctx(r.Context()).Error("voxhire add shortlist item failed", logger.Error(err))
			writeApplyError(w, http.StatusInternalServerError, "Failed to add shortlist item")
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"item": item})
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
