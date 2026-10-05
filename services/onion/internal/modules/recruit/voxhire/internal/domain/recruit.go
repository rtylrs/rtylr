package domain

import (
	"context"
	"errors"
	"time"
)

// Sentinel errors for the recruiter-admin surface.
var (
	ErrAdminNotFound       = errors.New("voxhire admin not found")
	ErrInvalidCredentials  = errors.New("invalid voxhire admin credentials")
	ErrAdminExists         = errors.New("voxhire admin already exists")
	ErrApplicationNotFound = errors.New("voxhire application not found")
	ErrCareerNotFound      = errors.New("voxhire career not found")
	ErrOwnerScopeMissing   = errors.New("voxhire owner scope missing")
)

type ownerScopeContextKey struct{}

// WithOwnerID attaches the authenticated VoxHire owner scope to a request.
func WithOwnerID(ctx context.Context, ownerID int64) context.Context {
	return context.WithValue(ctx, ownerScopeContextKey{}, ownerID)
}

// OwnerIDFromContext returns the authenticated owner scope, if present.
func OwnerIDFromContext(ctx context.Context) (int64, bool) {
	ownerID, ok := ctx.Value(ownerScopeContextKey{}).(int64)
	if !ok || ownerID <= 0 {
		return 0, false
	}
	return ownerID, true
}

// Admin is a recruiter/HR account that can sign in to the
// hire.voxire.com admin panel. Self-contained: not tied to the rtylr
// team/user system. Admins are managed from within the panel.
type Admin struct {
	ID           int64
	Email        string
	Name         string
	PasswordHash string
	CreatedAt    time.Time
}

// Opening is a job posting (read from the shared `careers` table) along
// with the count of applications received for it. Powers the position
// picker in the admin panel.
type Opening struct {
	ID             int64
	Slug           string
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
	IsActive       bool
	NewCount       int64
	TotalCount     int64
}

type OpeningPage struct {
	Openings []Opening
	Total    int64
	Limit    int32
	Offset   int32
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
	ID                 int64
	JobID              int64
	Name               string
	Email              string
	Phone              string
	Location           string
	SalaryExpectation  float64
	SalaryCurrency     string
	Status             string
	WorkHistory        string
	Education          string
	ExperienceSummary  string
	CoverLetter        string
	HasCV              bool
	CVProcessingStatus string
	CVProcessingError  string
	CVSummary          string
	CVSearchExcerpt    string
	CVProfile          CVProfile
	MatchSignal        MatchSignal
	CVProcessedAt      *time.Time
	ReviewRating       int
	ReviewNotes        string
	NextStep           string
	ReviewedAt         *time.Time
	CreatedAt          time.Time
}

// MatchSignal is a derived recruiter signal. It is intentionally computed
// from existing application, role, and CV fields so the admin can rank
// candidates without another migration or background job.
type MatchSignal struct {
	Score        int
	Label        string
	Summary      string
	Strengths    []string
	Risks        []string
	MatchedTerms []string
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
	Totals               AnalyticsTotals
	ApplicationsByStatus []AnalyticsBucket
	CVStatusMix          []AnalyticsBucket
	TaskStatusMix        []AnalyticsBucket
	ProfileFunctions     []AnalyticsBucket
	LeadServiceMix       []AnalyticsBucket
	LeadTimelineMix      []AnalyticsBucket
	RecentDays           []AnalyticsDay
	TopOpenings          []AnalyticsOpening
	RecentActivity       []AnalyticsActivity
}

type AnalyticsTotals struct {
	OpenJobs              int64
	ClosedJobs            int64
	ApplicationsTotal     int64
	Applications7d        int64
	Applications30d       int64
	NewApplications       int64
	LeadsTotal            int64
	Leads7d               int64
	Leads30d              int64
	AdminsTotal           int64
	TalentPoolTotal       int64
	TalentPool7d          int64
	TalentPool30d         int64
	CVProcessedTotal      int64
	CVPendingTotal        int64
	CVErrorTotal          int64
	CVRate                float64
	AvgSalaryExpectation  float64
	OpenTasks             int64
	OverdueTasks          int64
	ShortlistsTotal       int64
	ShortlistedCandidates int64
}

type AnalyticsBucket struct {
	Key   string
	Count int64
}

type AnalyticsDay struct {
	Date         string
	Applications int64
	Leads        int64
}

type AnalyticsOpening struct {
	ID                int64
	Title             string
	IsActive          bool
	TotalCount        int64
	NewCount          int64
	RecentCount       int64
	LastApplicationAt *time.Time
}

type AnalyticsActivity struct {
	Kind      string
	ID        int64
	Primary   string
	Secondary string
	Meta      string
	CreatedAt time.Time
}

// Lead is the admin-panel view of a company hiring inquiry submitted
// through the "Talk to us about a role" form (voxhire_company_leads).
type Lead struct {
	ID           int64
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
	Status       string
	CreatedAt    time.Time
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
	Profile       CVProfile
	Error         string
}

type CVProfile struct {
	Skills          []string
	Functions       []string
	Industries      []string
	RoleTitles      []string
	Seniority       string
	YearsExperience string
	Education       []string
	Languages       []string
	Certifications  []string
	Locations       []string
	Availability    string
	SalarySignals   []string
}

type CandidateFilter struct {
	Query    string
	Source   string
	CVStatus string
	Limit    int32
	Offset   int32
}

type Candidate struct {
	Source             string
	ID                 int64
	ApplicationID      int64
	JobID              int64
	OpeningTitle       string
	Name               string
	Email              string
	Phone              string
	Location           string
	Track              string
	CurrentRole        string
	DesiredRoles       string
	Status             string
	CVProcessingStatus string
	CVProcessingError  string
	CVSummary          string
	CVSearchExcerpt    string
	CVProfile          CVProfile
	MatchSignal        MatchSignal
	CreatedAt          time.Time
	CVProcessedAt      *time.Time
}

type CVJob struct {
	Source             string
	ID                 int64
	Name               string
	Email              string
	Role               string
	CVKey              string
	CVProcessingStatus string
	CVProcessingError  string
	CVSummary          string
	CVProfile          CVProfile
	CreatedAt          time.Time
	CVProcessedAt      *time.Time
}

type Task struct {
	ID             int64
	ApplicationID  int64
	JobID          int64
	OpeningTitle   string
	CandidateName  string
	CandidateEmail string
	Title          string
	Status         string
	DueAt          *time.Time
	CompletedAt    *time.Time
	CreatedAt      time.Time
}

type TaskInput struct {
	ID            int64
	ApplicationID int64
	Title         string
	Status        string
	DueAt         *time.Time
	// DueAtSet reports whether the request supplied due_at at all. A partial
	// update that omits it must keep the stored due date instead of clearing it.
	DueAtSet bool
}

type Shortlist struct {
	ID        int64
	JobID     int64
	JobTitle  string
	Name      string
	Status    string
	Notes     string
	Items     []ShortlistItem
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ShortlistItem struct {
	ID             int64
	ShortlistID    int64
	ApplicationID  int64
	CandidateName  string
	CandidateEmail string
	Status         string
	Rank           int
	Notes          string
	CVProfile      CVProfile
	MatchSignal    MatchSignal
	CreatedAt      time.Time
}

type ShortlistInput struct {
	ID     int64
	JobID  int64
	Name   string
	Status string
	Notes  string
}

type ShortlistItemInput struct {
	ShortlistID   int64
	ApplicationID int64
	Rank          int
	Notes         string
}

type TimelineEvent struct {
	Kind      string
	Title     string
	Detail    string
	CreatedAt time.Time
}
