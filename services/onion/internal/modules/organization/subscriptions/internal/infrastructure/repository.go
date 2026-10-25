package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/voxire/onion/internal/modules/organization/subscriptions/internal/domain"
	"github.com/voxire/onion/internal/modules/organization/subscriptions/internal/infrastructure/persistence/sqlc"
)

type Repository struct {
	queries *sqlc.Queries
	db      *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		db:      db,
	}
}

// ── Plans ─────────────────────────────────────────────────────────────────────

func hydratePlanJSON(p *domain.Plan, appsJSON, featuresJSON string, maxSeats, maxLocations sql.NullInt64) {
	if jsonErr := json.Unmarshal([]byte(appsJSON), &p.Apps); jsonErr != nil {
		p.Apps = []string{}
	}
	if jsonErr := json.Unmarshal([]byte(featuresJSON), &p.Features); jsonErr != nil {
		p.Features = []string{}
	}
	if maxSeats.Valid {
		v := int(maxSeats.Int64)
		p.MaxSeats = &v
	}
	if maxLocations.Valid {
		v := int(maxLocations.Int64)
		p.MaxLocations = &v
	}
}

func planFromSlugRow(row sqlc.GetPlanBySlugRow) domain.Plan {
	return planFromFields(
		row.ID, row.Name, row.Slug, row.Apps, row.PricePerSeatCents, row.AnnualPricePerSeatCents,
		row.TrialDays, row.PaddlePriceID, row.Tagline, row.MaxSeats, row.MaxLocations,
		row.AiCreditsPerMonth, row.Features, row.TierOrder, row.IsSolo, row.IsFeatured,
		row.IsActive, row.CreatedAt,
	)
}

func planFromIDRow(row sqlc.GetPlanByIDRow) domain.Plan {
	return planFromFields(
		row.ID, row.Name, row.Slug, row.Apps, row.PricePerSeatCents, row.AnnualPricePerSeatCents,
		row.TrialDays, row.PaddlePriceID, row.Tagline, row.MaxSeats, row.MaxLocations,
		row.AiCreditsPerMonth, row.Features, row.TierOrder, row.IsSolo, row.IsFeatured,
		row.IsActive, row.CreatedAt,
	)
}

func planFromListRow(row sqlc.ListPlansRow) domain.Plan {
	return planFromFields(
		row.ID, row.Name, row.Slug, row.Apps, row.PricePerSeatCents, row.AnnualPricePerSeatCents,
		row.TrialDays, row.PaddlePriceID, row.Tagline, row.MaxSeats, row.MaxLocations,
		row.AiCreditsPerMonth, row.Features, row.TierOrder, row.IsSolo, row.IsFeatured,
		row.IsActive, row.CreatedAt,
	)
}

func planFromFields(
	id int64, name, slug string, apps json.RawMessage,
	pricePerSeatCents, annualPricePerSeatCents, trialDays int32,
	paddlePriceID, tagline string,
	maxSeats, maxLocations sql.NullInt32,
	aiCreditsPerMonth int32, features json.RawMessage,
	tierOrder int32, isSolo, isFeatured bool,
	isActive sql.NullBool, createdAt sql.NullTime,
) domain.Plan {
	var p domain.Plan
	p.ID = id
	p.Name = name
	p.Slug = slug
	p.PricePerSeatCents = int(pricePerSeatCents)
	p.AnnualPricePerSeatCents = int(annualPricePerSeatCents)
	p.TrialDays = int(trialDays)
	p.PaddlePriceID = paddlePriceID
	p.Tagline = tagline
	p.AICreditsPerMonth = int(aiCreditsPerMonth)
	p.TierOrder = int(tierOrder)
	p.IsSolo = isSolo
	p.IsFeatured = isFeatured
	if isActive.Valid {
		p.IsActive = isActive.Bool
	}
	if createdAt.Valid {
		p.CreatedAt = createdAt.Time
	}
	var maxSeatsNull, maxLocationsNull sql.NullInt64
	if maxSeats.Valid {
		maxSeatsNull = sql.NullInt64{Int64: int64(maxSeats.Int32), Valid: true}
	}
	if maxLocations.Valid {
		maxLocationsNull = sql.NullInt64{Int64: int64(maxLocations.Int32), Valid: true}
	}
	hydratePlanJSON(&p, string(apps), string(features), maxSeatsNull, maxLocationsNull)
	return p
}

func (r *Repository) ListPlans(ctx context.Context) ([]domain.Plan, error) {
	rows, err := r.queries.ListPlans(ctx)
	if err != nil {
		return nil, err
	}

	plans := make([]domain.Plan, 0, len(rows))
	for _, row := range rows {
		plans = append(plans, planFromListRow(row))
	}
	return plans, nil
}

func (r *Repository) GetPlanBySlug(ctx context.Context, slug string) (*domain.Plan, error) {
	row, err := r.queries.GetPlanBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPlanNotFound
		}
		return nil, err
	}
	p := planFromSlugRow(row)
	return &p, nil
}

func (r *Repository) GetPlanByID(ctx context.Context, id int64) (*domain.Plan, error) {
	row, err := r.queries.GetPlanByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPlanNotFound
		}
		return nil, err
	}
	p := planFromIDRow(row)
	return &p, nil
}

// ── Subscriptions ─────────────────────────────────────────────────────────────

func (r *Repository) GetSubscription(ctx context.Context, businessID int64) (*domain.Subscription, error) {
	row, err := r.queries.GetSubscription(ctx, businessID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSubscriptionNotFound
		}
		return nil, err
	}

	sub := &domain.Subscription{
		ID:                row.ID,
		BusinessID:        row.BusinessID,
		PlanID:            row.PlanID,
		PlanName:          row.PlanName,
		PlanSlug:          row.PlanSlug,
		AICreditsPerMonth: int(row.AiCreditsPerMonth),
		PricePerSeatCents: int(row.PricePerSeatCents),
		SeatCount:         int(row.SeatCount),
	}
	if row.Status.Valid {
		sub.Status = string(row.Status.SubscriptionsStatus)
	}
	if jsonErr := json.Unmarshal(row.Apps, &sub.Apps); jsonErr != nil {
		sub.Apps = []string{}
	}
	if row.TrialEndsAt.Valid {
		t := row.TrialEndsAt.Time
		sub.TrialEndsAt = &t
	}
	if row.NextBillingDate.Valid {
		t := row.NextBillingDate.Time
		sub.NextBillingDate = &t
	}
	if row.CurrentPeriodEndsAt.Valid {
		t := row.CurrentPeriodEndsAt.Time
		sub.CurrentPeriodEndsAt = &t
	}
	if row.PaddleSubscriptionID.Valid {
		sub.PaddleSubscriptionID = row.PaddleSubscriptionID.String
	}
	if row.PaddleCustomerID.Valid {
		sub.PaddleCustomerID = row.PaddleCustomerID.String
	}
	if row.CreatedAt.Valid {
		sub.CreatedAt = row.CreatedAt.Time
	}

	if err := r.hydrateActiveGrant(ctx, sub); err != nil {
		return nil, err
	}

	return sub, nil
}

func (r *Repository) hydrateActiveGrant(ctx context.Context, sub *domain.Subscription) error {
	expiresAt, err := r.queries.GetActiveManualGrantExpiresAt(ctx, sqlc.GetActiveManualGrantExpiresAtParams{
		BusinessID: sub.BusinessID,
		PlanID:     sub.PlanID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	sub.IsGranted = true
	if expiresAt.Valid {
		t := expiresAt.Time
		sub.GrantExpiresAt = &t
	}
	return nil
}
