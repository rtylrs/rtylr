package domain

// ApplicationInput is the write shape for POST /careers/apply.
// The Voxire website posts multipart form data; the handler
// maps each part to this struct before calling the service.
// CVKey is the R2 object key for the uploaded file — empty if
// the upload was skipped or storage is not configured.
type ApplicationInput struct {
	JobID             int64
	Name              string
	Email             string
	Phone             string
	WorkHistory       string // JSON-encoded array from the frontend
	Education         string // JSON-encoded array from the frontend
	ExperienceSummary string
	CoverLetter       string
	CVKey             string
}
