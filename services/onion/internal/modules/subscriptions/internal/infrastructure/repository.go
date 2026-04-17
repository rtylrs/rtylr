package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/voxire/onion/internal/modules/subscriptions/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{db: dbclient.Pool()}
}

func (r *Repository) ListPlans(ctx context.Context) ([]domain.Plan, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, slug, apps, is_active, created_at FROM plans WHERE is_active = TRUE ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	plans := make([]domain.Plan, 0)
	for rows.Next() {
		var p domain.Plan
		var appsJSON string
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &appsJSON, &p.IsActive, &p.CreatedAt); err != nil {
			return nil, err
		}
		if jsonErr := json.Unmarshal([]byte(appsJSON), &p.Apps); jsonErr != nil {
			p.Apps = []string{}
		}
		plans = append(plans, p)
	}

	return plans, rows.Err()
}

func (r *Repository) GetPlanBySlug(ctx context.Context, slug string) (*domain.Plan, error) {
	var p domain.Plan
	var appsJSON string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, slug, apps, is_active, created_at FROM plans WHERE slug = ? AND is_active = TRUE`,
		slug).Scan(&p.ID, &p.Name, &p.Slug, &appsJSON, &p.IsActive, &p.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrPlanNotFound
		}

		return nil, err
	}

	if jsonErr := json.Unmarshal([]byte(appsJSON), &p.Apps); jsonErr != nil {
		p.Apps = []string{}
	}

	return &p, nil
}

func (r *Repository) GetPlanByID(ctx context.Context, id int64) (*domain.Plan, error) {
	var p domain.Plan
	var appsJSON string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, slug, apps, is_active, created_at FROM plans WHERE id = ?`,
		id).Scan(&p.ID, &p.Name, &p.Slug, &appsJSON, &p.IsActive, &p.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrPlanNotFound
		}

		return nil, err
	}

	if jsonErr := json.Unmarshal([]byte(appsJSON), &p.Apps); jsonErr != nil {
		p.Apps = []string{}
	}

	return &p, nil
}

func (r *Repository) GetSubscription(ctx context.Context, ownerID int64) (*domain.Subscription, error) {
	var sub domain.Subscription
	var appsJSON string
	var trialEndsAt, periodEndsAt sql.NullTime
	err := r.db.QueryRowContext(ctx,
		`SELECT s.id, s.owner_id, s.plan_id, p.name, p.slug, p.apps, s.status, s.trial_ends_at, s.current_period_ends_at, s.created_at
		 FROM subscriptions s
		 JOIN plans p ON p.id = s.plan_id
		 WHERE s.owner_id = ?`,
		ownerID).Scan(
		&sub.ID, &sub.OwnerID, &sub.PlanID, &sub.PlanName, &sub.PlanSlug, &appsJSON,
		&sub.Status, &trialEndsAt, &periodEndsAt, &sub.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrSubscriptionNotFound
		}

		return nil, err
	}

	if jsonErr := json.Unmarshal([]byte(appsJSON), &sub.Apps); jsonErr != nil {
		sub.Apps = []string{}
	}

	if trialEndsAt.Valid {
		t := trialEndsAt.Time
		sub.TrialEndsAt = &t
	}

	if periodEndsAt.Valid {
		t := periodEndsAt.Time
		sub.CurrentPeriodEndsAt = &t
	}

	return &sub, nil
}

func (r *Repository) UpsertSubscription(ctx context.Context, ownerID, planID int64) error {
	trialEnd := time.Now().AddDate(0, 0, 14)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO subscriptions (owner_id, plan_id, status, trial_ends_at)
		 VALUES (?, ?, 'trial', ?)
		 ON DUPLICATE KEY UPDATE plan_id = VALUES(plan_id), updated_at = CURRENT_TIMESTAMP`,
		ownerID, planID, trialEnd)

	return err
}

func (r *Repository) GetOwnerApps(ctx context.Context, ownerID int64) ([]string, error) {
	var appsJSON string
	err := r.db.QueryRowContext(ctx,
		`SELECT p.apps FROM subscriptions s JOIN plans p ON p.id = s.plan_id WHERE s.owner_id = ?`,
		ownerID).Scan(&appsJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrSubscriptionNotFound
		}

		return nil, err
	}

	var apps []string
	if jsonErr := json.Unmarshal([]byte(appsJSON), &apps); jsonErr != nil {
		return []string{}, nil
	}

	return apps, nil
}
