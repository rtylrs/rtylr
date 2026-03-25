package domain

import (
	"errors"
	"time"

	"github.com/voxire/onion/internal/platform/utils/convert"
	sqlc "github.com/voxire/onion/internal/platforms/zontriq/users/internal/infrastructure/persistence/sqlc"
)

var (
	ErrZontriqUserNotFound = errors.New("zontriq user not found")
	ErrPlanNotFound        = errors.New("plan not found")
	ErrUsageNotFound       = errors.New("usage not found")
	ErrPreferencesNotFound = errors.New("preferences not found")
)

type ZontriqUser struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	PlanID      int       `json:"plan_id"`
	Plan        Plan      `json:"plan"`
	Status      int       `json:"status"`
	Onboarded   bool      `json:"onboarded"`
	TrialEndsAt time.Time `json:"trial_ends_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (z *ZontriqUser) FromRows(row sqlc.GetZontriqUserByUserIDRow) error {
	z.ID = int(row.ID)
	z.UserID = int(row.UserID)
	z.Email = row.UserEmail
	z.Name = convert.ToString(row.UserName)
	z.PlanID = int(row.PlanID)
	z.Plan = Plan{
		ID:                int(row.PlanPk),
		Name:              row.PlanName,
		DisplayName:       row.PlanDisplayName,
		PriceCents:        int(row.PlanPriceCents),
		MaxVideosPerMonth: int(row.PlanMaxVideosPerMonth),
		MaxChannels:       int(row.PlanMaxChannels),
		MaxSeries:         int(row.PlanMaxSeries),
		MaxAutomations:    int(row.PlanMaxAutomations),
	}
	z.Status = int(row.Status)
	z.Onboarded = row.Onboarded
	z.TrialEndsAt = row.TrialEndsAt.Time
	z.CreatedAt = row.CreatedAt.Time
	z.UpdatedAt = row.UpdatedAt.Time

	return nil
}

type Plan struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	DisplayName       string `json:"display_name"`
	PriceCents        int    `json:"price_cents"`
	MaxVideosPerMonth int    `json:"max_videos_per_month"`
	MaxChannels       int    `json:"max_channels"`
	MaxSeries         int    `json:"max_series"`
	MaxAutomations    int    `json:"max_automations"`
}

func (p *Plan) FromRows(row sqlc.ZontriqPlan) error {
	p.ID = int(row.ID)
	p.Name = row.Name
	p.DisplayName = row.DisplayName
	p.PriceCents = int(row.PriceCents)
	p.MaxVideosPerMonth = int(row.MaxVideosPerMonth)
	p.MaxChannels = int(row.MaxChannels)
	p.MaxSeries = int(row.MaxSeries)
	p.MaxAutomations = int(row.MaxAutomations)

	return nil
}

type Usage struct {
	ID              int       `json:"id"`
	ZontriqUserID   int       `json:"zontriq_user_id"`
	PeriodStart     time.Time `json:"period_start"`
	PeriodEnd       time.Time `json:"period_end"`
	VideosGenerated int       `json:"videos_generated"`
	VideosPublished int       `json:"videos_published"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (u *Usage) FromRows(row sqlc.ZontriqUsage) error {
	u.ID = int(row.ID)
	u.ZontriqUserID = int(row.ZontriqUserID)
	u.PeriodStart = row.PeriodStart.Time
	u.PeriodEnd = row.PeriodEnd.Time
	u.VideosGenerated = int(row.VideosGenerated)
	u.VideosPublished = int(row.VideosPublished)
	u.CreatedAt = row.CreatedAt.Time
	u.UpdatedAt = row.UpdatedAt.Time

	return nil
}

type Preferences struct {
	ID                   int    `json:"id"`
	ZontriqUserID        int    `json:"zontriq_user_id"`
	DefaultVoice         string `json:"default_voice"`
	DefaultNiche         string `json:"default_niche"`
	NotifyVideoReady     bool   `json:"notify_video_ready"`
	NotifyPublishSuccess bool   `json:"notify_publish_success"`
	NotifyPublishFailure bool   `json:"notify_publish_failure"`
	NotifyQuotaWarning   bool   `json:"notify_quota_warning"`
	NotifyWeeklyReport   bool   `json:"notify_weekly_report"`
}

func (p *Preferences) FromRows(row sqlc.ZontriqPreference) error {
	p.ID = int(row.ID)
	p.ZontriqUserID = int(row.ZontriqUserID)
	p.DefaultVoice = convert.ToString(row.DefaultVoice)
	p.DefaultNiche = convert.ToString(row.DefaultNiche)
	p.NotifyVideoReady = row.NotifyVideoReady
	p.NotifyPublishSuccess = row.NotifyPublishSuccess
	p.NotifyPublishFailure = row.NotifyPublishFailure
	p.NotifyQuotaWarning = row.NotifyQuotaWarning
	p.NotifyWeeklyReport = row.NotifyWeeklyReport

	return nil
}
