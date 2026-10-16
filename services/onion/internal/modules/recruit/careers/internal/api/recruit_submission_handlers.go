package api

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/recruit/careers/internal/application"
	"github.com/voxire/onion/internal/modules/recruit/careers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/csvsafe"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// ─────────────────────────────────────────────
// Route declarations
// ─────────────────────────────────────────────

func ListSubmissionsRoutes() []string        { return []string{"/recruit/submissions"} }
func UpdateSubmissionStatusRoutes() []string { return []string{"/recruit/submissions/status"} }
func ExportSubmissionsRoutes() []string      { return []string{"/recruit/submissions/export"} }

// ─────────────────────────────────────────────
// Shared payload
// ─────────────────────────────────────────────

type SubmissionPayload struct {
	ID            int64  `json:"id"`
	CareerID      int64  `json:"career_id"`
	OwnerID       int64  `json:"owner_id"`
	CandidateName string `json:"candidate_name"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	ResumeURL     string `json:"resume_url"`
	CoverLetter   string `json:"cover_letter"`
	Status        string `json:"status"`
	AnswersJSON   string `json:"answers_json"`
	CareerTitle   string `json:"career_title"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

func toSubmissionPayload(s domain.Submission) *SubmissionPayload {
	return &SubmissionPayload{
		ID:            s.ID,
		CareerID:      s.CareerID,
		OwnerID:       s.OwnerID,
		CandidateName: s.CandidateName,
		Email:         s.Email,
		Phone:         s.Phone,
		ResumeURL:     s.ResumeURL,
		CoverLetter:   s.CoverLetter,
		Status:        s.Status,
		AnswersJSON:   s.AnswersJSON,
		CareerTitle:   s.CareerTitle,
		CreatedAt:     s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     s.UpdatedAt.Format(time.RFC3339),
	}
}

// ─────────────────────────────────────────────
// GET /recruit/submissions
// ─────────────────────────────────────────────

type ListSubmissionsHandler struct {
	CareersService *application.Service
}

type ListSubmissionsRequest struct {
	// Optional filter: new, reviewed, interviewing, offered, hired, rejected.
	Status string `json:"status"`
}

type ListSubmissionsResponse struct {
	Submissions []*SubmissionPayload `json:"submissions"`
}

func (h *ListSubmissionsHandler) Serve(ctx context.Context, req *ListSubmissionsRequest) (ListSubmissionsResponse, error) {
	authUser, err := requireRecruitPermission(ctx, "recruit:candidates:view")
	if err != nil {
		return ListSubmissionsResponse{}, err
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return ListSubmissionsResponse{}, onekiterrors.PermissionDenied()
	}

	submissions, err := h.CareersService.ListSubmissions(ctx, businessID, req.Status)
	if err != nil {
		return ListSubmissionsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list submissions")
	}

	payloads := make([]*SubmissionPayload, 0, len(submissions))
	for _, s := range submissions {
		payload := toSubmissionPayload(s)
		if !authPkg.HasAppPermission(authUser, authorization.AppRecruit, "recruit:candidates:sensitive_view") {
			payload.Email = ""
			payload.Phone = ""
			payload.ResumeURL = ""
			payload.AnswersJSON = ""
		}
		payloads = append(payloads, payload)
	}
	return ListSubmissionsResponse{Submissions: payloads}, nil
}

// ─────────────────────────────────────────────
// PUT /recruit/submissions/status
// ─────────────────────────────────────────────

type UpdateSubmissionStatusHandler struct {
	CareersService *application.Service
}

type UpdateSubmissionStatusRequest struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

type UpdateSubmissionStatusResponse struct {
	Message string `json:"message"`
}

func (h *UpdateSubmissionStatusHandler) Validate(_ context.Context, req *UpdateSubmissionStatusRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "id is required")
	}
	if req.Status == "" {
		return onekiterrors.New(http.StatusBadRequest, "status is required")
	}
	return nil
}

func (h *UpdateSubmissionStatusHandler) Serve(ctx context.Context, req *UpdateSubmissionStatusRequest) (UpdateSubmissionStatusResponse, error) {
	authUser, err := requireRecruitPermission(ctx, "recruit:candidates:edit")
	if err != nil {
		return UpdateSubmissionStatusResponse{}, err
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return UpdateSubmissionStatusResponse{}, onekiterrors.PermissionDenied()
	}

	if err := h.CareersService.UpdateSubmissionStatus(ctx, businessID, req.ID, req.Status); err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return UpdateSubmissionStatusResponse{}, onekiterrors.New(http.StatusNotFound, "Submission not found")
		}
		return UpdateSubmissionStatusResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update submission status")
	}
	return UpdateSubmissionStatusResponse{Message: "Submission updated"}, nil
}

type ExportSubmissionsHandler struct {
	CareersService *application.Service
}

type ExportSubmissionsRequest struct {
	Status string `json:"status"`
}

type ExportSubmissionsResponse struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	CSV         string `json:"csv"`
}

// submissionsExportPageSize is how many submissions the CSV export pulls per
// query. The export must contain every row, so it walks pages instead of
// reusing ListSubmissions, which is capped at the UI's first 100 rows.
const submissionsExportPageSize = 200

func (h *ExportSubmissionsHandler) Serve(ctx context.Context, req *ExportSubmissionsRequest) (ExportSubmissionsResponse, error) {
	authUser, err := requireRecruitPermission(ctx, "recruit:candidates:export")
	if err != nil {
		return ExportSubmissionsResponse{}, err
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return ExportSubmissionsResponse{}, onekiterrors.PermissionDenied()
	}
	status := ""
	if req != nil {
		status = req.Status
	}

	var body bytes.Buffer
	w := csv.NewWriter(&body)
	if err := w.Write([]string{"id", "career_id", "candidate_name", "email", "phone", "resume_url", "status", "created_at"}); err != nil {
		return ExportSubmissionsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to export submissions")
	}
	writeRows := func(submissions []domain.Submission) error {
		for _, submission := range submissions {
			if err := w.Write(csvsafe.Row(
				strconv.FormatInt(submission.ID, 10),
				strconv.FormatInt(submission.CareerID, 10),
				submission.CandidateName,
				submission.Email,
				submission.Phone,
				submission.ResumeURL,
				submission.Status,
				submission.CreatedAt.Format(time.RFC3339),
			)); err != nil {
				return err
			}
		}
		return nil
	}

	// Page until a short page proves the end of the result set. Without this
	// the export silently stopped at ListSubmissions' LIMIT 100.
	for offset := int32(0); ; offset += submissionsExportPageSize {
		page, listErr := h.CareersService.ListSubmissionsPage(ctx, businessID, status, submissionsExportPageSize, offset)
		if listErr != nil {
			return ExportSubmissionsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to export submissions")
		}
		if err := writeRows(page); err != nil {
			return ExportSubmissionsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to export submissions")
		}
		if len(page) < submissionsExportPageSize {
			break
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return ExportSubmissionsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to export submissions")
	}
	return ExportSubmissionsResponse{Filename: "recruit_candidates.csv", ContentType: "text/csv", CSV: body.String()}, nil
}
