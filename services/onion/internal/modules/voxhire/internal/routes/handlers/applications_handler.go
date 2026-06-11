package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/voxire/onion/internal/modules/voxhire/internal/application"
	"github.com/voxire/onion/internal/modules/voxhire/internal/domain"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
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
// is uploaded as a file. Best-effort: if R2 is not configured or the
// upload fails, the application is still saved without a CV key so the
// candidate isn't blocked.
func ApplicationsHandler(svc *application.Service) http.HandlerFunc {
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
			LinkedinURL:        strings.TrimSpace(r.FormValue("linkedinUrl")),
			PortfolioURL:       strings.TrimSpace(r.FormValue("portfolioUrl")),
			Pitch:              strings.TrimSpace(r.FormValue("pitch")),
			Consent:            consent,
		}

		// CV upload — best effort. If R2 is unavailable the application
		// still saves; the notification email flags that the CV wasn't
		// stored.
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
					trackSlug := slugify(in.Track)
					if trackSlug == "" {
						trackSlug = "unspecified"
					}

					key := fmt.Sprintf("voxhire/applications/%s/%s-%s",
						trackSlug,
						uuid.NewString(),
						safeName,
					)

					uploadCtx := r.Context()
					if uploadErr := client.Upload(uploadCtx, key, contentType, file); uploadErr != nil {
						logger.Ctx(r.Context()).Warn("VoxHire CV upload failed, proceeding without CV",
							logger.Error(uploadErr),
							logger.Any("applicant_email", email),
						)
					} else {
						in.CVKey = key
					}
				}
			}
		}

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

func writeApplyError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(errorModels.ErrorResponse{
		Error:  errorModels.ServerError{Code: code, Message: msg},
		Status: "error",
	}); err != nil {
		logger.Global().Error("Failed to encode application error response", logger.Error(err))
	}
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
