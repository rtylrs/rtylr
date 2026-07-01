package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type ScheduledReportRow struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	TemplateID string
	Cadence    string
	Channel    string
	Recipient  string
	IsActive   bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

var ErrScheduledReportNotFound = errors.New("scheduled report not found")

func scheduledFromModel(sr sqlc.InsightsScheduledReport) ScheduledReportRow {
	return ScheduledReportRow{
		ID: sr.ID, OwnerID: sr.OwnerID, BusinessID: sr.BusinessID,
		TemplateID: sr.TemplateID, Cadence: sr.Cadence, Channel: sr.Channel,
		Recipient: sr.Recipient, IsActive: sr.IsActive, CreatedAt: sr.CreatedAt, UpdatedAt: sr.UpdatedAt,
	}
}

func (r *Repository) ListScheduledReports(ctx context.Context, businessID int64) ([]ScheduledReportRow, error) {
	rows, err := r.queries.ListScheduledReports(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: list scheduled reports failed", logger.Error(err))
		return nil, err
	}
	out := make([]ScheduledReportRow, 0, len(rows))
	for _, sr := range rows {
		out = append(out, scheduledFromModel(sr))
	}
	return out, nil
}

func (r *Repository) CreateScheduledReport(ctx context.Context, row ScheduledReportRow) (int64, error) {
	res, err := r.queries.CreateScheduledReport(ctx, sqlc.CreateScheduledReportParams{
		OwnerID: row.OwnerID, BusinessID: row.BusinessID, TemplateID: row.TemplateID,
		Cadence: row.Cadence, Channel: row.Channel, Recipient: row.Recipient, IsActive: row.IsActive,
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: create scheduled report failed", logger.Error(err))
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repository) UpdateScheduledReport(ctx context.Context, row ScheduledReportRow) error {
	if _, err := r.queries.GetScheduledReport(ctx, sqlc.GetScheduledReportParams{ID: row.ID, BusinessID: row.BusinessID}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrScheduledReportNotFound
		}
		return err
	}
	err := r.queries.UpdateScheduledReport(ctx, sqlc.UpdateScheduledReportParams{
		TemplateID: row.TemplateID, Cadence: row.Cadence, Channel: row.Channel,
		Recipient: row.Recipient, IsActive: row.IsActive, ID: row.ID, BusinessID: row.BusinessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: update scheduled report failed", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) DeleteScheduledReport(ctx context.Context, businessID, id int64) error {
	if _, err := r.queries.GetScheduledReport(ctx, sqlc.GetScheduledReportParams{ID: id, BusinessID: businessID}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrScheduledReportNotFound
		}
		return err
	}
	if err := r.queries.DeleteScheduledReport(ctx, sqlc.DeleteScheduledReportParams{ID: id, BusinessID: businessID}); err != nil {
		logger.Ctx(ctx).Error("insights: delete scheduled report failed", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) GetScheduledReport(ctx context.Context, businessID, id int64) (ScheduledReportRow, error) {
	sr, err := r.queries.GetScheduledReport(ctx, sqlc.GetScheduledReportParams{ID: id, BusinessID: businessID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ScheduledReportRow{}, ErrScheduledReportNotFound
		}
		logger.Ctx(ctx).Error("insights: get scheduled report failed", logger.Error(err))
		return ScheduledReportRow{}, err
	}
	return scheduledFromModel(sr), nil
}
