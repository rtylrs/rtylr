package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/recruit/careers/internal/application"
	"github.com/voxire/onion/internal/modules/recruit/careers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/money"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// ─────────────────────────────────────────────
// Route declarations
// ─────────────────────────────────────────────

func ListRecruitCareersRoutes() []string  { return []string{"/recruit/careers"} }
func CreateRecruitCareerRoutes() []string { return []string{"/recruit/careers"} }
func UpdateRecruitCareerRoutes() []string { return []string{"/recruit/careers"} }
func DeleteRecruitCareerRoutes() []string { return []string{"/recruit/careers"} }

// ─────────────────────────────────────────────
// Shared payload types
// ─────────────────────────────────────────────

type RecruitCareerPayload struct {
	ID             int64   `json:"id"`
	OwnerID        int64   `json:"owner_id"`
	Slug           string  `json:"slug"`
	Title          string  `json:"title"`
	Department     string  `json:"department"`
	Location       string  `json:"location"`
	EmploymentType string  `json:"employment_type"`
	WorkMode       string  `json:"work_mode"`
	Seniority      string  `json:"seniority"`
	Details        string  `json:"details"`
	Requirements   string  `json:"requirements"`
	SalaryMin      float64 `json:"salary_min"`
	SalaryMax      float64 `json:"salary_max"`
	SalaryCurrency string  `json:"salary_currency"`
	IsPublished    bool    `json:"is_published"`
	IsActive       bool    `json:"is_active"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

func toRecruitCareerPayload(c domain.RecruitCareer) *RecruitCareerPayload {
	return &RecruitCareerPayload{
		ID:             c.ID,
		OwnerID:        c.OwnerID,
		Slug:           c.Slug,
		Title:          c.Title,
		Department:     c.Department,
		Location:       c.Location,
		EmploymentType: c.EmploymentType,
		WorkMode:       c.WorkMode,
		Seniority:      c.Seniority,
		Details:        c.Details,
		Requirements:   c.Requirements,
		SalaryMin:      c.SalaryMin.Float64(),
		SalaryMax:      c.SalaryMax.Float64(),
		SalaryCurrency: c.SalaryCurrency,
		IsPublished:    c.IsPublished,
		IsActive:       c.IsActive,
		CreatedAt:      c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      c.UpdatedAt.Format(time.RFC3339),
	}
}

// ─────────────────────────────────────────────
// GET /recruit/careers — list caller's jobs
// ─────────────────────────────────────────────

type ListRecruitCareersHandler struct {
	CareersService *application.Service
}

type ListRecruitCareersRequest struct {
	// Admins see draft + published; public pages pass ?published_only=1.
	PublishedOnly string `json:"published_only"`
}

type ListRecruitCareersResponse struct {
	Careers []*RecruitCareerPayload `json:"careers"`
}

func (h *ListRecruitCareersHandler) Serve(ctx context.Context, req *ListRecruitCareersRequest) (ListRecruitCareersResponse, error) {
	authUser, err := requireRecruitPermission(ctx, "recruit:jobs:view")
	if err != nil {
		return ListRecruitCareersResponse{}, err
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return ListRecruitCareersResponse{}, onekiterrors.PermissionDenied()
	}

	careers, err := h.CareersService.ListRecruitCareers(ctx, businessID, req.PublishedOnly == "1" || req.PublishedOnly == "true")
	if err != nil {
		return ListRecruitCareersResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list careers")
	}

	payloads := make([]*RecruitCareerPayload, 0, len(careers))
	for _, c := range careers {
		payloads = append(payloads, toRecruitCareerPayload(c))
	}
	return ListRecruitCareersResponse{Careers: payloads}, nil
}

// ─────────────────────────────────────────────
// POST /recruit/careers
// ─────────────────────────────────────────────

type CreateRecruitCareerHandler struct {
	CareersService *application.Service
}

type CreateRecruitCareerRequest struct {
	Slug           string  `json:"slug"`
	Title          string  `json:"title"`
	Department     string  `json:"department"`
	Location       string  `json:"location"`
	EmploymentType string  `json:"employment_type"`
	WorkMode       string  `json:"work_mode"`
	Seniority      string  `json:"seniority"`
	Details        string  `json:"details"`
	Requirements   string  `json:"requirements"`
	SalaryMin      float64 `json:"salary_min"`
	SalaryMax      float64 `json:"salary_max"`
	SalaryCurrency string  `json:"salary_currency"`
	IsPublished    bool    `json:"is_published"`
}

type CreateRecruitCareerResponse struct {
	Career *RecruitCareerPayload `json:"career"`
}

func (h *CreateRecruitCareerHandler) Validate(_ context.Context, req *CreateRecruitCareerRequest) error {
	if req.Title == "" {
		return onekiterrors.New(http.StatusBadRequest, "Title is required")
	}
	if req.Details == "" {
		return onekiterrors.New(http.StatusBadRequest, "Details are required")
	}
	return nil
}

func (h *CreateRecruitCareerHandler) Serve(ctx context.Context, req *CreateRecruitCareerRequest) (CreateRecruitCareerResponse, error) {
	authUser, err := requireRecruitPermission(ctx, "recruit:jobs:manage")
	if err != nil {
		return CreateRecruitCareerResponse{}, err
	}
	if req.IsPublished && !authPkg.HasAppPermission(authUser, authorization.AppRecruit, "recruit:jobs:publish") {
		return CreateRecruitCareerResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return CreateRecruitCareerResponse{}, onekiterrors.PermissionDenied()
	}

	salaryMin, salaryMax, salaryErr := recruitSalaryAmounts(req.SalaryMin, req.SalaryMax)
	if salaryErr != nil {
		return CreateRecruitCareerResponse{}, salaryErr
	}

	input := domain.RecruitCareerInput{
		OwnerID:        authUser.UserID,
		BusinessID:     businessID,
		Slug:           req.Slug,
		Title:          req.Title,
		Department:     req.Department,
		Location:       req.Location,
		EmploymentType: req.EmploymentType,
		WorkMode:       req.WorkMode,
		Seniority:      req.Seniority,
		Details:        req.Details,
		Requirements:   req.Requirements,
		SalaryMin:      salaryMin,
		SalaryMax:      salaryMax,
		SalaryCurrency: req.SalaryCurrency,
		IsPublished:    req.IsPublished,
	}

	// Round-trip the freshly-created row (resolved slug, timestamps, stored
	// published state) so the client doesn't need a second request. The
	// service reads the row back by id because the slug may have been
	// generated server-side from the title.
	created, err := h.CareersService.CreateCareer(ctx, input)
	if err != nil {
		return CreateRecruitCareerResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create career")
	}
	invalidatePublicCareersCache(ctx, businessID)
	return CreateRecruitCareerResponse{Career: toRecruitCareerPayload(created)}, nil
}

// ─────────────────────────────────────────────
// PUT /recruit/careers
// ─────────────────────────────────────────────

type UpdateRecruitCareerHandler struct {
	CareersService *application.Service
}

type UpdateRecruitCareerRequest struct {
	ID             int64   `json:"id"`
	Slug           string  `json:"slug"`
	Title          string  `json:"title"`
	Department     string  `json:"department"`
	Location       string  `json:"location"`
	EmploymentType string  `json:"employment_type"`
	WorkMode       string  `json:"work_mode"`
	Seniority      string  `json:"seniority"`
	Details        string  `json:"details"`
	Requirements   string  `json:"requirements"`
	SalaryMin      float64 `json:"salary_min"`
	SalaryMax      float64 `json:"salary_max"`
	SalaryCurrency string  `json:"salary_currency"`
	IsPublished    bool    `json:"is_published"`
}

type UpdateRecruitCareerResponse struct {
	Message string `json:"message"`
}

func (h *UpdateRecruitCareerHandler) Validate(_ context.Context, req *UpdateRecruitCareerRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "id is required")
	}
	if req.Title == "" {
		return onekiterrors.New(http.StatusBadRequest, "Title is required")
	}
	return nil
}

func (h *UpdateRecruitCareerHandler) Serve(ctx context.Context, req *UpdateRecruitCareerRequest) (UpdateRecruitCareerResponse, error) {
	authUser, err := requireRecruitPermission(ctx, "recruit:jobs:manage")
	if err != nil {
		return UpdateRecruitCareerResponse{}, err
	}
	if req.IsPublished && !authPkg.HasAppPermission(authUser, authorization.AppRecruit, "recruit:jobs:publish") {
		return UpdateRecruitCareerResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return UpdateRecruitCareerResponse{}, onekiterrors.PermissionDenied()
	}

	salaryMin, salaryMax, salaryErr := recruitSalaryAmounts(req.SalaryMin, req.SalaryMax)
	if salaryErr != nil {
		return UpdateRecruitCareerResponse{}, salaryErr
	}

	input := domain.RecruitCareerInput{
		ID:             req.ID,
		OwnerID:        authUser.UserID,
		BusinessID:     businessID,
		Slug:           req.Slug,
		Title:          req.Title,
		Department:     req.Department,
		Location:       req.Location,
		EmploymentType: req.EmploymentType,
		WorkMode:       req.WorkMode,
		Seniority:      req.Seniority,
		Details:        req.Details,
		Requirements:   req.Requirements,
		SalaryMin:      salaryMin,
		SalaryMax:      salaryMax,
		SalaryCurrency: req.SalaryCurrency,
		IsPublished:    req.IsPublished,
	}

	if err := h.CareersService.UpdateCareer(ctx, input); err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return UpdateRecruitCareerResponse{}, onekiterrors.New(http.StatusNotFound, "Career not found")
		}
		return UpdateRecruitCareerResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update career")
	}
	invalidatePublicCareersCache(ctx, businessID)
	return UpdateRecruitCareerResponse{Message: "Career updated"}, nil
}

// ─────────────────────────────────────────────
// DELETE /recruit/careers
// ─────────────────────────────────────────────

type DeleteRecruitCareerHandler struct {
	CareersService *application.Service
}

type DeleteRecruitCareerRequest struct {
	ID int64 `json:"id"`
}

type DeleteRecruitCareerResponse struct {
	Message string `json:"message"`
}

func (h *DeleteRecruitCareerHandler) Validate(_ context.Context, req *DeleteRecruitCareerRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "id is required")
	}
	return nil
}

func (h *DeleteRecruitCareerHandler) Serve(ctx context.Context, req *DeleteRecruitCareerRequest) (DeleteRecruitCareerResponse, error) {
	authUser, err := requireRecruitPermission(ctx, "recruit:jobs:manage")
	if err != nil {
		return DeleteRecruitCareerResponse{}, err
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return DeleteRecruitCareerResponse{}, onekiterrors.PermissionDenied()
	}

	if err := h.CareersService.DeleteCareer(ctx, businessID, req.ID); err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return DeleteRecruitCareerResponse{}, onekiterrors.New(http.StatusNotFound, "Career not found")
		}
		return DeleteRecruitCareerResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to archive career")
	}
	invalidatePublicCareersCache(ctx, businessID)
	return DeleteRecruitCareerResponse{Message: "Career archived"}, nil
}

func recruitSalaryAmounts(minSalary, maxSalary float64) (money.Amount, money.Amount, error) {
	salaryMin, err := money.FromFloat64(minSalary)
	if err != nil {
		return 0, 0, onekiterrors.New(http.StatusBadRequest, "salary_min must be a valid monetary amount")
	}
	salaryMax, err := money.FromFloat64(maxSalary)
	if err != nil {
		return 0, 0, onekiterrors.New(http.StatusBadRequest, "salary_max must be a valid monetary amount")
	}
	if minSalary < 0 || maxSalary < 0 {
		return 0, 0, onekiterrors.New(http.StatusBadRequest, "salary range cannot be negative")
	}
	if maxSalary > 0 && minSalary > maxSalary {
		return 0, 0, onekiterrors.New(http.StatusBadRequest, "salary_min cannot exceed salary_max")
	}
	return salaryMin, salaryMax, nil
}
