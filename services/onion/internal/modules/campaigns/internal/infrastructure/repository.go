package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/campaigns/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Campaign, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, owner_id, name, type, status, audience_segment_id, subject, body, sent_count, open_count, scheduled_at, sent_at, created_at, updated_at
		 FROM crm_campaigns WHERE owner_id = ? ORDER BY created_at DESC`, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list campaigns", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	campaigns := make([]domain.Campaign, 0)
	for rows.Next() {
		c, scanErr := scanCampaign(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		campaigns = append(campaigns, c)
	}
	return campaigns, rows.Err()
}

func (r *Repository) GetByID(ctx context.Context, ownerID, id int64) (domain.Campaign, error) {
	row := r.pool.QueryRowContext(ctx,
		`SELECT id, owner_id, name, type, status, audience_segment_id, subject, body, sent_count, open_count, scheduled_at, sent_at, created_at, updated_at
		 FROM crm_campaigns WHERE id = ? AND owner_id = ?`, id, ownerID,
	)
	var c domain.Campaign
	var segID sql.NullInt64
	var subject, body sql.NullString
	var scheduledAt, sentAt sql.NullTime
	err := row.Scan(&c.ID, &c.OwnerID, &c.Name, &c.Type, &c.Status, &segID, &subject, &body, &c.SentCount, &c.OpenCount, &scheduledAt, &sentAt, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Campaign{}, domain.ErrCampaignNotFound
		}
		return domain.Campaign{}, err
	}
	if segID.Valid {
		c.AudienceSegmentID = &segID.Int64
	}
	c.Subject = subject.String
	c.Body = body.String
	if scheduledAt.Valid {
		t := scheduledAt.Time
		c.ScheduledAt = &t
	}
	if sentAt.Valid {
		t := sentAt.Time
		c.SentAt = &t
	}
	return c, nil
}

func (r *Repository) Create(ctx context.Context, input domain.CampaignInput) (int64, error) {
	result, err := r.pool.ExecContext(ctx,
		`INSERT INTO crm_campaigns (owner_id, name, type, status, audience_segment_id, subject, body, scheduled_at)
		 VALUES (?, ?, ?, 'draft', ?, ?, ?, ?)`,
		input.OwnerID, input.Name, input.Type, input.AudienceSegmentID,
		nullableString(input.Subject), nullableString(input.Body), nullableTime(input.ScheduledAt),
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create campaign", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.CampaignUpdateInput) error {
	result, err := r.pool.ExecContext(ctx,
		`UPDATE crm_campaigns SET name = ?, type = ?, audience_segment_id = ?, subject = ?, body = ?, scheduled_at = ?
		 WHERE id = ? AND owner_id = ?`,
		input.Name, input.Type, input.AudienceSegmentID,
		nullableString(input.Subject), nullableString(input.Body), nullableTime(input.ScheduledAt),
		input.ID, input.OwnerID,
	)
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
	result, err := r.pool.ExecContext(ctx,
		"UPDATE crm_campaigns SET status = ? WHERE id = ? AND owner_id = ?",
		status, id, ownerID,
	)
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

func scanCampaign(rows *sql.Rows) (domain.Campaign, error) {
	var c domain.Campaign
	var segID sql.NullInt64
	var subject, body sql.NullString
	var scheduledAt, sentAt sql.NullTime
	if err := rows.Scan(&c.ID, &c.OwnerID, &c.Name, &c.Type, &c.Status, &segID, &subject, &body, &c.SentCount, &c.OpenCount, &scheduledAt, &sentAt, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return domain.Campaign{}, err
	}
	if segID.Valid {
		c.AudienceSegmentID = &segID.Int64
	}
	c.Subject = subject.String
	c.Body = body.String
	if scheduledAt.Valid {
		t := scheduledAt.Time
		c.ScheduledAt = &t
	}
	if sentAt.Valid {
		t := sentAt.Time
		c.SentAt = &t
	}
	return c, nil
}

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullableTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
