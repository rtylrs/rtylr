package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/voxire/onion/internal/modules/recruit/careers/internal/application"
	"github.com/voxire/onion/internal/modules/recruit/careers/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/storage"
)

// publicCareersCacheScope namespaces the per-business public careers page
// cache. The careers, careers_customizations and careers_sections tables
// are written only within this module, so a short TTL plus invalidation
// on every careers/customization/section mutation keeps the anonymous
// public page consistent.
//
// The cache is keyed by business_id — the id every write handler already
// has — not by the customization's owner_id. owner_id is a users.id
// (businesses.owner_id) and only coincides with the business id by
// accident; keying the reader on owner_id while writers invalidated
// business_id left the public page stale for the whole TTL whenever they
// differed.
const publicCareersCacheScope = "careers_public"

// publicCareersCacheKey is the single source of truth for the cached public
// careers page key. The reader (ListPublicCareersHandler) and the
// invalidation helper must derive it from the same id: the business id.
func publicCareersCacheKey(businessID int64) string {
	return cacheKeys.Scoped(publicCareersCacheScope, businessID, "list")
}

// invalidatePublicCareersCache wipes the cached public careers page for a
// business. Call from every careers/customization/section create/update/delete
// so the public site reflects changes within the TTL window at worst.
func invalidatePublicCareersCache(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(publicCareersCacheScope, businessID))
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

func (h *GetPublicBrandingHandler) Validate(_ context.Context, req *GetPublicBrandingRequest) error {
	if req.Slug == "" {
		return onekiterrors.New(http.StatusBadRequest, "slug is required")
	}
	return nil
}

func (h *GetPublicBrandingHandler) Serve(ctx context.Context, req *GetPublicBrandingRequest) (GetPublicBrandingResponse, error) {
	c, err := h.CareersService.GetCustomizationBySlug(ctx, req.Slug)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return GetPublicBrandingResponse{}, onekiterrors.New(http.StatusNotFound, "Careers page not found")
		}
		return GetPublicBrandingResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to load branding")
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

func (h *ListPublicCareersHandler) Validate(_ context.Context, req *ListPublicCareersRequest) error {
	if req.Slug == "" {
		return onekiterrors.New(http.StatusBadRequest, "slug is required")
	}
	return nil
}

func (h *ListPublicCareersHandler) Serve(ctx context.Context, req *ListPublicCareersRequest) (ListPublicCareersResponse, error) {
	custom, err := h.CareersService.GetCustomizationBySlug(ctx, req.Slug)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return ListPublicCareersResponse{}, onekiterrors.New(http.StatusNotFound, "Careers page not found")
		}
		return ListPublicCareersResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to load careers")
	}

	res, cacheErr := cacher.Of[ListPublicCareersResponse]().GetOrComputeSingle(
		ctx,
		publicCareersCacheKey(custom.BusinessID),
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
		return ListPublicCareersResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to load careers")
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

func (h *GetPublicCareerHandler) Validate(_ context.Context, req *GetPublicCareerRequest) error {
	if req.Slug == "" || req.CareerSlug == "" {
		return onekiterrors.New(http.StatusBadRequest, "slug and career_slug are required")
	}
	return nil
}

func (h *GetPublicCareerHandler) Serve(ctx context.Context, req *GetPublicCareerRequest) (GetPublicCareerResponse, error) {
	custom, err := h.CareersService.GetCustomizationBySlug(ctx, req.Slug)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return GetPublicCareerResponse{}, onekiterrors.New(http.StatusNotFound, "Careers page not found")
		}
		return GetPublicCareerResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to load career")
	}

	c, err := h.CareersService.GetCareerBySlug(ctx, custom.BusinessID, req.CareerSlug)
	if err != nil || !c.IsPublished {
		return GetPublicCareerResponse{}, onekiterrors.New(http.StatusNotFound, "Role not found")
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

func (h *SubmitApplicationHandler) Validate(_ context.Context, req *SubmitApplicationRequest) error {
	if req.Slug == "" || req.CareerSlug == "" {
		return onekiterrors.New(http.StatusBadRequest, "slug and career_slug are required")
	}
	if req.CandidateName == "" {
		return onekiterrors.New(http.StatusBadRequest, "Name is required")
	}
	if req.Email == "" {
		return onekiterrors.New(http.StatusBadRequest, "Email is required")
	}
	return validateSubmissionFields(req)
}

const (
	maxCandidateNameLen = 160
	maxEmailLen         = 160
	maxPhoneLen         = 40
	maxResumeURLLen     = 500
	maxCoverLetterLen   = 20000
	maxAnswersJSONBytes = 64 * 1024
)

// validateSubmissionFields rejects input the columns cannot hold, so a
// candidate gets a specific 400 instead of a generic failure that loses the
// application, and the HR inbox never has to decode malformed answers.
func validateSubmissionFields(req *SubmitApplicationRequest) error {
	if utf8.RuneCountInString(req.CandidateName) > maxCandidateNameLen {
		return onekiterrors.New(http.StatusBadRequest, "Name is too long")
	}
	if len(req.Email) > maxEmailLen {
		return onekiterrors.New(http.StatusBadRequest, "Email is too long")
	}
	if addr, err := mail.ParseAddress(req.Email); err != nil || addr.Address != strings.TrimSpace(req.Email) {
		return onekiterrors.New(http.StatusBadRequest, "Email is not valid")
	}
	if utf8.RuneCountInString(req.Phone) > maxPhoneLen {
		return onekiterrors.New(http.StatusBadRequest, "Phone number is too long")
	}
	if len(req.ResumeURL) > maxResumeURLLen {
		return onekiterrors.New(http.StatusBadRequest, "Resume link is too long")
	}
	if utf8.RuneCountInString(req.CoverLetter) > maxCoverLetterLen {
		return onekiterrors.New(http.StatusBadRequest, "Cover letter is too long")
	}
	if req.AnswersJSON != "" {
		if len(req.AnswersJSON) > maxAnswersJSONBytes {
			return onekiterrors.New(http.StatusBadRequest, "Answers are too long")
		}
		var answers map[string]any
		if err := json.Unmarshal([]byte(req.AnswersJSON), &answers); err != nil {
			return onekiterrors.New(http.StatusBadRequest, "Answers must be a JSON object")
		}
	}
	return nil
}

func (h *SubmitApplicationHandler) Serve(ctx context.Context, req *SubmitApplicationRequest) (SubmitApplicationResponse, error) {
	// Resolve owner + career via the slug pair. Looking both up here
	// means the candidate never sends a tenant id; a malicious client
	// can only apply to careers that are actually published by the
	// given brand.
	custom, err := h.CareersService.GetCustomizationBySlug(ctx, req.Slug)
	if err != nil {
		return SubmitApplicationResponse{}, onekiterrors.New(http.StatusNotFound, "Careers page not found")
	}

	career, err := h.CareersService.GetCareerBySlug(ctx, custom.BusinessID, req.CareerSlug)
	if err != nil || !career.IsPublished {
		return SubmitApplicationResponse{}, onekiterrors.New(http.StatusNotFound, "Role not found or no longer open")
	}
	if !isOwnedResumeURL(req.ResumeURL, custom.OwnerID, career.ID) {
		return SubmitApplicationResponse{}, onekiterrors.New(http.StatusBadRequest, "Resume does not belong to this role")
	}
	if isInternalResumeKey(req.ResumeURL) {
		client, storageErr := storage.Default()
		if storageErr != nil {
			code, message := storage.BuildHTTPError(storageErr)
			return SubmitApplicationResponse{}, onekiterrors.New(code, message)
		}
		metadata, metadataErr := client.Metadata(ctx, req.ResumeURL)
		if metadataErr != nil {
			return SubmitApplicationResponse{}, onekiterrors.New(http.StatusBadRequest, "Resume upload was not found")
		}
		if err := validateResumeMetadata(metadata); err != nil {
			return SubmitApplicationResponse{}, err
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
		return SubmitApplicationResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to submit application")
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

func isOwnedResumeURL(resumeURL string, ownerID, careerID int64) bool {
	if !isInternalResumeKey(resumeURL) {
		return true
	}

	return strings.HasPrefix(resumeURL, fmt.Sprintf("resumes/%d/%d/", ownerID, careerID))
}

func isInternalResumeKey(resumeURL string) bool {
	return resumeURL != "" && !strings.HasPrefix(resumeURL, "http://") && !strings.HasPrefix(resumeURL, "https://")
}

// The messages are shown to the applicant verbatim, so they stay sentence-case.
// Building them with onekiterrors rather than fmt.Errorf keeps that copy intact
// while satisfying the lowercase-error-string rule, and matches how the sibling
// upload handler reports the same two validation failures.
func validateResumeMetadata(metadata storage.ObjectMetadata) error {
	if !storage.AllowedResumeContentTypes[metadata.ContentType] {
		return onekiterrors.New(http.StatusBadRequest, "Only PDF or Word resumes are accepted")
	}
	if metadata.Size > storage.MaxResumeBytes {
		return onekiterrors.New(http.StatusBadRequest, "Resume must be 10 MB or smaller")
	}
	return nil
}
