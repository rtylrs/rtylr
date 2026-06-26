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
	ID             int64  `json:"id"`
	Title          string `json:"title"`
	Department     string `json:"department"`
	Location       string `json:"location"`
	EmploymentType string `json:"employment_type"`
	WorkMode       string `json:"work_mode"`
	Seniority      string `json:"seniority"`
	NewCount       int64  `json:"new_count"`
	TotalCount     int64  `json:"total_count"`
}

// Application is the HR-panel view of a candidate application, read from
// voxire_career_applications.
type Application struct {
	ID                int64     `json:"id"`
	JobID             int64     `json:"job_id"`
	Name              string    `json:"name"`
	Email             string    `json:"email"`
	Phone             string    `json:"phone"`
	Location          string    `json:"location"`
	SalaryExpectation float64   `json:"salary_expectation"`
	SalaryCurrency    string    `json:"salary_currency"`
	Status            string    `json:"status"`
	WorkHistory       string    `json:"work_history"`
	Education         string    `json:"education"`
	ExperienceSummary string    `json:"experience_summary"`
	CoverLetter       string    `json:"cover_letter"`
	HasCV             bool      `json:"has_cv"`
	CreatedAt         time.Time `json:"created_at"`
}

// ApplicationFilter controls the admin applications list query.
type ApplicationFilter struct {
	JobID    int64
	Sort     string // "recent" | "salary_asc" | "salary_desc"
	Location string // substring match on location
	Query    string // experience search over work history / summary / cover letter
	Limit    int32
	Offset   int32
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
