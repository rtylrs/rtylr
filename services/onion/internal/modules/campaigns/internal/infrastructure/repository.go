package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/campaigns/internal/domain"
	"github.com/voxire/onion/internal/modules/campaigns/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
	}
}

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.Campaign, error) {
	rows, err := r.queries.ListCampaigns(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list campaigns", logger.Error(err))

		return nil, err
	}

	campaigns := make([]domain.Campaign, 0, len(rows))

	for _, row := range rows {
		campaigns = append(campaigns, campaignFromRow(row))
	}

	return campaigns, nil
}

func (r *Repository) ListByStatus(ctx context.Context, businessID int64, status string) ([]domain.Campaign, error) {
	rows, err := r.queries.ListCampaignsByStatus(ctx, sqlc.ListCampaignsByStatusParams{
		BusinessID: businessID,
		Status:     sqlc.NullCrmCampaignsStatus{CrmCampaignsStatus: sqlc.CrmCampaignsStatus(status), Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list campaigns by status", logger.Error(err))

		return nil, err
	}

	campaigns := make([]domain.Campaign, 0, len(rows))
	for _, row := range rows {
		campaigns = append(campaigns, campaignFromRow(row))
	}

	return campaigns, nil
}

// ListPaginated returns a single page of campaigns for the business, ordered
// by creation date. Limit and offset are applied as-is; callers clamp them.
func (r *Repository) ListPaginated(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Campaign, error) {
	rows, err := r.queries.ListCampaignsPaginated(ctx, sqlc.ListCampaignsPaginatedParams{
		BusinessID: businessID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list campaigns page", logger.Error(err))

		return nil, err
	}

	campaigns := make([]domain.Campaign, 0, len(rows))

	for _, row := range rows {
		campaigns = append(campaigns, campaignFromRow(row))
	}

	return campaigns, nil
}

// Count returns the total number of campaigns for the business.
func (r *Repository) Count(ctx context.Context, businessID int64) (int64, error) {
	count, err := r.queries.CountCampaigns(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count campaigns", logger.Error(err))

		return 0, err
	}

	return count, nil
}

func (r *Repository) GetByID(ctx context.Context, businessID, id int64) (domain.Campaign, error) {
	row, err := r.queries.GetCampaignByID(ctx, sqlc.GetCampaignByIDParams{
		ID:         id,
		BusinessID: businessID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Campaign{}, domain.ErrCampaignNotFound
		}

		return domain.Campaign{}, err
	}

	return campaignFromRow(row), nil
}

func (r *Repository) Create(ctx context.Context, input domain.CampaignInput) (int64, error) {
	var segmentID sql.NullInt64
	if input.AudienceSegmentID != nil {
		segmentID = sql.NullInt64{Int64: *input.AudienceSegmentID, Valid: true}
	}

	result, err := r.queries.CreateCampaign(ctx, sqlc.CreateCampaignParams{
		OwnerID:     input.OwnerID,
		BusinessID:  input.BusinessID,
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Channel:     sqlc.CrmCampaignsChannel(input.Type),
		SegmentID:   segmentID,
		Subject:     sql.NullString{String: input.Subject, Valid: input.Subject != ""},
		Content:     input.Body,
		ScheduledAt: nullableTime(input.ScheduledAt),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create campaign", logger.Error(err))

		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.CampaignUpdateInput) error {
	businessID := input.BusinessID
	if businessID == 0 {
		businessID = input.OwnerID
	}

	var segmentID sql.NullInt64
	if input.AudienceSegmentID != nil {
		segmentID = sql.NullInt64{Int64: *input.AudienceSegmentID, Valid: true}
	}

	result, err := r.queries.UpdateCampaign(ctx, sqlc.UpdateCampaignParams{
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Channel:     sqlc.CrmCampaignsChannel(input.Type),
		SegmentID:   segmentID,
		Subject:     sql.NullString{String: input.Subject, Valid: input.Subject != ""},
		Content:     input.Body,
		ScheduledAt: nullableTime(input.ScheduledAt),
		ID:          input.ID,
		BusinessID:  businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update campaign", logger.Error(err))

		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		// MySQL reports zero changed rows when the UPDATE matched a row whose
		// values are already identical (an idempotent re-send). Only a campaign
		// that does not exist for this business is a genuine not-found.
		if _, err := r.GetByID(ctx, businessID, input.ID); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) UpdateStatus(ctx context.Context, businessID, id int64, status string) error {
	result, err := r.queries.UpdateCampaignStatus(ctx, sqlc.UpdateCampaignStatusParams{
		Status:     sqlc.NullCrmCampaignsStatus{CrmCampaignsStatus: sqlc.CrmCampaignsStatus(status), Valid: true},
		ID:         id,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update campaign status", logger.Error(err))

		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		// Same as Update: an unchanged status is an idempotent success, not a
		// missing campaign.
		if _, err := r.GetByID(ctx, businessID, id); err != nil {
			return err
		}
	}

	return nil
}

func campaignFromRow(row sqlc.CrmCampaign) domain.Campaign {
	c := domain.Campaign{
		ID:             row.ID,
		OwnerID:        row.OwnerID,
		BusinessID:     row.BusinessID,
		Name:           row.Name,
		Description:    row.Description.String,
		Type:           string(row.Channel),
		Status:         string(row.Status.CrmCampaignsStatus),
		Subject:        row.Subject.String,
		Body:           row.Content,
		AudienceCount:  int(row.AudienceCount.Int32),
		SentCount:      int(row.SentCount.Int32),
		DeliveredCount: int(row.DeliveredCount.Int32),
		OpenCount:      int(row.OpenedCount.Int32),
		ClickedCount:   int(row.ClickedCount.Int32),
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
	}

	if row.SegmentID.Valid {
		id := row.SegmentID.Int64
		c.AudienceSegmentID = &id
	}

	if row.ScheduledAt.Valid {
		t := row.ScheduledAt.Time
		c.ScheduledAt = &t
	}

	if row.SentAt.Valid {
		t := row.SentAt.Time
		c.SentAt = &t
	}

	return c
}

func nullableTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}

	return sql.NullTime{Time: *t, Valid: true}
}

func (r *Repository) Delete(ctx context.Context, businessID, id int64) error {
	return r.queries.DeleteCampaign(ctx, sqlc.DeleteCampaignParams{
		ID:         id,
		BusinessID: businessID,
	})
}
