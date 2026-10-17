package api

import (
	"context"
	"io"
	"net/http"
	"path"
	"unicode/utf8"

	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/storage"
)

// Machine-readable codes for the public apply endpoints. Every error body
// carries one, so the frontend can show a specific message instead of a
// single generic failure for every non-2xx.
const (
	applyCodeInvalidForm       = "invalid_form"
	applyCodeUploadTooLarge    = "upload_too_large"
	applyCodeMissingField      = "missing_field"
	applyCodeConsentRequired   = "consent_required"
	applyCodeInvalidJobID      = "invalid_job_id"
	applyCodeInvalidSalary     = "invalid_salary"
	applyCodeFieldTooLong      = "field_too_long"
	applyCodeCVRequired        = "cv_required"
	applyCodeCVTooLarge        = "cv_too_large"
	applyCodeCVUnsupportedType = "cv_unsupported_type"
	applyCodeSaveFailed        = "application_save_failed"

	// applyCodeCVUploadFailed is a warning on a 200: the application was
	// saved, but the CV was not stored and the team has been alerted.
	applyCodeCVUploadFailed = "cv_upload_failed"
)

const (
	// mysqlTextMaxBytes is the capacity of a MySQL TEXT column, in bytes.
	mysqlTextMaxBytes = 65535
	// maxCVErrorLength bounds the storage error kept in cv_processing_error.
	maxCVErrorLength = 500
	// maxCVFileNameLength bounds the sanitized file name inside the object key,
	// keeping the key well inside cv_key (VARCHAR(500) on careers).
	maxCVFileNameLength = 120
)

// TalentApplicationSubmitter and CareerApplicationSubmitter are the parts of
// the application service the public apply handlers use.
type TalentApplicationSubmitter interface {
	RecentTalentApplication(ctx context.Context, email string) (int64, error)
	SubmitTalentApplication(ctx context.Context, in domain.TalentApplicationInput) (int64, error)
}

type CareerApplicationSubmitter interface {
	RecentCareerApplication(ctx context.Context, jobID int64, email string) (int64, error)
	SubmitCareerApplication(ctx context.Context, in domain.CareerApplicationInput) (int64, error)
}

// recentDuplicate reports an application the same applicant already saved a
// few minutes ago. A failed lookup is logged and treated as "none", so the
// guard can never block a submission.
func recentDuplicate(ctx context.Context, find func() (int64, error), email string) (int64, bool) {
	id, err := find()
	if err != nil {
		logger.Ctx(ctx).Warn("VoxHire duplicate check failed",
			logger.Error(err),
			logger.Any("applicant_email", email),
		)
		return 0, false
	}
	if id > 0 {
		logger.Ctx(ctx).Info("VoxHire duplicate application suppressed",
			logger.Any("application_id", id),
			logger.Any("applicant_email", email),
		)
	}
	return id, id > 0
}

// cvObjectStore is the part of storage.Client the apply handlers need.
type cvObjectStore interface {
	Upload(ctx context.Context, key, contentType string, body io.Reader) error
}

// openCVStore returns the CV object store. It is a variable so tests can
// simulate missing configuration and failed uploads.
var openCVStore = func() (cvObjectStore, error) {
	client, err := storage.Default()
	if err != nil {
		return nil, err
	}
	return client, nil
}

type applyErrorBody struct {
	Message   string `json:"message"`
	Code      string `json:"code"`
	Field     string `json:"field,omitempty"`
	MaxLength int    `json:"max_length,omitempty"`
}

func writeApplyCodeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, applyErrorBody{Message: msg, Code: code})
}

func writeApplyFieldError(w http.ResponseWriter, status int, code, field, msg string) {
	writeJSON(w, status, applyErrorBody{Message: msg, Code: code, Field: field})
}

func writeFieldTooLong(w http.ResponseWriter, f fieldLimit) {
	writeJSON(w, http.StatusBadRequest, applyErrorBody{
		Message:   f.name + " is too long",
		Code:      applyCodeFieldTooLong,
		Field:     f.name,
		MaxLength: f.max,
	})
}

// writeApplySuccess reports a saved application. A CV that could not be stored
// still gets a 200, since the application is safe and the team is alerted,
// but with a warning code.
func writeApplySuccess(w http.ResponseWriter, id int64, cvUploadFailed bool) {
	body := map[string]any{
		"id":        id,
		"message":   "Application received",
		"status":    "success",
		"cv_stored": !cvUploadFailed,
	}
	if cvUploadFailed {
		body["warning_code"] = applyCodeCVUploadFailed
		body["message"] = "Application received, but we could not store your CV. Our team will contact you to collect it."
	}
	writeJSON(w, http.StatusOK, body)
}

func writeApplyDuplicate(w http.ResponseWriter, id int64) {
	writeJSON(w, http.StatusOK, map[string]any{
		"id":        id,
		"message":   "Application received",
		"status":    "success",
		"duplicate": true,
	})
}

// fieldLimit is one submitted value checked against the column it lands in.
// VARCHAR(n) limits characters; TEXT limits bytes.
type fieldLimit struct {
	name  string
	value string
	max   int
	bytes bool
}

func varcharField(name, value string, maxChars int) fieldLimit {
	return fieldLimit{name: name, value: value, max: maxChars}
}

func textField(name, value string) fieldLimit {
	return fieldLimit{name: name, value: value, max: mysqlTextMaxBytes, bytes: true}
}

// firstTooLong returns the first value that would not fit its column. Without
// this check MySQL strict mode rejects the insert with "Data too long" and the
// applicant gets a 500 with no hint of which field to shorten.
func firstTooLong(limits []fieldLimit) (fieldLimit, bool) {
	for _, f := range limits {
		size := utf8.RuneCountInString(f.value)
		if f.bytes {
			size = len(f.value)
		}
		if size > f.max {
			return f, true
		}
	}
	return fieldLimit{}, false
}

// talentFieldLimits mirrors voxhire_talent_applications. Names are the form
// field names, so the client can point at the offending input.
func talentFieldLimits(in domain.TalentApplicationInput) []fieldLimit {
	return []fieldLimit{
		varcharField("fullName", in.FullName, 255),
		varcharField("email", in.Email, 255),
		varcharField("phone", in.Phone, 64),
		varcharField("countryOfResidence", in.CountryOfResidence, 128),
		varcharField("nationality", in.Nationality, 128),
		varcharField("track", in.Track, 64),
		varcharField("yearsExperience", in.YearsExperience, 32),
		varcharField("noticePeriod", in.NoticePeriod, 32),
		varcharField("openTo", in.OpenTo, 255),
		varcharField("currentRole", in.CurrentRole, 255),
		textField("desiredRoles", in.DesiredRoles),
		varcharField("salaryExpectation", in.SalaryExpectation, 128),
		varcharField("workAuthorization", in.WorkAuthorization, 255),
		textField("workHistory", in.WorkHistory),
		textField("education", in.Education),
		varcharField("linkedinUrl", in.LinkedinURL, 512),
		varcharField("portfolioUrl", in.PortfolioURL, 512),
		textField("pitch", in.Pitch),
	}
}

// careerFieldLimits mirrors voxire_career_applications.
func careerFieldLimits(in domain.CareerApplicationInput) []fieldLimit {
	return []fieldLimit{
		varcharField("name", in.Name, 255),
		varcharField("email", in.Email, 255),
		varcharField("phone", in.Phone, 100),
		varcharField("location", in.Location, 160),
		textField("work_history", in.WorkHistory),
		textField("education", in.Education),
		textField("experience_summary", in.ExperienceSummary),
		textField("cover_letter", in.CoverLetter),
	}
}

// maxSalaryExpectation is the largest value DECIMAL(12,2) holds.
const maxSalaryExpectation = 9_999_999_999.99

// uploadCV stores the CV and returns the error instead of failing the request:
// callers save the application either way. logLabel keeps the existing log
// messages ("VoxHire CV upload failed", "VoxHire career CV upload failed")
// so dashboards and saved searches keep matching.
func uploadCV(ctx context.Context, logLabel, key, contentType string, file io.Reader, email string) error {
	store, err := openCVStore()
	if err != nil {
		logger.Ctx(ctx).Error(logLabel+" CV storage unavailable",
			logger.Error(err),
			logger.Any("applicant_email", email),
		)
		return err
	}
	if err := store.Upload(ctx, key, contentType, file); err != nil {
		logger.Ctx(ctx).Error(logLabel+" CV upload failed",
			logger.Error(err),
			logger.Any("applicant_email", email),
			logger.String("cv_key", key),
		)
		return err
	}
	return nil
}

func truncateRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n])
}

// cvObjectName sanitizes the uploaded file name for the object key and caps
// its length, keeping the extension: the CV worker picks its parser from it.
func cvObjectName(filename string) string {
	name := storage.SanitizeObjectKey(filename)
	if len(name) <= maxCVFileNameLength {
		return name
	}
	ext := path.Ext(name)
	if len(ext) > 10 {
		ext = ""
	}
	return name[:maxCVFileNameLength-len(ext)] + ext
}
