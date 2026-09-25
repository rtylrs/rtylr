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

type ScheduledReportWrite struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	TemplateID string
	Cadence    string
	Channel    string
	Recipient  string
	IsActive   *bool
}

var ErrScheduledReportNotFound = errors.New("scheduled report not found")

func nullableBool(value *bool) sql.NullBool {
	if value == nil {
		return sql.NullBool{}
	}
	return sql.NullBool{Bool: *value, Valid: true}
}

func scheduledFromModel(row sqlc.InsightsScheduledReport) ScheduledReportRow {
	return ScheduledReportRow{
		ID: row.ID, OwnerID: row.OwnerID, BusinessID: row.BusinessID,
		TemplateID: row.TemplateID, Cadence: row.Cadence, Channel: row.Channel,
		Recipient: row.Recipient, IsActive: row.IsActive, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func (r *Repository) ListScheduledReports(ctx context.Context, businessID int64, sortBy, sortDir string, limit, offset int32) ([]ScheduledReportRow, int64, error) {
	total, err := r.queries.CountScheduledReports(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: count scheduled reports failed", logger.Error(err))
		return nil, 0, err
	}
	rows, err := r.queries.ListScheduledReports(ctx, sqlc.ListScheduledReportsParams{
		BusinessID: businessID,
		SortBy:     sortBy,
		SortDir:    sortDir,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: list scheduled reports failed", logger.Error(err))
		return nil, 0, err
	}
	out := make([]ScheduledReportRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, scheduledFromModel(row))
	}
	return out, total, nil
}

func (r *Repository) CreateScheduledReport(ctx context.Context, row ScheduledReportWrite) (int64, error) {
	result, err := r.queries.CreateScheduledReport(ctx, sqlc.CreateScheduledReportParams{
		OwnerID: row.OwnerID, BusinessID: row.BusinessID, TemplateID: row.TemplateID,
		Cadence: row.Cadence, Channel: row.Channel, Recipient: row.Recipient, IsActive: nullableBool(row.IsActive),
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: create scheduled report failed", logger.Error(err))
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repository) UpdateScheduledReport(ctx context.Context, row ScheduledReportWrite) error {
	if _, err := r.queries.GetScheduledReport(ctx, sqlc.GetScheduledReportParams{ID: row.ID, BusinessID: row.BusinessID}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrScheduledReportNotFound
		}
		return err
	}
	err := r.queries.UpdateScheduledReport(ctx, sqlc.UpdateScheduledReportParams{
		TemplateID: row.TemplateID, Cadence: row.Cadence, Channel: row.Channel,
		Recipient: row.Recipient, IsActive: nullableBool(row.IsActive), ID: row.ID, BusinessID: row.BusinessID,
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
	row, err := r.queries.GetScheduledReport(ctx, sqlc.GetScheduledReportParams{ID: id, BusinessID: businessID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ScheduledReportRow{}, ErrScheduledReportNotFound
		}
		logger.Ctx(ctx).Error("insights: get scheduled report failed", logger.Error(err))
		return ScheduledReportRow{}, err
	}
	return scheduledFromModel(row), nil
}
