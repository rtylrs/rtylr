package domain

import "errors"

// Sentinel errors for the voxhire module.
var (
	ErrInvalidInput = errors.New("invalid voxhire input")
)

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
	CurrentRole        string
	DesiredRoles       string
	SalaryExpectation  string
	WorkAuthorization  string
	WorkHistory        string
	Education          string
	LinkedinURL        string
	PortfolioURL       string
	Pitch              string
	Consent            bool
	CVKey              string
}
