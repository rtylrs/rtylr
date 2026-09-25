package domain

import (
	"time"

	"github.com/voxire/onion/internal/platform/money"
)

// RecruitCareer is the tenant-scoped projection of the `careers` table
// used by the recruit product. The legacy `Career` struct (see
// career.go) stays wired to the sqlc-generated row for the global
// /careers endpoint; this struct is populated by hand-written queries
// in the repository so we don't have to regenerate sqlc every time we
// add a column.
type RecruitCareer struct {
	ID             int64
	OwnerID        int64
	BusinessID     int64
	Slug           string
	Title          string
	Department     string
	Location       string
	EmploymentType string
	WorkMode       string
	Seniority      string
	Details        string
	Requirements   string
	SalaryMin      money.Amount
	SalaryMax      money.Amount
	SalaryCurrency string
	IsPublished    bool
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// RecruitCareerInput is the write shape for create/update. We keep it
// separate from RecruitCareer so the service can compute IDs and
// timestamps without the handler having to know about them.
type RecruitCareerInput struct {
	ID             int64
	OwnerID        int64
	BusinessID     int64
	Slug           string
	Title          string
	Department     string
	Location       string
	EmploymentType string
	WorkMode       string
	Seniority      string
	Details        string
	Requirements   string
	SalaryMin      money.Amount
	SalaryMax      money.Amount
	SalaryCurrency string
	IsPublished    bool
}

// Customization is the per-business branding config. A business that
// signs up for the recruit product creates exactly one of these; its
// slug becomes the first path segment on the public careers site
// (recruit.rtylr.com/<slug>).
//
// Palette fields default to empty; the frontend resolver falls back
// to the selected template's defaults when a value is missing, so a
// half-configured row still renders coherently.
type Customization struct {
	ID             int64
	OwnerID        int64
	BusinessID     int64
	Slug           string
	BrandName      string
	Tagline        string
	AboutBlurb     string
	LogoURL        string
	PrimaryColor   string
	AccentColor    string
	BgColor        string
	TextColor      string
	MutedTextColor string
	FontPairing    string
	HeroImageURL   string
	FaviconURL     string
	Template       string
	HeaderHTML     string
	// SEO overrides — defaults are derived from brand_name + tagline
	// when these are empty.
	MetaTitle       string
	MetaDescription string
	// Legal / consent links surfaced in the apply form. Empty = hide.
	PrivacyURL string
	TermsURL   string
	// Analytics — frontend injects GA / Pixel loaders conditionally on
	// each ID being non-empty.
	GAMeasurementID string
	MetaPixelID     string
	// RequireConsent flips a GDPR consent checkbox on the apply form.
	RequireConsent bool
	// Email notification config. NotificationEmail is the HR-facing
	// inbox (often a team alias like jobs@acme.com); empty = no HR
	// ping on new submissions. EmailNotifications is the master
	// switch so the admin can pause without wiping the address.
	NotificationEmail  string
	EmailNotifications bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// CustomizationInput is the write shape for create/update. Slug is
// validated by the service — the unique index on the table is the
// backstop, not the front line.
type CustomizationInput struct {
	OwnerID            int64
	BusinessID         int64
	Slug               string
	BrandName          string
	Tagline            string
	AboutBlurb         string
	LogoURL            string
	PrimaryColor       string
	AccentColor        string
	BgColor            string
	TextColor          string
	MutedTextColor     string
	FontPairing        string
	HeroImageURL       string
	FaviconURL         string
	Template           string
	HeaderHTML         string
	MetaTitle          string
	MetaDescription    string
	PrivacyURL         string
	TermsURL           string
	GAMeasurementID    string
	MetaPixelID        string
	RequireConsent     bool
	NotificationEmail  string
	EmailNotifications bool
}

// Submission is a candidate's application to a career. resume_url is a
// plain URL for the MVP — we rely on the caller to have uploaded
// through whatever storage they configured. A proper upload endpoint
// is a follow-up; storing the URL keeps the schema stable across that
// change.
//
// AnswersJSON is a JSON object keyed by form_field_id → answer. The
// server never introspects it; the admin UI decodes it alongside the
// field definitions to render the HR inbox.
type Submission struct {
	ID            int64
	CareerID      int64
	OwnerID       int64
	BusinessID    int64
	CandidateName string
	Email         string
	Phone         string
	ResumeURL     string
	CoverLetter   string
	Status        string
	AnswersJSON   string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	// CareerTitle is a denormalised display field populated by the
	// join on ListSubmissions so HR screens don't need a second fetch.
	CareerTitle string
}

// SubmissionInput is the write shape for POST /recruit/public/apply.
// OwnerID is resolved server-side from the customization slug so the
// candidate never sends a tenant ID.
type SubmissionInput struct {
	CareerID      int64
	OwnerID       int64
	BusinessID    int64
	CandidateName string
	Email         string
	Phone         string
	ResumeURL     string
	CoverLetter   string
	AnswersJSON   string
}
