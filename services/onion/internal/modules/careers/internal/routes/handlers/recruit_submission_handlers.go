package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/careers/internal/application"
	"github.com/voxire/onion/internal/modules/careers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// ─────────────────────────────────────────────
// Route declarations
// ─────────────────────────────────────────────

func ListSubmissionsRoutes() []string        { return []string{"/recruit/submissions"} }
func UpdateSubmissionStatusRoutes() []string { return []string{"/recruit/submissions/status"} }

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

func (h *ListSubmissionsHandler) Serve(ctx context.Context, req *ListSubmissionsRequest) (ListSubmissionsResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListSubmissionsResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return ListSubmissionsResponse{}, errorModels.PermissionDeniedServerError()
	}

	submissions, err := h.CareersService.ListSubmissions(ctx, businessID, req.Status)
	if err != nil {
		return ListSubmissionsResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list submissions",
		}
	}

	payloads := make([]*SubmissionPayload, 0, len(submissions))
	for _, s := range submissions {
		payloads = append(payloads, toSubmissionPayload(s))
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

func (h *UpdateSubmissionStatusHandler) Validate(_ context.Context, req *UpdateSubmissionStatusRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "id is required"}
	}
	if req.Status == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "status is required"}
	}
	return nil
}

func (h *UpdateSubmissionStatusHandler) Serve(ctx context.Context, req *UpdateSubmissionStatusRequest) (UpdateSubmissionStatusResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateSubmissionStatusResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return UpdateSubmissionStatusResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.CareersService.UpdateSubmissionStatus(ctx, businessID, req.ID, req.Status); err != nil {
		return UpdateSubmissionStatusResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update submission status",
		}
	}
	return UpdateSubmissionStatusResponse{Message: "Submission updated"}, nil
}
