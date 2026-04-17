package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrPlanNotFound         = errors.New("plan not found")
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

type Plan struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Apps      []string  `json:"apps"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type Subscription struct {
	ID                  int64      `json:"id"`
	OwnerID             int64      `json:"owner_id"`
	PlanID              int64      `json:"plan_id"`
	PlanName            string     `json:"plan_name"`
	PlanSlug            string     `json:"plan_slug"`
	Apps                []string   `json:"apps"`
	Status              string     `json:"status"`
	TrialEndsAt         *time.Time `json:"trial_ends_at,omitempty"`
	CurrentPeriodEndsAt *time.Time `json:"current_period_ends_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
}

type Repository interface {
	ListPlans(ctx context.Context) ([]Plan, error)
	GetPlanBySlug(ctx context.Context, slug string) (*Plan, error)
	GetPlanByID(ctx context.Context, id int64) (*Plan, error)
	GetSubscription(ctx context.Context, ownerID int64) (*Subscription, error)
	UpsertSubscription(ctx context.Context, ownerID, planID int64) error
	GetOwnerApps(ctx context.Context, ownerID int64) ([]string, error)
}
