package api

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// CareerApplyHandler handles POST /voxhire/careers/apply from the
// hire.voxire.com careers form. Multipart/form-data because the CV is
// attached as a file. Writes to voxire_career_applications (the shared
// careers applications table) with the recruiter-panel fields (location,
// salary expectation). A CV must be attached. If it cannot be stored, the
// application is still saved, flagged upload_failed, and the team is alerted.
func CareerApplyHandler(svc CareerApplicationSubmitter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxApplicationMultipartBytes)
		// #nosec G120 -- MaxBytesReader caps the complete request before parsing.
		if err := r.ParseMultipartForm(maxResumeUploadBytes); err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				writeApplyCodeError(w, http.StatusRequestEntityTooLarge, applyCodeUploadTooLarge, "Upload is too large")
				return
			}
			writeApplyCodeError(w, http.StatusBadRequest, applyCodeInvalidForm, "Invalid multipart form")
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		email := strings.TrimSpace(r.FormValue("email"))
		if name == "" {
			writeApplyFieldError(w, http.StatusBadRequest, applyCodeMissingField, "name", "name is required")
			return
		}
		if email == "" {
			writeApplyFieldError(w, http.StatusBadRequest, applyCodeMissingField, "email", "email is required")
			return
		}

		var jobID int64
		if s := strings.TrimSpace(r.FormValue("job_id")); s != "" {
			if n, err := strconv.ParseInt(s, 10, 64); err == nil {
				jobID = n
			}
		}
		if jobID <= 0 {
			writeApplyFieldError(w, http.StatusBadRequest, applyCodeInvalidJobID, "job_id", "job_id is required")
			return
		}

		salary, salaryOK := parseSalaryExpectation(r.FormValue("salary_expectation"))
		if !salaryOK {
			writeApplyFieldError(w, http.StatusBadRequest, applyCodeInvalidSalary, "salary_expectation",
				"salary_expectation is out of range")
			return
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

		if f, tooLong := firstTooLong(careerFieldLimits(in)); tooLong {
			writeFieldTooLong(w, f)
			return
		}

		file, header, fileErr := r.FormFile("cv")
		if fileErr != nil {
			writeApplyFieldError(w, http.StatusBadRequest, applyCodeCVRequired, "cv", "CV is required")
			return
		}
		defer file.Close()
		if header.Size > maxResumeUploadBytes {
			writeApplyFieldError(w, http.StatusRequestEntityTooLarge, applyCodeCVTooLarge, "cv", "CV must be 5 MB or smaller")
			return
		}
		contentType := resumeContentType(header)
		if !allowedVoxhireResumeContentType(contentType) {
			writeApplyFieldError(w, http.StatusBadRequest, applyCodeCVUnsupportedType, "cv", "Only PDF or DOCX CVs are accepted")
			return
		}

		if id, dup := recentDuplicate(r.Context(), func() (int64, error) {
			return svc.RecentCareerApplication(r.Context(), jobID, email)
		}, email); dup {
			writeApplyDuplicate(w, id)
			return
		}

		key := fmt.Sprintf("voxire/applications/%d/%s-%s", jobID, uuid.NewString(), cvObjectName(header.Filename))
		if uploadErr := uploadCV(r.Context(), "VoxHire career", key, contentType, file, email); uploadErr != nil {
			in.CVUploadFailed = true
			in.CVUploadError = truncateRunes(uploadErr.Error(), maxCVErrorLength)
		} else {
			in.CVKey = key
		}

		id, err := svc.SubmitCareerApplication(r.Context(), in)
		if err != nil {
			logger.Ctx(r.Context()).Error("Failed to save VoxHire career application",
				logger.Error(err),
				logger.Any("email", email),
				logger.Any("job_id", jobID),
				logger.Bool("cv_upload_failed", in.CVUploadFailed),
			)
			writeApplyCodeError(w, http.StatusInternalServerError, applyCodeSaveFailed,
				"We could not save your application. Our team has been alerted.")
			return
		}

		if in.CVUploadFailed {
			logger.Ctx(r.Context()).Error("VoxHire career application saved without CV",
				logger.Any("application_id", id),
				logger.Any("job_id", jobID),
				logger.Any("applicant_email", email),
			)
		}

		writeApplySuccess(w, id, in.CVUploadFailed)
	}
}

// parseSalaryExpectation keeps the old leniency: an empty or unparsable value
// is stored as 0 rather than blocking the applicant. It rejects only values
// the DECIMAL(12,2) column cannot hold (NaN, infinities, out of range), which
// used to reach the insert and fail it with a 500.
func parseSalaryExpectation(raw string) (float64, bool) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, true
	}
	f, err := strconv.ParseFloat(s, 64)
	var numErr *strconv.NumError
	if errors.As(err, &numErr) && errors.Is(numErr.Err, strconv.ErrRange) {
		return 0, false
	}
	if err != nil {
		return 0, true
	}
	if math.IsNaN(f) || math.IsInf(f, 0) || math.Abs(f) > maxSalaryExpectation {
		return 0, false
	}
	return f, true
}
