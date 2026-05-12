package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/campaigns/internal/domain"
	"github.com/voxire/onion/internal/modules/campaigns/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		queries: sqlc.New(dbclient.Pool()),
	}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Campaign, error) {
	rows, err := r.queries.ListCampaigns(ctx, ownerID)
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

func (r *Repository) GetByID(ctx context.Context, ownerID, id int64) (domain.Campaign, error) {
	row, err := r.queries.GetCampaignByID(ctx, sqlc.GetCampaignByIDParams{
		ID:      id,
		OwnerID: ownerID,
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
		Name:        input.Name,
		Description: sql.NullString{},
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
	var segmentID sql.NullInt64
	if input.AudienceSegmentID != nil {
		segmentID = sql.NullInt64{Int64: *input.AudienceSegmentID, Valid: true}
	}

	result, err := r.queries.UpdateCampaign(ctx, sqlc.UpdateCampaignParams{
		Name:        input.Name,
		Description: sql.NullString{},
		Channel:     sqlc.CrmCampaignsChannel(input.Type),
		SegmentID:   segmentID,
		Subject:     sql.NullString{String: input.Subject, Valid: input.Subject != ""},
		Content:     input.Body,
		ScheduledAt: nullableTime(input.ScheduledAt),
		ID:          input.ID,
		OwnerID:     input.OwnerID,
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
		return domain.ErrCampaignNotFound
	}

	return nil
}

func (r *Repository) UpdateStatus(ctx context.Context, ownerID, id int64, status string) error {
	result, err := r.queries.UpdateCampaignStatus(ctx, sqlc.UpdateCampaignStatusParams{
		Status:  sqlc.NullCrmCampaignsStatus{CrmCampaignsStatus: sqlc.CrmCampaignsStatus(status), Valid: true},
		ID:      id,
		OwnerID: ownerID,
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
		return domain.ErrCampaignNotFound
	}

	return nil
}

func campaignFromRow(row sqlc.CrmCampaign) domain.Campaign {
	c := domain.Campaign{
		ID:        row.ID,
		OwnerID:   row.OwnerID,
		Name:      row.Name,
		Type:      string(row.Channel),
		Status:    string(row.Status.CrmCampaignsStatus),
		Subject:   row.Subject.String,
		Body:      row.Content,
		SentCount: int(row.SentCount.Int32),
		OpenCount: int(row.OpenedCount.Int32),
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
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

func (r *Repository) Delete(ctx context.Context, ownerID, id int64) error {
	return r.queries.DeleteCampaign(ctx, sqlc.DeleteCampaignParams{
		ID:      id,
		OwnerID: ownerID,
	})
}
