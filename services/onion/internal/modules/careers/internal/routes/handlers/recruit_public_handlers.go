package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/careers/internal/application"
	"github.com/voxire/onion/internal/modules/careers/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// publicCareersCacheScope namespaces the per-owner public careers page
// cache. The careers, careers_customizations and careers_sections tables
// are written only within this module, so a short TTL plus invalidation
// on every careers/customization/section mutation keeps the anonymous
// public page consistent. The cache is keyed by owner_id (resolved from
// the slug), so write handlers — which know the owner_id — can wipe it.
const publicCareersCacheScope = "careers_public"

// invalidatePublicCareersCache wipes the cached public careers page for an
// owner. Call from every careers/customization/section create/update/delete
// so the public site reflects changes within the TTL window at worst.
func invalidatePublicCareersCache(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(publicCareersCacheScope, ownerID))
	}
}

// The "public" handlers power the candidate-facing site at
// recruit.rtylr.com. They are declared anonymous in the module's
// GetAnonymousRoutes so the auth middleware lets them through without
// a JWT. Every one takes the business slug as a query param; the
// handler resolves that to an owner_id before touching tenant data,
// so candidates never see or need to know a numeric tenant id.

// ─────────────────────────────────────────────
// Route declarations
// ─────────────────────────────────────────────

func GetPublicBrandingRoutes() []string { return []string{"/recruit/public/branding"} }
func ListPublicCareersRoutes() []string { return []string{"/recruit/public/careers"} }
func GetPublicCareerRoutes() []string   { return []string{"/recruit/public/career"} }
func SubmitApplicationRoutes() []string { return []string{"/recruit/public/apply"} }

// ─────────────────────────────────────────────
// GET /recruit/public/branding?slug=<slug>
// ─────────────────────────────────────────────

type GetPublicBrandingHandler struct {
	CareersService *application.Service
}

type GetPublicBrandingRequest struct {
	Slug string `json:"slug"`
}

type GetPublicBrandingResponse struct {
	Customization *CustomizationPayload `json:"customization"`
}

func (h *GetPublicBrandingHandler) Validate(_ context.Context, req *GetPublicBrandingRequest) *errorModels.ServerError {
	if req.Slug == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "slug is required"}
	}
	return nil
}

func (h *GetPublicBrandingHandler) Serve(ctx context.Context, req *GetPublicBrandingRequest) (GetPublicBrandingResponse, *errorModels.ServerError) {
	c, err := h.CareersService.GetCustomizationBySlug(ctx, req.Slug)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return GetPublicBrandingResponse{}, &errorModels.ServerError{
				Code:    http.StatusNotFound,
				Message: "Careers page not found",
			}
		}
		return GetPublicBrandingResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to load branding",
		}
	}
	return GetPublicBrandingResponse{Customization: toCustomizationPayload(c)}, nil
}

// ─────────────────────────────────────────────
// GET /recruit/public/careers?slug=<slug>
// ─────────────────────────────────────────────

type ListPublicCareersHandler struct {
	CareersService *application.Service
}

type ListPublicCareersRequest struct {
	Slug string `json:"slug"`
}

type ListPublicCareersResponse struct {
	Careers       []*RecruitCareerPayload `json:"careers"`
	Customization *CustomizationPayload   `json:"customization"`
	// Sections are the content blocks (perks, culture, etc.) a
	// business has enabled. Returned inline so the public page render
	// is one request, not three.
	Sections []*SectionPayload `json:"sections"`
}

func (h *ListPublicCareersHandler) Validate(_ context.Context, req *ListPublicCareersRequest) *errorModels.ServerError {
	if req.Slug == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "slug is required"}
	}
	return nil
}

func (h *ListPublicCareersHandler) Serve(ctx context.Context, req *ListPublicCareersRequest) (ListPublicCareersResponse, *errorModels.ServerError) {
	custom, err := h.CareersService.GetCustomizationBySlug(ctx, req.Slug)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return ListPublicCareersResponse{}, &errorModels.ServerError{
				Code:    http.StatusNotFound,
				Message: "Careers page not found",
			}
		}
		return ListPublicCareersResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to load careers",
		}
	}

	res, cacheErr := cacher.Of[ListPublicCareersResponse]().GetOrComputeSingle(
		ctx,
		cacheKeys.Scoped(publicCareersCacheScope, custom.OwnerID, "list"),
		func(ctx context.Context) (ListPublicCareersResponse, error) {
			return h.computePublicCareers(ctx, custom)
		},
		cacher.WithExpiry(ctx, 3*time.Minute),
	)
	if cacheErr == nil {
		return res, nil
	}

	res, computeErr := h.computePublicCareers(ctx, custom)
	if computeErr != nil {
		return ListPublicCareersResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to load careers",
		}
	}
	return res, nil
}

func (h *ListPublicCareersHandler) computePublicCareers(ctx context.Context, custom domain.Customization) (ListPublicCareersResponse, error) {
	careers, err := h.CareersService.ListRecruitCareers(ctx, custom.BusinessID, true)
	if err != nil {
		return ListPublicCareersResponse{}, err
	}

	payloads := make([]*RecruitCareerPayload, 0, len(careers))
	for _, c := range careers {
		payloads = append(payloads, toRecruitCareerPayload(c))
	}

	// Include enabled sections inline. Errors here are non-fatal —
	// the public page still renders without sections if the fetch
	// fails, which is better than dropping the whole careers list.
	sections, secErr := h.CareersService.ListSections(ctx, custom.BusinessID, true)
	if secErr != nil {
		sections = nil
	}
	sectionPayloads := make([]*SectionPayload, 0, len(sections))
	for _, s := range sections {
		sectionPayloads = append(sectionPayloads, toSectionPayload(s))
	}

	return ListPublicCareersResponse{
		Careers:       payloads,
		Customization: toCustomizationPayload(custom),
		Sections:      sectionPayloads,
	}, nil
}

// ─────────────────────────────────────────────
// GET /recruit/public/career?slug=<slug>&career_slug=<slug>
// ─────────────────────────────────────────────

type GetPublicCareerHandler struct {
	CareersService *application.Service
}

type GetPublicCareerRequest struct {
	Slug       string `json:"slug"`
	CareerSlug string `json:"career_slug"`
}

type GetPublicCareerResponse struct {
	Career        *RecruitCareerPayload `json:"career"`
	Customization *CustomizationPayload `json:"customization"`
	// Fields are the custom form questions the HR admin added on this
	// role. The candidate's browser renders these inline after the
	// default name/email/resume block.
	Fields []*FormFieldPayload `json:"fields"`
}

func (h *GetPublicCareerHandler) Validate(_ context.Context, req *GetPublicCareerRequest) *errorModels.ServerError {
	if req.Slug == "" || req.CareerSlug == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "slug and career_slug are required"}
	}
	return nil
}

func (h *GetPublicCareerHandler) Serve(ctx context.Context, req *GetPublicCareerRequest) (GetPublicCareerResponse, *errorModels.ServerError) {
	custom, err := h.CareersService.GetCustomizationBySlug(ctx, req.Slug)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return GetPublicCareerResponse{}, &errorModels.ServerError{
				Code:    http.StatusNotFound,
				Message: "Careers page not found",
			}
		}
		return GetPublicCareerResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to load career",
		}
	}

	c, err := h.CareersService.GetCareerBySlug(ctx, custom.BusinessID, req.CareerSlug)
	if err != nil || !c.IsPublished {
		return GetPublicCareerResponse{}, &errorModels.ServerError{
			Code:    http.StatusNotFound,
			Message: "Role not found",
		}
	}

	// Pull custom form fields inline — a broken fetch shouldn't take
	// down the role page, so we swallow errors and continue.
	fields, fieldErr := h.CareersService.ListFormFieldsByCareer(ctx, c.ID)
	if fieldErr != nil {
		fields = nil
	}
	fieldPayloads := make([]*FormFieldPayload, 0, len(fields))
	for _, f := range fields {
		fieldPayloads = append(fieldPayloads, toFormFieldPayload(f))
	}

	return GetPublicCareerResponse{
		Career:        toRecruitCareerPayload(c),
		Customization: toCustomizationPayload(custom),
		Fields:        fieldPayloads,
	}, nil
}

// ─────────────────────────────────────────────
// POST /recruit/public/apply
// ─────────────────────────────────────────────

type SubmitApplicationHandler struct {
	CareersService *application.Service
}

type SubmitApplicationRequest struct {
	Slug          string `json:"slug"`
	CareerSlug    string `json:"career_slug"`
	CandidateName string `json:"candidate_name"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	ResumeURL     string `json:"resume_url"`
	CoverLetter   string `json:"cover_letter"`
	// AnswersJSON is a JSON-encoded map of field_id → answer for the
	// custom questions the admin added. The server writes it through
	// unchanged — the HR inbox decodes it on read.
	AnswersJSON string `json:"answers_json"`
}

type SubmitApplicationResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *SubmitApplicationHandler) Validate(_ context.Context, req *SubmitApplicationRequest) *errorModels.ServerError {
	if req.Slug == "" || req.CareerSlug == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "slug and career_slug are required"}
	}
	if req.CandidateName == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}
	if req.Email == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Email is required"}
	}
	return nil
}

func (h *SubmitApplicationHandler) Serve(ctx context.Context, req *SubmitApplicationRequest) (SubmitApplicationResponse, *errorModels.ServerError) {
	// Resolve owner + career via the slug pair. Looking both up here
	// means the candidate never sends a tenant id; a malicious client
	// can only apply to careers that are actually published by the
	// given brand.
	custom, err := h.CareersService.GetCustomizationBySlug(ctx, req.Slug)
	if err != nil {
		return SubmitApplicationResponse{}, &errorModels.ServerError{
			Code:    http.StatusNotFound,
			Message: "Careers page not found",
		}
	}

	career, err := h.CareersService.GetCareerBySlug(ctx, custom.BusinessID, req.CareerSlug)
	if err != nil || !career.IsPublished {
		return SubmitApplicationResponse{}, &errorModels.ServerError{
			Code:    http.StatusNotFound,
			Message: "Role not found or no longer open",
		}
	}

	input := domain.SubmissionInput{
		CareerID:      career.ID,
		OwnerID:       custom.OwnerID,
		BusinessID:    custom.BusinessID,
		CandidateName: req.CandidateName,
		Email:         req.Email,
		Phone:         req.Phone,
		ResumeURL:     req.ResumeURL,
		CoverLetter:   req.CoverLetter,
		AnswersJSON:   req.AnswersJSON,
	}

	id, err := h.CareersService.CreateSubmission(ctx, input)
	if err != nil {
		return SubmitApplicationResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to submit application",
		}
	}

	// Fire-and-forget confirmation + HR notification emails. The
	// service spawns goroutines and swallows errors; the candidate
	// should never see their submission fail because Resend had a
	// bad day.
	submission := domain.Submission{
		ID:            id,
		CareerID:      career.ID,
		OwnerID:       custom.OwnerID,
		BusinessID:    custom.BusinessID,
		CandidateName: req.CandidateName,
		Email:         req.Email,
		Phone:         req.Phone,
		ResumeURL:     req.ResumeURL,
		CoverLetter:   req.CoverLetter,
		AnswersJSON:   req.AnswersJSON,
		Status:        "new",
	}
	adminURL := "https://recruit.rtylr.com/admin/submissions"
	h.CareersService.SendSubmissionEmails(
		ctx, submission, career, custom,
		req.CandidateName, req.Email, adminURL,
	)

	return SubmitApplicationResponse{ID: id, Message: "Application received"}, nil
}
