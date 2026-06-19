package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/voxire/onion/internal/modules/careers/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/storage"
	"github.com/voxire/onion/internal/platform/tenant"
)

// ─────────────────────────────────────────────
// Route declarations
// ─────────────────────────────────────────────

func ResumeUploadURLRoutes() []string   { return []string{"/recruit/public/resume-upload-url"} }
func ResumeDownloadURLRoutes() []string { return []string{"/recruit/submissions/resume-url"} }

// ─────────────────────────────────────────────
// POST /recruit/public/resume-upload-url  (anonymous)
//
// Returns a short-lived presigned PUT URL the candidate's browser
// uploads to directly. We keep the backend out of the file-bytes path
// so a 5 MB PDF doesn't occupy an API worker for the duration of the
// upload.
// ─────────────────────────────────────────────

type ResumeUploadURLHandler struct {
	CareersService *application.Service
}

type ResumeUploadURLRequest struct {
	Slug        string `json:"slug"`
	CareerSlug  string `json:"career_slug"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
}

type ResumeUploadURLResponse struct {
	// UploadURL is the short-lived presigned PUT URL.
	UploadURL string `json:"upload_url"`
	// Key is the R2 object key. The client should send it back on the
	// apply call as the resume_url so we can build a signed read URL
	// later without the client needing to keep the full URL around.
	Key string `json:"key"`
	// ExpiresAt is when the presigned URL stops working. The frontend
	// uses this to decide if it needs to re-request before retrying.
	ExpiresAt string `json:"expires_at"`
	// MaxBytes mirrors the server-side ceiling so the client can fail
	// fast before even starting the upload.
	MaxBytes int64 `json:"max_bytes"`
}

func (h *ResumeUploadURLHandler) Validate(_ context.Context, req *ResumeUploadURLRequest) *errorModels.ServerError {
	if req.Slug == "" || req.CareerSlug == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "slug and career_slug are required"}
	}
	if req.FileName == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "file_name is required"}
	}
	if req.ContentType == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "content_type is required"}
	}
	if !storage.AllowedResumeContentTypes[req.ContentType] {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Only PDF or Word resumes are accepted"}
	}
	return nil
}

func (h *ResumeUploadURLHandler) Serve(ctx context.Context, req *ResumeUploadURLRequest) (ResumeUploadURLResponse, *errorModels.ServerError) {
	// Resolve the career before handing out an upload URL — stops
	// anyone from using the presigner as free S3 for unrelated files.
	custom, err := h.CareersService.GetCustomizationBySlug(ctx, req.Slug)
	if err != nil {
		return ResumeUploadURLResponse{}, &errorModels.ServerError{
			Code:    http.StatusNotFound,
			Message: "Careers page not found",
		}
	}
	career, err := h.CareersService.GetCareerBySlug(ctx, custom.BusinessID, req.CareerSlug)
	if err != nil || !career.IsPublished {
		return ResumeUploadURLResponse{}, &errorModels.ServerError{
			Code:    http.StatusNotFound,
			Message: "Role not found or no longer open",
		}
	}

	client, err := storage.Default()
	if err != nil {
		code, msg := storage.BuildHTTPError(err)
		return ResumeUploadURLResponse{}, &errorModels.ServerError{Code: code, Message: msg}
	}

	// Object layout: resumes/<owner>/<career>/<uuid>-<sanitized-name>
	// — tenant-scoped so listing a bucket prefix can quickly pull
	// every submission for one business without a join through the DB.
	safeName := storage.SanitizeObjectKey(req.FileName)
	key := fmt.Sprintf("resumes/%d/%d/%s-%s", custom.OwnerID, career.ID, uuid.NewString(), safeName)

	const expiresIn = 15 * time.Minute
	uploadURL, err := client.PresignUpload(ctx, key, req.ContentType, expiresIn)
	if err != nil {
		return ResumeUploadURLResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to generate upload URL",
		}
	}

	return ResumeUploadURLResponse{
		UploadURL: uploadURL,
		Key:       key,
		ExpiresAt: time.Now().Add(expiresIn).UTC().Format(time.RFC3339),
		MaxBytes:  storage.MaxResumeBytes,
	}, nil
}

// ─────────────────────────────────────────────
// GET /recruit/submissions/resume-url?submission_id=<id>  (authenticated)
//
// Returns a 1-hour signed read URL for HR to open a candidate's
// resume. We don't return the R2 URL in ListSubmissions because those
// URLs would otherwise live forever in browser history; signing on
// demand keeps each view time-boxed.
// ─────────────────────────────────────────────

type ResumeDownloadURLHandler struct {
	CareersService *application.Service
}

type ResumeDownloadURLRequest struct {
	SubmissionID int64 `json:"submission_id"`
}

type ResumeDownloadURLResponse struct {
	URL       string `json:"url"`
	ExpiresAt string `json:"expires_at"`
}

func (h *ResumeDownloadURLHandler) Validate(_ context.Context, req *ResumeDownloadURLRequest) *errorModels.ServerError {
	if req.SubmissionID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "submission_id is required"}
	}
	return nil
}

func (h *ResumeDownloadURLHandler) Serve(ctx context.Context, req *ResumeDownloadURLRequest) (ResumeDownloadURLResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ResumeDownloadURLResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return ResumeDownloadURLResponse{}, errorModels.PermissionDeniedServerError()
	}

	// Look up the submission, scoped to the caller's tenant. The list
	// method filters by business_id already; finding it here means the
	// caller owns it.
	submissions, err := h.CareersService.ListSubmissions(ctx, businessID, "")
	if err != nil {
		return ResumeDownloadURLResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to locate submission",
		}
	}
	var resumeKey string
	for _, s := range submissions {
		if s.ID == req.SubmissionID {
			resumeKey = s.ResumeURL
			break
		}
	}
	if resumeKey == "" {
		return ResumeDownloadURLResponse{}, &errorModels.ServerError{
			Code:    http.StatusNotFound,
			Message: "Submission not found or has no resume",
		}
	}

	// Legacy rows may still carry a full URL rather than an R2 key.
	// If it looks like an absolute URL, return it as-is so HR can
	// still click through — presigning only applies to internal keys.
	if strings.HasPrefix(resumeKey, "http://") || strings.HasPrefix(resumeKey, "https://") {
		return ResumeDownloadURLResponse{
			URL:       resumeKey,
			ExpiresAt: time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339),
		}, nil
	}

	client, err := storage.Default()
	if err != nil {
		code, msg := storage.BuildHTTPError(err)
		return ResumeDownloadURLResponse{}, &errorModels.ServerError{Code: code, Message: msg}
	}

	const expiresIn = time.Hour
	downloadURL, err := client.PresignDownload(ctx, resumeKey, expiresIn)
	if err != nil {
		return ResumeDownloadURLResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to generate resume URL",
		}
	}
	return ResumeDownloadURLResponse{
		URL:       downloadURL,
		ExpiresAt: time.Now().Add(expiresIn).UTC().Format(time.RFC3339),
	}, nil
}
