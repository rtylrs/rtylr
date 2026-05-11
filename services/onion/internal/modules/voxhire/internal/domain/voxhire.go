package domain

import "errors"

// Sentinel errors for the voxhire module.
var (
	ErrInvalidInput = errors.New("invalid voxhire input")
)

// CompanyLead is a company hiring inquiry submitted through the
// "Tell us about the role" form on hire.voxire.com.
type CompanyLead struct {
	ID           int64  `json:"id"`
	CompanyName  string `json:"company_name"`
	ContactName  string `json:"contact_name"`
	ContactTitle string `json:"contact_title"`
	WorkEmail    string `json:"work_email"`
	Phone        string `json:"phone"`
	Role         string `json:"role"`
	Headcount    string `json:"headcount"`
	Timeline     string `json:"timeline"`
	Service      string `json:"service"`
	Message      string `json:"message"`
	AgencyOptIn  bool   `json:"agency_opt_in"`
}

// CompanyLeadInput is what the service accepts when a new lead arrives.
type CompanyLeadInput struct {
	CompanyName  string
	ContactName  string
	ContactTitle string
	WorkEmail    string
	Phone        string
	Role         string
	Headcount    string
	Timeline     string
	Service      string
	Message      string
	AgencyOptIn  bool
}

// TalentApplication is a candidate joining the VoxHire talent pool
// through the "Apply" form on hire.voxire.com. Multipart upload, with
// the CV stored in R2 and only the key persisted in MySQL.
type TalentApplication struct {
	ID                 int64  `json:"id"`
	FullName           string `json:"full_name"`
	Email              string `json:"email"`
	Phone              string `json:"phone"`
	CountryOfResidence string `json:"country_of_residence"`
	Nationality        string `json:"nationality"`
	Track              string `json:"track"`
	YearsExperience    string `json:"years_experience"`
	NoticePeriod       string `json:"notice_period"`
	OpenTo             string `json:"open_to"`
	LinkedinURL        string `json:"linkedin_url"`
	PortfolioURL       string `json:"portfolio_url"`
	Pitch              string `json:"pitch"`
	Consent            bool   `json:"consent"`
	CVKey              string `json:"cv_key"`
}

// TalentApplicationInput is what the service accepts when an application arrives.
// CVKey is set after a successful R2 upload (handler-level concern).
type TalentApplicationInput struct {
	FullName           string
	Email              string
	Phone              string
	CountryOfResidence string
	Nationality        string
	Track              string
	YearsExperience    string
	NoticePeriod       string
	OpenTo             string
	LinkedinURL        string
	PortfolioURL       string
	Pitch              string
	Consent            bool
	CVKey              string
}
