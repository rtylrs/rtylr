package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/application"
	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/domain"
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

func ApplicationsRoutes() []string { return []string{"/voxhire/applications"} }

// ApplicationsHandler accepts talent-pool applications from the
// "Apply" form on hire.voxire.com. Multipart/form-data because the CV
// is uploaded as a file. A stored CV is required: if upload storage is
// unavailable or the file cannot be stored, the submission is rejected.
func ApplicationsHandler(svc *application.Service) http.HandlerFunc {
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

		fullName := strings.TrimSpace(r.FormValue("fullName"))
		email := strings.TrimSpace(r.FormValue("email"))

		if fullName == "" {
			writeApplyError(w, http.StatusBadRequest, "fullName is required")
			return
		}
		if email == "" {
			writeApplyError(w, http.StatusBadRequest, "email is required")
			return
		}

		consent := strings.EqualFold(strings.TrimSpace(r.FormValue("consent")), "true")
		if !consent {
			writeApplyError(w, http.StatusBadRequest, "consent is required")
			return
		}

		in := domain.TalentApplicationInput{
			FullName:           fullName,
			Email:              email,
			Phone:              strings.TrimSpace(r.FormValue("phone")),
			CountryOfResidence: strings.TrimSpace(r.FormValue("countryOfResidence")),
			Nationality:        strings.TrimSpace(r.FormValue("nationality")),
			Track:              strings.TrimSpace(r.FormValue("track")),
			YearsExperience:    strings.TrimSpace(r.FormValue("yearsExperience")),
			NoticePeriod:       strings.TrimSpace(r.FormValue("noticePeriod")),
			OpenTo:             strings.TrimSpace(r.FormValue("openTo")),
			CurrentRole:        strings.TrimSpace(r.FormValue("currentRole")),
			DesiredRoles:       strings.TrimSpace(r.FormValue("desiredRoles")),
			SalaryExpectation:  strings.TrimSpace(r.FormValue("salaryExpectation")),
			WorkAuthorization:  strings.TrimSpace(r.FormValue("workAuthorization")),
			WorkHistory:        strings.TrimSpace(r.FormValue("workHistory")),
			Education:          strings.TrimSpace(r.FormValue("education")),
			LinkedinURL:        strings.TrimSpace(r.FormValue("linkedinUrl")),
			PortfolioURL:       strings.TrimSpace(r.FormValue("portfolioUrl")),
			Pitch:              strings.TrimSpace(r.FormValue("pitch")),
			Consent:            consent,
		}

		file, header, fileErr := r.FormFile("cv")
		if fileErr != nil {
			writeApplyError(w, http.StatusBadRequest, "CV is required")
			return
		}
		defer file.Close()
		if header.Size > maxResumeUploadBytes {
			writeApplyError(w, http.StatusRequestEntityTooLarge, "CV must be 5 MB or smaller")
			return
		}

		contentType := resumeContentType(header)
		if !allowedVoxhireResumeContentType(contentType) {
			writeApplyError(w, http.StatusBadRequest, "Only PDF or DOCX CVs are accepted")
			return
		}

		client, storageErr := storage.Default()
		if storageErr != nil {
			logger.Ctx(r.Context()).Error("VoxHire CV storage unavailable",
				logger.Error(storageErr),
				logger.Any("applicant_email", email),
			)
			writeApplyError(w, http.StatusInternalServerError, "CV upload storage is unavailable")
			return
		}

		safeName := storage.SanitizeObjectKey(header.Filename)
		trackSlug := slugify(in.Track)
		if trackSlug == "" {
			trackSlug = "unspecified"
		}

		key := fmt.Sprintf("voxhire/applications/%s/%s-%s",
			trackSlug,
			uuid.NewString(),
			safeName,
		)

		if uploadErr := client.Upload(r.Context(), key, contentType, file); uploadErr != nil {
			logger.Ctx(r.Context()).Error("VoxHire CV upload failed",
				logger.Error(uploadErr),
				logger.Any("applicant_email", email),
			)
			writeApplyError(w, http.StatusInternalServerError, "Failed to store CV")
			return
		}
		in.CVKey = key

		id, err := svc.SubmitTalentApplication(r.Context(), in)
		if err != nil {
			logger.Ctx(r.Context()).Error("Failed to save VoxHire talent application",
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
			logger.Global().Error("Failed to encode application success response", logger.Error(err))
		}
	}
}

func resumeContentType(header *multipart.FileHeader) string {
	contentType := strings.ToLower(strings.TrimSpace(header.Header.Get("Content-Type")))
	if contentType != "" && contentType != "application/octet-stream" {
		return contentType
	}
	fname := strings.ToLower(header.Filename)
	switch {
	case strings.HasSuffix(fname, ".pdf"):
		return "application/pdf"
	case strings.HasSuffix(fname, ".docx"):
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case strings.HasSuffix(fname, ".doc"):
		return "application/msword"
	default:
		return ""
	}
}

func allowedVoxhireResumeContentType(contentType string) bool {
	switch contentType {
	case "application/pdf", "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return true
	default:
		return false
	}
}

func writeApplyError(w http.ResponseWriter, code int, msg string) {
	serverHttp.WriteOneKitError(context.Background(), w, onekiterrors.New(code, msg))
}

// slugify lowercases and turns whitespace + slashes into hyphens for
// use in R2 object keys. Keeps alphanumeric and hyphens.
func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return ""
	}
	var b strings.Builder
	prevHyphen := false
	for _, r := range s {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			prevHyphen = false
		case r == ' ' || r == '-' || r == '/' || r == '_':
			if !prevHyphen && b.Len() > 0 {
				b.WriteRune('-')
				prevHyphen = true
			}
		}
	}
	out := strings.TrimRight(b.String(), "-")
	return out
}
