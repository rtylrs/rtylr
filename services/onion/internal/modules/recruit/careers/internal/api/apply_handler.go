package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/voxire/onion/internal/modules/recruit/careers/internal/application"
	"github.com/voxire/onion/internal/modules/recruit/careers/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platform/storage"
)

const (
	maxResumeUploadBytes                 = 5 << 20 // 5 MiB
	maxApplicationMultipartOverheadBytes = 512 << 10
	maxApplicationMultipartBytes         = maxResumeUploadBytes + maxApplicationMultipartOverheadBytes
)

func ApplyRoutes() []string { return []string{"/careers/apply"} }

// ApplyHandler handles POST /careers/apply from the Voxire website.
// The request is multipart/form-data because the browser attaches the
// applicant's CV as a file. The generic JSON wrapHandler can't parse
// multipart, so this handler is registered directly on chi and manages
// the full request/response cycle itself.
func ApplyHandler(_ interface{ Done() <-chan struct{} }, svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxApplicationMultipartBytes)
		// #nosec G120 -- MaxBytesReader caps the complete request before parsing.
		if err := r.ParseMultipartForm(maxResumeUploadBytes); err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				writeApplyError(w, http.StatusRequestEntityTooLarge, "Upload is too large")
				return
			}
			writeApplyError(w, http.StatusBadRequest, "Invalid multipart form")
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		email := strings.TrimSpace(r.FormValue("email"))

		if name == "" {
			writeApplyError(w, http.StatusBadRequest, "name is required")
			return
		}
		if email == "" {
			writeApplyError(w, http.StatusBadRequest, "email is required")
			return
		}

		if addr, parseErr := mail.ParseAddress(email); parseErr != nil || addr.Address != email || len(email) > maxEmailLen {
			writeApplyError(w, http.StatusBadRequest, "email is not valid")
			return
		}
		if utf8.RuneCountInString(name) > maxCandidateNameLen {
			writeApplyError(w, http.StatusBadRequest, "name is too long")
			return
		}

		jobIDStr := strings.TrimSpace(r.FormValue("job_id"))
		var jobID int64
		if jobIDStr != "" {
			n, parseErr := strconv.ParseInt(jobIDStr, 10, 64)
			if parseErr != nil {
				writeApplyError(w, http.StatusBadRequest, "job_id must be a number")
				return
			}
			jobID = n
		}
		// job_id=0 rows join no opening in the recruiter pipeline, so reject
		// a missing/malformed id instead of storing an orphan application.
		if jobID <= 0 {
			writeApplyError(w, http.StatusBadRequest, "job_id is required")
			return
		}

		in := domain.ApplicationInput{
			JobID:             jobID,
			Name:              name,
			Email:             email,
			Phone:             strings.TrimSpace(r.FormValue("phone")),
			WorkHistory:       r.FormValue("work_history"),
			Education:         r.FormValue("education"),
			ExperienceSummary: r.FormValue("experience_summary"),
			CoverLetter:       r.FormValue("cover_letter"),
		}

		// ── CV upload ──────────────────────────────────────────────────
		// Best-effort: if R2 storage is not configured or the upload
		// fails, we still save the application without a CV key so the
		// applicant isn't blocked. The notification email will note that
		// the CV was not attached.
		file, header, fileErr := r.FormFile("cv")
		if fileErr == nil {
			defer file.Close()
			if header.Size > maxResumeUploadBytes {
				writeApplyError(w, http.StatusRequestEntityTooLarge, "CV must be 5 MB or smaller")
				return
			}

			contentType := header.Header.Get("Content-Type")
			if contentType == "" || contentType == "application/octet-stream" {
				// Fallback MIME detection from extension
				fname := strings.ToLower(header.Filename)
				switch {
				case strings.HasSuffix(fname, ".pdf"):
					contentType = "application/pdf"
				case strings.HasSuffix(fname, ".docx"):
					contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
				case strings.HasSuffix(fname, ".doc"):
					contentType = "application/msword"
				}
			}

			if !storage.AllowedResumeContentTypes[contentType] {
				writeApplyError(w, http.StatusBadRequest, "Only PDF or Word CVs are accepted")
				return
			}
			client, storageErr := storage.Default()
			if storageErr == nil {
				safeName := storage.SanitizeObjectKey(header.Filename)
				key := fmt.Sprintf("voxire/applications/%d/%s-%s",
					jobID,
					uuid.NewString(),
					safeName,
				)

				uploadCtx := r.Context()

				uploadErr := client.Upload(uploadCtx, key, contentType, file)
				if uploadErr != nil {
					logger.Ctx(r.Context()).Warn("CV upload failed, proceeding without CV",
						logger.Error(uploadErr),
						logger.Any("applicant_email", email),
					)
				} else {
					in.CVKey = key
				}
			}
		}

		id, err := svc.SaveApplication(r.Context(), in)
		if err != nil {
			logger.Ctx(r.Context()).Error("Failed to save career application",
				logger.Error(err),
				logger.Any("email", email),
			)
			writeApplyError(w, http.StatusInternalServerError, "Failed to submit application")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      id,
			"message": "Application received",
			"status":  "success",
		}); err != nil {
			logger.Ctx(r.Context()).Warn("failed to write apply success response", logger.Error(err))
		}
	}
}

func writeApplyError(w http.ResponseWriter, code int, msg string) {
	serverHttp.WriteOneKitError(context.Background(), w, onekiterrors.New(code, msg))
}
