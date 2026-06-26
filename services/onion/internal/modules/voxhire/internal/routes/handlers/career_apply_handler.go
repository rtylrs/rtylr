package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/voxire/onion/internal/modules/voxhire/internal/application"
	"github.com/voxire/onion/internal/modules/voxhire/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/storage"
)

// CareerApplyHandler handles POST /voxhire/careers/apply from the
// hire.voxire.com careers form. Multipart/form-data because the CV is
// attached as a file. Writes to voxire_career_applications (the shared
// careers applications table) with the recruiter-panel fields (location,
// salary expectation). CV upload is best-effort.
func CareerApplyHandler(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxApplicationMultipartBytes)
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

		var jobID int64
		if s := strings.TrimSpace(r.FormValue("job_id")); s != "" {
			if n, err := strconv.ParseInt(s, 10, 64); err == nil {
				jobID = n
			}
		}
		if jobID <= 0 {
			writeApplyError(w, http.StatusBadRequest, "job_id is required")
			return
		}

		var salary float64
		if s := strings.TrimSpace(r.FormValue("salary_expectation")); s != "" {
			if f, err := strconv.ParseFloat(s, 64); err == nil {
				salary = f
			}
		}

		in := domain.CareerApplicationInput{
			JobID:             jobID,
			Name:              name,
			Email:             email,
			Phone:             strings.TrimSpace(r.FormValue("phone")),
			Location:          strings.TrimSpace(r.FormValue("location")),
			SalaryExpectation: salary,
			WorkHistory:       r.FormValue("work_history"),
			Education:         r.FormValue("education"),
			ExperienceSummary: r.FormValue("experience_summary"),
			CoverLetter:       r.FormValue("cover_letter"),
		}

		// CV upload — best-effort. If R2 is not configured or the upload
		// fails, the application still saves without a CV key.
		file, header, fileErr := r.FormFile("cv")
		if fileErr == nil {
			defer file.Close()
			if header.Size > maxResumeUploadBytes {
				writeApplyError(w, http.StatusRequestEntityTooLarge, "CV must be 5 MB or smaller")
				return
			}

			contentType := header.Header.Get("Content-Type")
			if contentType == "" {
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

			if storage.AllowedResumeContentTypes[contentType] {
				client, storageErr := storage.Default()
				if storageErr == nil {
					safeName := storage.SanitizeObjectKey(header.Filename)
					key := fmt.Sprintf("voxire/applications/%d/%s-%s", jobID, uuid.NewString(), safeName)
					if uploadErr := client.Upload(r.Context(), key, contentType, file); uploadErr != nil {
						logger.Ctx(r.Context()).Warn("VoxHire career CV upload failed, proceeding without CV",
							logger.Error(uploadErr),
							logger.Any("applicant_email", email),
						)
					} else {
						in.CVKey = key
					}
				}
			}
		}

		id, err := svc.SubmitCareerApplication(r.Context(), in)
		if err != nil {
			logger.Ctx(r.Context()).Error("Failed to save VoxHire career application",
				logger.Error(err),
				logger.Any("email", email),
			)
			writeApplyError(w, http.StatusInternalServerError, "Failed to submit application")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"id":      id,
			"message": "Application received",
			"status":  "success",
		})
	}
}
