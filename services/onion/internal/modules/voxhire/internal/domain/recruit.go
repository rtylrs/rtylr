package domain

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Sentinel errors for the recruiter-admin surface.
var (
	ErrAdminNotFound       = errors.New("voxhire admin not found")
	ErrInvalidCredentials  = errors.New("invalid voxhire admin credentials")
	ErrAdminExists         = errors.New("voxhire admin already exists")
	ErrApplicationNotFound = errors.New("voxhire application not found")
	ErrCareerNotFound      = errors.New("voxhire career not found")
)

// Admin is a recruiter/HR account that can sign in to the
// hire.voxire.com admin panel. Self-contained: not tied to the rtylr
// team/user system. Admins are managed from within the panel.
type Admin struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// AdminClaims is the JWT payload for an authenticated admin session.
type AdminClaims struct {
	AdminID int64  `json:"admin_id"`
	Email   string `json:"email"`
	jwt.RegisteredClaims
}

func (c *AdminClaims) SetIssuer(issuer string) { c.Issuer = issuer }

func (c *AdminClaims) SetExpiration(exp time.Time) {
	if !exp.IsZero() {
		c.ExpiresAt = jwt.NewNumericDate(exp)
	}
}

// Opening is a job posting (read from the shared `careers` table) along
// with the count of applications received for it. Powers the position
// picker in the admin panel.
type Opening struct {
	ID             int64   `json:"id"`
	Slug           string  `json:"slug"`
	Title          string  `json:"title"`
	Department     string  `json:"department"`
	Location       string  `json:"location"`
	EmploymentType string  `json:"employment_type"`
	WorkMode       string  `json:"work_mode"`
	Seniority      string  `json:"seniority"`
	Details        string  `json:"details"`
	SalaryMin      float64 `json:"salary_min"`
	SalaryMax      float64 `json:"salary_max"`
	SalaryCurrency string  `json:"salary_currency"`
	IsActive       bool    `json:"is_active"`
	NewCount       int64   `json:"new_count"`
	TotalCount     int64   `json:"total_count"`
}

// CareerInput is the editable shape for the VoxHire admin jobs panel.
// It writes to the shared careers table that powers GET /careers.
type CareerInput struct {
	ID             int64
	Title          string
	Department     string
	Location       string
	EmploymentType string
	WorkMode       string
	Seniority      string
	Details        string
	SalaryMin      float64
	SalaryMax      float64
	SalaryCurrency string
}

// Application is the HR-panel view of a candidate application, read from
// voxire_career_applications.
type Application struct {
	ID                 int64      `json:"id"`
	JobID              int64      `json:"job_id"`
	Name               string     `json:"name"`
	Email              string     `json:"email"`
	Phone              string     `json:"phone"`
	Location           string     `json:"location"`
	SalaryExpectation  float64    `json:"salary_expectation"`
	SalaryCurrency     string     `json:"salary_currency"`
	Status             string     `json:"status"`
	WorkHistory        string     `json:"work_history"`
	Education          string     `json:"education"`
	ExperienceSummary  string     `json:"experience_summary"`
	CoverLetter        string     `json:"cover_letter"`
	HasCV              bool       `json:"has_cv"`
	CVProcessingStatus string     `json:"cv_processing_status"`
	CVProcessingError  string     `json:"cv_processing_error"`
	CVSummary          string     `json:"cv_summary"`
	CVProcessedAt      *time.Time `json:"cv_processed_at,omitempty"`
	ReviewRating       int        `json:"review_rating"`
	ReviewNotes        string     `json:"review_notes"`
	NextStep           string     `json:"next_step"`
	ReviewedAt         *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

// ApplicationReviewInput is the recruiter-owned review state for a
// candidate application.
type ApplicationReviewInput struct {
	Status       string
	ReviewRating int
	ReviewNotes  string
	NextStep     string
}

// ApplicationFilter controls the admin applications list query.
type ApplicationFilter struct {
	JobID    int64
	Sort     string // "recent" | "salary_asc" | "salary_desc"
	Location string // substring match on location
	Query    string // search over application fields and extracted CV text
	Limit    int32
	Offset   int32
}

// Analytics is the aggregate recruiter view used by the admin overview.
type Analytics struct {
	Totals               AnalyticsTotals     `json:"totals"`
	ApplicationsByStatus []AnalyticsBucket   `json:"applications_by_status"`
	LeadServiceMix       []AnalyticsBucket   `json:"lead_service_mix"`
	LeadTimelineMix      []AnalyticsBucket   `json:"lead_timeline_mix"`
	RecentDays           []AnalyticsDay      `json:"recent_days"`
	TopOpenings          []AnalyticsOpening  `json:"top_openings"`
	RecentActivity       []AnalyticsActivity `json:"recent_activity"`
}

type AnalyticsTotals struct {
	OpenJobs             int64   `json:"open_jobs"`
	ClosedJobs           int64   `json:"closed_jobs"`
	ApplicationsTotal    int64   `json:"applications_total"`
	Applications7d       int64   `json:"applications_7d"`
	Applications30d      int64   `json:"applications_30d"`
	NewApplications      int64   `json:"new_applications"`
	LeadsTotal           int64   `json:"leads_total"`
	Leads7d              int64   `json:"leads_7d"`
	Leads30d             int64   `json:"leads_30d"`
	AdminsTotal          int64   `json:"admins_total"`
	TalentPoolTotal      int64   `json:"talent_pool_total"`
	TalentPool7d         int64   `json:"talent_pool_7d"`
	TalentPool30d        int64   `json:"talent_pool_30d"`
	CVProcessedTotal     int64   `json:"cv_processed_total"`
	CVPendingTotal       int64   `json:"cv_pending_total"`
	CVErrorTotal         int64   `json:"cv_error_total"`
	CVRate               float64 `json:"cv_rate"`
	AvgSalaryExpectation float64 `json:"avg_salary_expectation"`
}

type AnalyticsBucket struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

type AnalyticsDay struct {
	Date         string `json:"date"`
	Applications int64  `json:"applications"`
	Leads        int64  `json:"leads"`
}

type AnalyticsOpening struct {
	ID                int64      `json:"id"`
	Title             string     `json:"title"`
	IsActive          bool       `json:"is_active"`
	TotalCount        int64      `json:"total_count"`
	NewCount          int64      `json:"new_count"`
	RecentCount       int64      `json:"recent_count"`
	LastApplicationAt *time.Time `json:"last_application_at,omitempty"`
}

type AnalyticsActivity struct {
	Kind      string    `json:"kind"`
	ID        int64     `json:"id"`
	Primary   string    `json:"primary"`
	Secondary string    `json:"secondary"`
	Meta      string    `json:"meta"`
	CreatedAt time.Time `json:"created_at"`
}

// Lead is the admin-panel view of a company hiring inquiry submitted
// through the "Talk to us about a role" form (voxhire_company_leads).
type Lead struct {
	ID           int64     `json:"id"`
	CompanyName  string    `json:"company_name"`
	ContactName  string    `json:"contact_name"`
	ContactTitle string    `json:"contact_title"`
	WorkEmail    string    `json:"work_email"`
	Phone        string    `json:"phone"`
	Role         string    `json:"role"`
	Headcount    string    `json:"headcount"`
	Timeline     string    `json:"timeline"`
	Service      string    `json:"service"`
	Message      string    `json:"message"`
	AgencyOptIn  bool      `json:"agency_opt_in"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

// CareerApplicationInput is a public application submitted through the
// careers form on hire.voxire.com. Written to voxire_career_applications.
type CareerApplicationInput struct {
	JobID             int64
	Name              string
	Email             string
	Phone             string
	Location          string
	SalaryExpectation float64
	WorkHistory       string
	Education         string
	ExperienceSummary string
	CoverLetter       string
	CVKey             string
}

// CVProcessingTarget is a stored CV that the worker can download,
// extract, and write back into searchable columns.
type CVProcessingTarget struct {
	Source string
	ID     int64
	CVKey  string
}

type CVProcessingResult struct {
	Source        string
	ID            int64
	Status        string
	ExtractedText string
	Summary       string
	Error         string
}
