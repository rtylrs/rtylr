package api

import (
	"time"

	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/domain"
)

type adminView struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func adminViewOf(value domain.Admin) adminView {
	return adminView{
		ID:        value.ID,
		Email:     value.Email,
		Name:      value.Name,
		CreatedAt: value.CreatedAt,
	}
}

type openingView struct {
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

func openingViewOf(value domain.Opening) openingView {
	return openingView{
		ID:             value.ID,
		Slug:           value.Slug,
		Title:          value.Title,
		Department:     value.Department,
		Location:       value.Location,
		EmploymentType: value.EmploymentType,
		WorkMode:       value.WorkMode,
		Seniority:      value.Seniority,
		Details:        value.Details,
		SalaryMin:      value.SalaryMin,
		SalaryMax:      value.SalaryMax,
		SalaryCurrency: value.SalaryCurrency,
		IsActive:       value.IsActive,
		NewCount:       value.NewCount,
		TotalCount:     value.TotalCount,
	}
}

type openingPageView struct {
	Openings []openingView `json:"openings"`
	Total    int64         `json:"total"`
	Limit    int32         `json:"limit"`
	Offset   int32         `json:"offset"`
}

func openingPageViewOf(value domain.OpeningPage) openingPageView {
	return openingPageView{
		Openings: mapViews(value.Openings, openingViewOf),
		Total:    value.Total,
		Limit:    value.Limit,
		Offset:   value.Offset,
	}
}

type applicationView struct {
	ID                 int64           `json:"id"`
	JobID              int64           `json:"job_id"`
	Name               string          `json:"name"`
	Email              string          `json:"email"`
	Phone              string          `json:"phone"`
	Location           string          `json:"location"`
	SalaryExpectation  float64         `json:"salary_expectation"`
	SalaryCurrency     string          `json:"salary_currency"`
	Status             string          `json:"status"`
	WorkHistory        string          `json:"work_history"`
	Education          string          `json:"education"`
	ExperienceSummary  string          `json:"experience_summary"`
	CoverLetter        string          `json:"cover_letter"`
	HasCV              bool            `json:"has_cv"`
	CVProcessingStatus string          `json:"cv_processing_status"`
	CVProcessingError  string          `json:"cv_processing_error"`
	CVSummary          string          `json:"cv_summary"`
	CVSearchExcerpt    string          `json:"cv_search_excerpt,omitempty"`
	CVProfile          cVProfileView   `json:"cv_profile,omitempty"`
	MatchSignal        matchSignalView `json:"match_signal,omitempty"`
	CVProcessedAt      *time.Time      `json:"cv_processed_at,omitempty"`
	ReviewRating       int             `json:"review_rating"`
	ReviewNotes        string          `json:"review_notes"`
	NextStep           string          `json:"next_step"`
	ReviewedAt         *time.Time      `json:"reviewed_at,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
}

func applicationViewOf(value domain.Application) applicationView {
	return applicationView{
		ID:                 value.ID,
		JobID:              value.JobID,
		Name:               value.Name,
		Email:              value.Email,
		Phone:              value.Phone,
		Location:           value.Location,
		SalaryExpectation:  value.SalaryExpectation,
		SalaryCurrency:     value.SalaryCurrency,
		Status:             value.Status,
		WorkHistory:        value.WorkHistory,
		Education:          value.Education,
		ExperienceSummary:  value.ExperienceSummary,
		CoverLetter:        value.CoverLetter,
		HasCV:              value.HasCV,
		CVProcessingStatus: value.CVProcessingStatus,
		CVProcessingError:  value.CVProcessingError,
		CVSummary:          value.CVSummary,
		CVSearchExcerpt:    value.CVSearchExcerpt,
		CVProfile:          cVProfileViewOf(value.CVProfile),
		MatchSignal:        matchSignalViewOf(value.MatchSignal),
		CVProcessedAt:      value.CVProcessedAt,
		ReviewRating:       value.ReviewRating,
		ReviewNotes:        value.ReviewNotes,
		NextStep:           value.NextStep,
		ReviewedAt:         value.ReviewedAt,
		CreatedAt:          value.CreatedAt,
	}
}

type matchSignalView struct {
	Score        int      `json:"score"`
	Label        string   `json:"label"`
	Summary      string   `json:"summary,omitempty"`
	Strengths    []string `json:"strengths,omitempty"`
	Risks        []string `json:"risks,omitempty"`
	MatchedTerms []string `json:"matched_terms,omitempty"`
}

func matchSignalViewOf(value domain.MatchSignal) matchSignalView {
	return matchSignalView{
		Score:        value.Score,
		Label:        value.Label,
		Summary:      value.Summary,
		Strengths:    value.Strengths,
		Risks:        value.Risks,
		MatchedTerms: value.MatchedTerms,
	}
}

type analyticsView struct {
	Totals               analyticsTotalsView     `json:"totals"`
	ApplicationsByStatus []analyticsBucketView   `json:"applications_by_status"`
	CVStatusMix          []analyticsBucketView   `json:"cv_status_mix"`
	TaskStatusMix        []analyticsBucketView   `json:"task_status_mix"`
	ProfileFunctions     []analyticsBucketView   `json:"profile_functions"`
	LeadServiceMix       []analyticsBucketView   `json:"lead_service_mix"`
	LeadTimelineMix      []analyticsBucketView   `json:"lead_timeline_mix"`
	RecentDays           []analyticsDayView      `json:"recent_days"`
	TopOpenings          []analyticsOpeningView  `json:"top_openings"`
	RecentActivity       []analyticsActivityView `json:"recent_activity"`
}

func analyticsViewOf(value domain.Analytics) analyticsView {
	return analyticsView{
		Totals:               analyticsTotalsViewOf(value.Totals),
		ApplicationsByStatus: mapViews(value.ApplicationsByStatus, analyticsBucketViewOf),
		CVStatusMix:          mapViews(value.CVStatusMix, analyticsBucketViewOf),
		TaskStatusMix:        mapViews(value.TaskStatusMix, analyticsBucketViewOf),
		ProfileFunctions:     mapViews(value.ProfileFunctions, analyticsBucketViewOf),
		LeadServiceMix:       mapViews(value.LeadServiceMix, analyticsBucketViewOf),
		LeadTimelineMix:      mapViews(value.LeadTimelineMix, analyticsBucketViewOf),
		RecentDays:           mapViews(value.RecentDays, analyticsDayViewOf),
		TopOpenings:          mapViews(value.TopOpenings, analyticsOpeningViewOf),
		RecentActivity:       mapViews(value.RecentActivity, analyticsActivityViewOf),
	}
}

type analyticsTotalsView struct {
	OpenJobs              int64   `json:"open_jobs"`
	ClosedJobs            int64   `json:"closed_jobs"`
	ApplicationsTotal     int64   `json:"applications_total"`
	Applications7d        int64   `json:"applications_7d"`
	Applications30d       int64   `json:"applications_30d"`
	NewApplications       int64   `json:"new_applications"`
	LeadsTotal            int64   `json:"leads_total"`
	Leads7d               int64   `json:"leads_7d"`
	Leads30d              int64   `json:"leads_30d"`
	AdminsTotal           int64   `json:"admins_total"`
	TalentPoolTotal       int64   `json:"talent_pool_total"`
	TalentPool7d          int64   `json:"talent_pool_7d"`
	TalentPool30d         int64   `json:"talent_pool_30d"`
	CVProcessedTotal      int64   `json:"cv_processed_total"`
	CVPendingTotal        int64   `json:"cv_pending_total"`
	CVErrorTotal          int64   `json:"cv_error_total"`
	CVRate                float64 `json:"cv_rate"`
	AvgSalaryExpectation  float64 `json:"avg_salary_expectation"`
	OpenTasks             int64   `json:"open_tasks"`
	OverdueTasks          int64   `json:"overdue_tasks"`
	ShortlistsTotal       int64   `json:"shortlists_total"`
	ShortlistedCandidates int64   `json:"shortlisted_candidates"`
}

func analyticsTotalsViewOf(value domain.AnalyticsTotals) analyticsTotalsView {
	return analyticsTotalsView{
		OpenJobs:              value.OpenJobs,
		ClosedJobs:            value.ClosedJobs,
		ApplicationsTotal:     value.ApplicationsTotal,
		Applications7d:        value.Applications7d,
		Applications30d:       value.Applications30d,
		NewApplications:       value.NewApplications,
		LeadsTotal:            value.LeadsTotal,
		Leads7d:               value.Leads7d,
		Leads30d:              value.Leads30d,
		AdminsTotal:           value.AdminsTotal,
		TalentPoolTotal:       value.TalentPoolTotal,
		TalentPool7d:          value.TalentPool7d,
		TalentPool30d:         value.TalentPool30d,
		CVProcessedTotal:      value.CVProcessedTotal,
		CVPendingTotal:        value.CVPendingTotal,
		CVErrorTotal:          value.CVErrorTotal,
		CVRate:                value.CVRate,
		AvgSalaryExpectation:  value.AvgSalaryExpectation,
		OpenTasks:             value.OpenTasks,
		OverdueTasks:          value.OverdueTasks,
		ShortlistsTotal:       value.ShortlistsTotal,
		ShortlistedCandidates: value.ShortlistedCandidates,
	}
}

type analyticsBucketView struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

func analyticsBucketViewOf(value domain.AnalyticsBucket) analyticsBucketView {
	return analyticsBucketView{
		Key:   value.Key,
		Count: value.Count,
	}
}

type analyticsDayView struct {
	Date         string `json:"date"`
	Applications int64  `json:"applications"`
	Leads        int64  `json:"leads"`
}

func analyticsDayViewOf(value domain.AnalyticsDay) analyticsDayView {
	return analyticsDayView{
		Date:         value.Date,
		Applications: value.Applications,
		Leads:        value.Leads,
	}
}

type analyticsOpeningView struct {
	ID                int64      `json:"id"`
	Title             string     `json:"title"`
	IsActive          bool       `json:"is_active"`
	TotalCount        int64      `json:"total_count"`
	NewCount          int64      `json:"new_count"`
	RecentCount       int64      `json:"recent_count"`
	LastApplicationAt *time.Time `json:"last_application_at,omitempty"`
}

func analyticsOpeningViewOf(value domain.AnalyticsOpening) analyticsOpeningView {
	return analyticsOpeningView{
		ID:                value.ID,
		Title:             value.Title,
		IsActive:          value.IsActive,
		TotalCount:        value.TotalCount,
		NewCount:          value.NewCount,
		RecentCount:       value.RecentCount,
		LastApplicationAt: value.LastApplicationAt,
	}
}

type analyticsActivityView struct {
	Kind      string    `json:"kind"`
	ID        int64     `json:"id"`
	Primary   string    `json:"primary"`
	Secondary string    `json:"secondary"`
	Meta      string    `json:"meta"`
	CreatedAt time.Time `json:"created_at"`
}

func analyticsActivityViewOf(value domain.AnalyticsActivity) analyticsActivityView {
	return analyticsActivityView{
		Kind:      value.Kind,
		ID:        value.ID,
		Primary:   value.Primary,
		Secondary: value.Secondary,
		Meta:      value.Meta,
		CreatedAt: value.CreatedAt,
	}
}

type leadView struct {
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

func leadViewOf(value domain.Lead) leadView {
	return leadView{
		ID:           value.ID,
		CompanyName:  value.CompanyName,
		ContactName:  value.ContactName,
		ContactTitle: value.ContactTitle,
		WorkEmail:    value.WorkEmail,
		Phone:        value.Phone,
		Role:         value.Role,
		Headcount:    value.Headcount,
		Timeline:     value.Timeline,
		Service:      value.Service,
		Message:      value.Message,
		AgencyOptIn:  value.AgencyOptIn,
		Status:       value.Status,
		CreatedAt:    value.CreatedAt,
	}
}

type cVProfileView struct {
	Skills          []string `json:"skills,omitempty"`
	Functions       []string `json:"functions,omitempty"`
	Industries      []string `json:"industries,omitempty"`
	RoleTitles      []string `json:"role_titles,omitempty"`
	Seniority       string   `json:"seniority,omitempty"`
	YearsExperience string   `json:"years_experience,omitempty"`
	Education       []string `json:"education,omitempty"`
	Languages       []string `json:"languages,omitempty"`
	Certifications  []string `json:"certifications,omitempty"`
	Locations       []string `json:"locations,omitempty"`
	Availability    string   `json:"availability,omitempty"`
	SalarySignals   []string `json:"salary_signals,omitempty"`
}

func cVProfileViewOf(value domain.CVProfile) cVProfileView {
	return cVProfileView{
		Skills:          value.Skills,
		Functions:       value.Functions,
		Industries:      value.Industries,
		RoleTitles:      value.RoleTitles,
		Seniority:       value.Seniority,
		YearsExperience: value.YearsExperience,
		Education:       value.Education,
		Languages:       value.Languages,
		Certifications:  value.Certifications,
		Locations:       value.Locations,
		Availability:    value.Availability,
		SalarySignals:   value.SalarySignals,
	}
}

type candidateView struct {
	Source             string          `json:"source"`
	ID                 int64           `json:"id"`
	ApplicationID      int64           `json:"application_id,omitempty"`
	JobID              int64           `json:"job_id,omitempty"`
	OpeningTitle       string          `json:"opening_title,omitempty"`
	Name               string          `json:"name"`
	Email              string          `json:"email"`
	Phone              string          `json:"phone,omitempty"`
	Location           string          `json:"location,omitempty"`
	Track              string          `json:"track,omitempty"`
	CurrentRole        string          `json:"current_role,omitempty"`
	DesiredRoles       string          `json:"desired_roles,omitempty"`
	Status             string          `json:"status"`
	CVProcessingStatus string          `json:"cv_processing_status"`
	CVProcessingError  string          `json:"cv_processing_error,omitempty"`
	CVSummary          string          `json:"cv_summary,omitempty"`
	CVSearchExcerpt    string          `json:"cv_search_excerpt,omitempty"`
	CVProfile          cVProfileView   `json:"cv_profile,omitempty"`
	MatchSignal        matchSignalView `json:"match_signal,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	CVProcessedAt      *time.Time      `json:"cv_processed_at,omitempty"`
}

func candidateViewOf(value domain.Candidate) candidateView {
	return candidateView{
		Source:             value.Source,
		ID:                 value.ID,
		ApplicationID:      value.ApplicationID,
		JobID:              value.JobID,
		OpeningTitle:       value.OpeningTitle,
		Name:               value.Name,
		Email:              value.Email,
		Phone:              value.Phone,
		Location:           value.Location,
		Track:              value.Track,
		CurrentRole:        value.CurrentRole,
		DesiredRoles:       value.DesiredRoles,
		Status:             value.Status,
		CVProcessingStatus: value.CVProcessingStatus,
		CVProcessingError:  value.CVProcessingError,
		CVSummary:          value.CVSummary,
		CVSearchExcerpt:    value.CVSearchExcerpt,
		CVProfile:          cVProfileViewOf(value.CVProfile),
		MatchSignal:        matchSignalViewOf(value.MatchSignal),
		CreatedAt:          value.CreatedAt,
		CVProcessedAt:      value.CVProcessedAt,
	}
}

type cVJobView struct {
	Source             string        `json:"source"`
	ID                 int64         `json:"id"`
	Name               string        `json:"name"`
	Email              string        `json:"email"`
	Role               string        `json:"role"`
	CVKey              string        `json:"cv_key"`
	CVProcessingStatus string        `json:"cv_processing_status"`
	CVProcessingError  string        `json:"cv_processing_error,omitempty"`
	CVSummary          string        `json:"cv_summary,omitempty"`
	CVProfile          cVProfileView `json:"cv_profile,omitempty"`
	CreatedAt          time.Time     `json:"created_at"`
	CVProcessedAt      *time.Time    `json:"cv_processed_at,omitempty"`
}

func cVJobViewOf(value domain.CVJob) cVJobView {
	return cVJobView{
		Source:             value.Source,
		ID:                 value.ID,
		Name:               value.Name,
		Email:              value.Email,
		Role:               value.Role,
		CVKey:              value.CVKey,
		CVProcessingStatus: value.CVProcessingStatus,
		CVProcessingError:  value.CVProcessingError,
		CVSummary:          value.CVSummary,
		CVProfile:          cVProfileViewOf(value.CVProfile),
		CreatedAt:          value.CreatedAt,
		CVProcessedAt:      value.CVProcessedAt,
	}
}

type taskView struct {
	ID             int64      `json:"id"`
	ApplicationID  int64      `json:"application_id"`
	JobID          int64      `json:"job_id"`
	OpeningTitle   string     `json:"opening_title"`
	CandidateName  string     `json:"candidate_name"`
	CandidateEmail string     `json:"candidate_email"`
	Title          string     `json:"title"`
	Status         string     `json:"status"`
	DueAt          *time.Time `json:"due_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

func taskViewOf(value domain.Task) taskView {
	return taskView{
		ID:             value.ID,
		ApplicationID:  value.ApplicationID,
		JobID:          value.JobID,
		OpeningTitle:   value.OpeningTitle,
		CandidateName:  value.CandidateName,
		CandidateEmail: value.CandidateEmail,
		Title:          value.Title,
		Status:         value.Status,
		DueAt:          value.DueAt,
		CompletedAt:    value.CompletedAt,
		CreatedAt:      value.CreatedAt,
	}
}

type shortlistView struct {
	ID        int64               `json:"id"`
	JobID     int64               `json:"job_id"`
	JobTitle  string              `json:"job_title"`
	Name      string              `json:"name"`
	Status    string              `json:"status"`
	Notes     string              `json:"notes"`
	Items     []shortlistItemView `json:"items,omitempty"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

func shortlistViewOf(value domain.Shortlist) shortlistView {
	return shortlistView{
		ID:        value.ID,
		JobID:     value.JobID,
		JobTitle:  value.JobTitle,
		Name:      value.Name,
		Status:    value.Status,
		Notes:     value.Notes,
		Items:     mapViews(value.Items, shortlistItemViewOf),
		CreatedAt: value.CreatedAt,
		UpdatedAt: value.UpdatedAt,
	}
}

type shortlistItemView struct {
	ID             int64           `json:"id"`
	ShortlistID    int64           `json:"shortlist_id"`
	ApplicationID  int64           `json:"application_id"`
	CandidateName  string          `json:"candidate_name"`
	CandidateEmail string          `json:"candidate_email"`
	Status         string          `json:"status"`
	Rank           int             `json:"rank"`
	Notes          string          `json:"notes"`
	CVProfile      cVProfileView   `json:"cv_profile,omitempty"`
	MatchSignal    matchSignalView `json:"match_signal,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

func shortlistItemViewOf(value domain.ShortlistItem) shortlistItemView {
	return shortlistItemView{
		ID:             value.ID,
		ShortlistID:    value.ShortlistID,
		ApplicationID:  value.ApplicationID,
		CandidateName:  value.CandidateName,
		CandidateEmail: value.CandidateEmail,
		Status:         value.Status,
		Rank:           value.Rank,
		Notes:          value.Notes,
		CVProfile:      cVProfileViewOf(value.CVProfile),
		MatchSignal:    matchSignalViewOf(value.MatchSignal),
		CreatedAt:      value.CreatedAt,
	}
}

type timelineEventView struct {
	Kind      string    `json:"kind"`
	Title     string    `json:"title"`
	Detail    string    `json:"detail,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func timelineEventViewOf(value domain.TimelineEvent) timelineEventView {
	return timelineEventView{
		Kind:      value.Kind,
		Title:     value.Title,
		Detail:    value.Detail,
		CreatedAt: value.CreatedAt,
	}
}

func mapViews[T, V any](values []T, convert func(T) V) []V {
	if values == nil {
		return nil
	}
	result := make([]V, len(values))
	for i, value := range values {
		result[i] = convert(value)
	}
	return result
}

func adminViewPointer(value *domain.Admin) *adminView {
	if value == nil {
		return nil
	}
	result := adminViewOf(*value)
	return &result
}

func analyticsViewPointer(value *domain.Analytics) *analyticsView {
	if value == nil {
		return nil
	}
	result := analyticsViewOf(*value)
	return &result
}

func openingViewPointer(value *domain.Opening) *openingView {
	if value == nil {
		return nil
	}
	result := openingViewOf(*value)
	return &result
}

func taskViewPointer(value *domain.Task) *taskView {
	if value == nil {
		return nil
	}
	result := taskViewOf(*value)
	return &result
}

func shortlistViewPointer(value *domain.Shortlist) *shortlistView {
	if value == nil {
		return nil
	}
	result := shortlistViewOf(*value)
	return &result
}

func shortlistItemViewPointer(value *domain.ShortlistItem) *shortlistItemView {
	if value == nil {
		return nil
	}
	result := shortlistItemViewOf(*value)
	return &result
}
