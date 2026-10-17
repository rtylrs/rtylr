package api

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

const (
	maxResumeUploadBytes                 = 5 << 20 // 5 MiB
	maxApplicationMultipartOverheadBytes = 512 << 10
	maxApplicationMultipartBytes         = maxResumeUploadBytes + maxApplicationMultipartOverheadBytes
)

func ApplicationsRoutes() []string { return []string{"/voxhire/applications"} }

// ApplicationsHandler accepts talent-pool applications from the
// "Apply" form on hire.voxire.com. Multipart/form-data because the CV
// is uploaded as a file. A CV must be attached. If it cannot be stored,
// the application is still saved, flagged upload_failed, and the team is
// alerted, so an applicant is never turned away by a storage fault.
func ApplicationsHandler(svc TalentApplicationSubmitter) http.HandlerFunc {
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

		fullName := strings.TrimSpace(r.FormValue("fullName"))
		email := strings.TrimSpace(r.FormValue("email"))

		if fullName == "" {
			writeApplyFieldError(w, http.StatusBadRequest, applyCodeMissingField, "fullName", "fullName is required")
			return
		}
		if email == "" {
			writeApplyFieldError(w, http.StatusBadRequest, applyCodeMissingField, "email", "email is required")
			return
		}

		consent := strings.EqualFold(strings.TrimSpace(r.FormValue("consent")), "true")
		if !consent {
			writeApplyFieldError(w, http.StatusBadRequest, applyCodeConsentRequired, "consent", "consent is required")
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

		if f, tooLong := firstTooLong(talentFieldLimits(in)); tooLong {
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
			return svc.RecentTalentApplication(r.Context(), email)
		}, email); dup {
			writeApplyDuplicate(w, id)
			return
		}

		trackSlug := slugify(in.Track)
		if trackSlug == "" {
			trackSlug = "unspecified"
		}

		key := fmt.Sprintf("voxhire/applications/%s/%s-%s",
			trackSlug,
			uuid.NewString(),
			cvObjectName(header.Filename),
		)

		if uploadErr := uploadCV(r.Context(), "VoxHire", key, contentType, file, email); uploadErr != nil {
			in.CVUploadFailed = true
			in.CVUploadError = truncateRunes(uploadErr.Error(), maxCVErrorLength)
		} else {
			in.CVKey = key
		}

		id, err := svc.SubmitTalentApplication(r.Context(), in)
		if err != nil {
			logger.Ctx(r.Context()).Error("Failed to save VoxHire talent application",
				logger.Error(err),
				logger.Any("email", email),
				logger.Bool("cv_upload_failed", in.CVUploadFailed),
			)
			writeApplyCodeError(w, http.StatusInternalServerError, applyCodeSaveFailed,
				"We could not save your application. Our team has been alerted.")
			return
		}

		if in.CVUploadFailed {
			logger.Ctx(r.Context()).Error("VoxHire talent application saved without CV",
				logger.Any("application_id", id),
				logger.Any("applicant_email", email),
			)
		}

		writeApplySuccess(w, id, in.CVUploadFailed)
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

// writeApplyError writes a plain onekit error. The admin handlers still use
// it; the public apply handlers use writeApplyCodeError, which adds a code.
func writeApplyError(w http.ResponseWriter, code int, msg string) {
	serverHttp.WriteOneKitError(context.Background(), w, onekiterrors.New(code, msg))
}

// slugify lowercases and turns whitespace + slashes into hyphens for
// use in object keys. Keeps alphanumeric and hyphens.
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
